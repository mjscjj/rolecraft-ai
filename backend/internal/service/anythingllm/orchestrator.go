package anythingllm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

var errWorkspaceNotFound = errors.New("workspace not found")

type OrchestratorConfig struct {
	DefaultProvider string
	DefaultModel    string
	OpenRouterKey   string
	TavilyKey       string
}

type Orchestrator struct {
	baseURL         string
	apiKey          string
	defaultProvider string
	defaultModel    string
	openRouterKey   string
	tavilyKey       string
}

type ChatPayload struct {
	WorkspaceSlug   string
	Message         string
	Mode            string
	Model           string
	Provider        string
	SessionID       string
	EnsureWorkspace bool
	WorkspaceName   string
}

type ChatResult struct {
	Content  string
	Sources  []map[string]interface{}
	Thoughts []string
	Type     string
	Raw      map[string]interface{}
}

type CleanupResult struct {
	OK          bool
	Method      string
	Path        string
	StatusCode  int
	Workspace   string
	Diagnostics []string
}

func NewOrchestrator(baseURL, apiKey string, cfg OrchestratorConfig) *Orchestrator {
	provider := strings.TrimSpace(cfg.DefaultProvider)
	if provider == "" {
		provider = "openrouter"
	}
	return &Orchestrator{
		baseURL:         normalizeBaseURL(baseURL),
		apiKey:          strings.TrimSpace(apiKey),
		defaultProvider: provider,
		defaultModel:    NormalizeWorkspaceModel(cfg.DefaultModel),
		openRouterKey:   strings.TrimSpace(cfg.OpenRouterKey),
		tavilyKey:       strings.TrimSpace(cfg.TavilyKey),
	}
}

func (o *Orchestrator) Enabled() bool {
	return strings.TrimSpace(o.baseURL) != "" && strings.TrimSpace(o.apiKey) != ""
}

func UserWorkspaceSlug(userID string) string {
	raw := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(userID), "-", ""))
	if raw == "" {
		return "user_default"
	}
	slug := "user_" + raw
	if len(slug) > 20 {
		return slug[:20]
	}
	return slug
}

func NormalizeWorkspaceSlug(slug string) (string, error) {
	raw := strings.TrimSpace(strings.ToLower(slug))
	if raw == "" {
		return "", fmt.Errorf("workspace slug is required")
	}
	var b strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	normalized := strings.TrimSpace(b.String())
	if normalized == "" {
		return "", fmt.Errorf("workspace slug is invalid")
	}
	if len(normalized) > 80 {
		normalized = normalized[:80]
	}
	return normalized, nil
}

func NormalizeWorkspaceModel(model string) string {
	raw := strings.TrimSpace(model)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(raw), "openrouter/") {
		return strings.TrimSpace(strings.SplitN(raw, "/", 2)[1])
	}
	return raw
}

func NormalizeMode(mode string) string {
	raw := strings.ToLower(strings.TrimSpace(mode))
	switch raw {
	case "", "chat", "对话":
		return "chat"
	case "agent", "deep", "深度思考":
		return "agent"
	case "query", "ask":
		return "query"
	default:
		return "chat"
	}
}

func EnsureAgentPrefix(message string) string {
	content := strings.TrimSpace(message)
	if strings.HasPrefix(strings.ToLower(content), "@agent") {
		return content
	}
	return strings.TrimSpace("@agent " + content)
}

func StripAgentPrefix(message string) string {
	content := strings.TrimSpace(message)
	if strings.HasPrefix(strings.ToLower(content), "@agent") {
		return strings.TrimSpace(content[6:])
	}
	return content
}

func (o *Orchestrator) ConfigureSystemKeys(ctx context.Context) error {
	if !o.Enabled() {
		return fmt.Errorf("anythingllm is not configured")
	}

	payload := map[string]interface{}{}
	if o.openRouterKey != "" {
		payload["OpenRouterApiKey"] = o.openRouterKey
		payload["GenericOpenAiEmbeddingApiKey"] = o.openRouterKey
	}
	if o.tavilyKey != "" {
		payload["AgentTavilyApiKey"] = o.tavilyKey
	}

	if len(payload) == 0 {
		return nil
	}

	status, body, err := o.doJSON(ctx, http.MethodPost, "/system/update-env", payload, 30*time.Second)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("configure system keys failed: status=%d body=%s", status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (o *Orchestrator) EnsureWorkspaceBySlug(ctx context.Context, slug, name, systemPrompt string) (*Workspace, error) {
	if !o.Enabled() {
		return nil, fmt.Errorf("anythingllm is not configured")
	}
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(name) == "" {
		name = normalizedSlug
	}

	ws, err := o.GetWorkspaceBySlug(ctx, normalizedSlug)
	if err == nil {
		return ws, nil
	}
	if !errors.Is(err, errWorkspaceNotFound) {
		return nil, err
	}

	payload := map[string]interface{}{
		"name":         name,
		"slug":         normalizedSlug,
		"systemPrompt": systemPrompt,
	}
	if o.defaultModel != "" {
		payload["chatProvider"] = o.defaultProvider
		payload["chatModel"] = o.defaultModel
		payload["agentProvider"] = o.defaultProvider
		payload["agentModel"] = o.defaultModel
	}

	status, body, err := o.doJSON(ctx, http.MethodPost, "/workspace/new", payload, 30*time.Second)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("create workspace failed: status=%d body=%s", status, strings.TrimSpace(string(body)))
	}

	created, parseErr := parseWorkspace(body)
	if parseErr == nil && created != nil {
		return created, nil
	}

	return o.GetWorkspaceBySlug(ctx, normalizedSlug)
}

func (o *Orchestrator) GetWorkspaceBySlug(ctx context.Context, slug string) (*Workspace, error) {
	if !o.Enabled() {
		return nil, fmt.Errorf("anythingllm is not configured")
	}
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return nil, err
	}
	status, body, err := o.doJSON(ctx, http.MethodGet, fmt.Sprintf("/workspace/%s", normalizedSlug), nil, 30*time.Second)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, errWorkspaceNotFound
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get workspace failed: status=%d body=%s", status, strings.TrimSpace(string(body)))
	}
	ws, err := parseWorkspace(body)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, errWorkspaceNotFound
	}
	return ws, nil
}

func (o *Orchestrator) UpdateWorkspaceSystemPrompt(ctx context.Context, slug, systemPrompt string) error {
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return err
	}
	status, body, err := o.doJSON(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/update", normalizedSlug), map[string]interface{}{
		"systemPrompt": systemPrompt,
	}, 30*time.Second)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("update workspace prompt failed: status=%d body=%s", status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (o *Orchestrator) SyncWorkspaceRuntimeModel(ctx context.Context, slug, model, provider string) error {
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return err
	}
	normalizedModel := NormalizeWorkspaceModel(model)
	if normalizedModel == "" {
		normalizedModel = o.defaultModel
	}
	if normalizedModel == "" {
		return nil
	}
	normalizedProvider := strings.TrimSpace(provider)
	if normalizedProvider == "" {
		normalizedProvider = o.defaultProvider
	}

	if _, err := o.EnsureWorkspaceBySlug(ctx, normalizedSlug, normalizedSlug, ""); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"chatProvider":  normalizedProvider,
		"chatModel":     normalizedModel,
		"agentProvider": normalizedProvider,
		"agentModel":    normalizedModel,
	}

	apply := func() error {
		status, body, err := o.doJSON(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/update", normalizedSlug), payload, 30*time.Second)
		if err != nil {
			return err
		}
		if status != http.StatusOK {
			return fmt.Errorf("workspace model update failed: status=%d body=%s", status, strings.TrimSpace(string(body)))
		}
		return nil
	}

	if err := apply(); err != nil {
		return err
	}

	if ok, err := o.verifyWorkspaceModel(ctx, normalizedSlug, normalizedProvider, normalizedModel); err == nil && ok {
		return nil
	}

	// Force one more time, then verify again.
	if err := apply(); err != nil {
		return err
	}
	ok, err := o.verifyWorkspaceModel(ctx, normalizedSlug, normalizedProvider, normalizedModel)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("workspace model sync verification failed: slug=%s provider=%s model=%s", normalizedSlug, normalizedProvider, normalizedModel)
	}
	return nil
}

func (o *Orchestrator) Chat(ctx context.Context, req ChatPayload) (*ChatResult, error) {
	if !o.Enabled() {
		return nil, fmt.Errorf("anythingllm is not configured")
	}
	slug, err := NormalizeWorkspaceSlug(req.WorkspaceSlug)
	if err != nil {
		return nil, err
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}

	mode := NormalizeMode(req.Mode)
	finalMode := mode
	finalMessage := message
	if mode == "agent" {
		finalMode = "chat"
		finalMessage = EnsureAgentPrefix(message)
	}
	if finalMode != "chat" && finalMode != "query" {
		finalMode = "chat"
	}

	model := NormalizeWorkspaceModel(req.Model)
	if model == "" {
		model = o.defaultModel
	}
	provider := strings.TrimSpace(req.Provider)
	if provider == "" {
		provider = o.defaultProvider
	}

	if req.EnsureWorkspace {
		if _, err := o.EnsureWorkspaceBySlug(ctx, slug, req.WorkspaceName, ""); err != nil {
			return nil, err
		}
	}
	if model != "" {
		if err := o.SyncWorkspaceRuntimeModel(ctx, slug, model, provider); err != nil {
			return nil, err
		}
	}

	payload := map[string]interface{}{
		"message": finalMessage,
		"mode":    finalMode,
	}
	if model != "" {
		payload["model"] = model
	}
	if strings.TrimSpace(req.SessionID) != "" {
		payload["sessionId"] = strings.TrimSpace(req.SessionID)
	}

	status, body, err := o.doJSON(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/chat", slug), payload, 120*time.Second)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		// Agent key missing fallback to normal chat.
		if mode == "agent" && isAgentKeyError(body) {
			fallbackPayload := map[string]interface{}{
				"message": StripAgentPrefix(finalMessage),
				"mode":    "chat",
			}
			if model != "" {
				fallbackPayload["model"] = model
			}
			if strings.TrimSpace(req.SessionID) != "" {
				fallbackPayload["sessionId"] = strings.TrimSpace(req.SessionID)
			}
			retryStatus, retryBody, retryErr := o.doJSON(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/chat", slug), fallbackPayload, 120*time.Second)
			if retryErr != nil {
				return nil, retryErr
			}
			if retryStatus != http.StatusOK {
				return nil, fmt.Errorf("anythingllm chat fallback failed: status=%d body=%s", retryStatus, strings.TrimSpace(string(retryBody)))
			}
			return parseChatResult(retryBody)
		}
		return nil, fmt.Errorf("anythingllm chat failed: status=%d body=%s", status, strings.TrimSpace(string(body)))
	}
	return parseChatResult(body)
}

func (o *Orchestrator) GetChatHistory(ctx context.Context, slug string, limit int) ([]map[string]interface{}, error) {
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/workspace/%s/chats", normalizedSlug)
	if limit > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, limit)
	}
	status, body, err := o.doJSON(ctx, http.MethodGet, path, nil, 60*time.Second)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get chat history failed: status=%d body=%s", status, strings.TrimSpace(string(body)))
	}

	var asArray []map[string]interface{}
	if err := json.Unmarshal(body, &asArray); err == nil {
		return asArray, nil
	}

	var asObj map[string]interface{}
	if err := json.Unmarshal(body, &asObj); err != nil {
		return nil, fmt.Errorf("decode chat history failed: %w", err)
	}
	for _, key := range []string{"history", "chats", "data"} {
		if raw, ok := asObj[key].([]interface{}); ok {
			items := make([]map[string]interface{}, 0, len(raw))
			for _, item := range raw {
				if m, ok := item.(map[string]interface{}); ok {
					items = append(items, m)
				}
			}
			return items, nil
		}
	}
	return []map[string]interface{}{asObj}, nil
}

func (o *Orchestrator) DeleteChatHistory(ctx context.Context, slug string) error {
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return err
	}
	status, body, err := o.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/workspace/%s/chats", normalizedSlug), nil, 30*time.Second)
	if err != nil {
		return err
	}
	if status == http.StatusOK || status == http.StatusNoContent || status == http.StatusNotFound {
		return nil
	}
	return fmt.Errorf("delete chat history failed: status=%d body=%s", status, strings.TrimSpace(string(body)))
}

func (o *Orchestrator) UploadDocumentToWorkspace(ctx context.Context, slug, fileName string, fileData []byte) (string, error) {
	if !o.Enabled() {
		return "", fmt.Errorf("anythingllm is not configured")
	}
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return "", err
	}
	resolvedSlug := normalizedSlug
	workspace, err := o.EnsureWorkspaceBySlug(ctx, normalizedSlug, normalizedSlug, "")
	if err != nil {
		return "", err
	}
	if workspace != nil && strings.TrimSpace(workspace.Slug) != "" {
		resolvedSlug = strings.TrimSpace(workspace.Slug)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(fileData); err != nil {
		return "", err
	}
	if err := writer.WriteField("addToWorkspaces", resolvedSlug); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/document/upload", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upload document failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return "", err
	}
	for _, key := range []string{"filename", "fileName", "name"} {
		if v, ok := data[key].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v), nil
		}
	}
	return fileName, nil
}

func (o *Orchestrator) UpdateEmbeddings(ctx context.Context, slug string, adds []string, deletes []string) error {
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return err
	}
	var payload interface{}
	payloadMap := map[string]interface{}{}
	if len(adds) > 0 {
		payloadMap["adds"] = adds
	}
	if len(deletes) > 0 {
		payloadMap["deletes"] = deletes
	}
	if len(payloadMap) > 0 {
		payload = payloadMap
	}
	status, body, err := o.doJSON(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/update-embeddings", normalizedSlug), payload, 10*time.Minute)
	if err != nil {
		return err
	}
	if status == http.StatusBadRequest && payload == nil {
		// Compatibility retry for versions that require an explicit JSON object body.
		status, body, err = o.doJSON(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/update-embeddings", normalizedSlug), map[string]interface{}{}, 10*time.Minute)
		if err != nil {
			return err
		}
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return fmt.Errorf("update embeddings failed: slug=%s status=%d body=%s", normalizedSlug, status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (o *Orchestrator) RemoveDocument(ctx context.Context, slug, filename string) error {
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return err
	}
	status, body, err := o.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/workspace/%s/remove-document", normalizedSlug), map[string]interface{}{
		"filename": filename,
	}, 30*time.Second)
	if err != nil {
		return err
	}
	if status == http.StatusOK || status == http.StatusNoContent || status == http.StatusNotFound {
		return nil
	}
	return fmt.Errorf("remove document failed: status=%d body=%s", status, strings.TrimSpace(string(body)))
}

func (o *Orchestrator) VectorSearch(ctx context.Context, slug, query string, topN int) ([]interface{}, error) {
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return nil, err
	}
	if topN <= 0 {
		topN = 4
	}
	status, body, err := o.doJSON(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/vector-search", normalizedSlug), map[string]interface{}{
		"query": strings.TrimSpace(query),
		"topN":  topN,
	}, 30*time.Second)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("vector search failed: status=%d body=%s", status, strings.TrimSpace(string(body)))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	for _, key := range []string{"responses", "items", "results"} {
		if arr, ok := result[key].([]interface{}); ok {
			return arr, nil
		}
	}
	return []interface{}{result}, nil
}

func (o *Orchestrator) CleanupWorkspace(ctx context.Context, slug string) (*CleanupResult, error) {
	normalizedSlug, err := NormalizeWorkspaceSlug(slug)
	if err != nil {
		return nil, err
	}
	attempts := []struct {
		method  string
		path    string
		payload interface{}
	}{
		{method: http.MethodDelete, path: fmt.Sprintf("/workspace/%s", normalizedSlug), payload: nil},
		{method: http.MethodPost, path: fmt.Sprintf("/workspace/%s/delete", normalizedSlug), payload: map[string]interface{}{"slug": normalizedSlug}},
		{method: http.MethodPost, path: fmt.Sprintf("/workspace/%s/archive", normalizedSlug), payload: map[string]interface{}{"slug": normalizedSlug, "archived": true}},
	}
	diagnostics := make([]string, 0, len(attempts))
	for _, attempt := range attempts {
		status, body, reqErr := o.doJSON(ctx, attempt.method, attempt.path, attempt.payload, 30*time.Second)
		if reqErr != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("%s %s error=%v", attempt.method, attempt.path, reqErr))
			continue
		}
		if status == http.StatusOK || status == http.StatusCreated || status == http.StatusNoContent || status == http.StatusNotFound {
			return &CleanupResult{
				OK:         true,
				Method:     attempt.method,
				Path:       attempt.path,
				StatusCode: status,
				Workspace:  normalizedSlug,
			}, nil
		}
		diagnostics = append(diagnostics, fmt.Sprintf("%s %s status=%d body=%s", attempt.method, attempt.path, status, strings.TrimSpace(string(body))))
	}
	return &CleanupResult{
		OK:          false,
		Workspace:   normalizedSlug,
		Diagnostics: diagnostics,
	}, fmt.Errorf("workspace cleanup failed: %s", strings.Join(diagnostics, " | "))
}

func (o *Orchestrator) verifyWorkspaceModel(ctx context.Context, slug, provider, model string) (bool, error) {
	ws, err := o.GetWorkspaceBySlug(ctx, slug)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(ws.ChatProvider) == strings.TrimSpace(provider) &&
		strings.TrimSpace(ws.ChatModel) == strings.TrimSpace(model) &&
		strings.TrimSpace(ws.AgentProvider) == strings.TrimSpace(provider) &&
		strings.TrimSpace(ws.AgentModel) == strings.TrimSpace(model), nil
}

func (o *Orchestrator) doJSON(ctx context.Context, method, path string, payload interface{}, timeout time.Duration) (int, []byte, error) {
	if !o.Enabled() {
		return 0, nil, fmt.Errorf("anythingllm is not configured")
	}
	url := fmt.Sprintf("%s%s", o.baseURL, path)

	var bodyBytes []byte
	if payload != nil {
		var err error
		bodyBytes, err = json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		var body io.Reader
		if bodyBytes != nil {
			body = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return 0, nil, err
		}
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		client := &http.Client{Timeout: timeout}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < 2 {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			break
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if attempt < 2 {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			break
		}

		if resp.StatusCode >= 500 && attempt < 2 {
			lastErr = fmt.Errorf("server error: status=%d", resp.StatusCode)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		return resp.StatusCode, respBody, nil
	}

	if lastErr != nil {
		return 0, nil, lastErr
	}
	return 0, nil, fmt.Errorf("request failed")
}

func parseWorkspace(payload []byte) (*Workspace, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("decode workspace response failed: %w", err)
	}
	if apiErr, ok := data["error"].(string); ok && strings.TrimSpace(apiErr) != "" {
		return nil, fmt.Errorf("api error: %s", strings.TrimSpace(apiErr))
	}

	rawWorkspace, exists := data["workspace"]
	if !exists {
		return nil, nil
	}

	var wsMap map[string]interface{}
	switch typed := rawWorkspace.(type) {
	case map[string]interface{}:
		wsMap = typed
	case []interface{}:
		if len(typed) > 0 {
			if m, ok := typed[0].(map[string]interface{}); ok {
				wsMap = m
			}
		}
	}
	if wsMap == nil {
		return nil, nil
	}
	ws := &Workspace{
		ID:            toInt64Value(wsMap["id"]),
		Name:          toStringValue(wsMap["name"]),
		Slug:          toStringValue(wsMap["slug"]),
		Status:        toStringValue(wsMap["status"]),
		ChatProvider:  toStringValue(wsMap["chatProvider"]),
		ChatModel:     toStringValue(wsMap["chatModel"]),
		AgentProvider: toStringValue(wsMap["agentProvider"]),
		AgentModel:    toStringValue(wsMap["agentModel"]),
		OpenAiApi:     toStringValue(wsMap["openAiApi"]),
		OpenAiModel:   toStringValue(wsMap["openAiModel"]),
	}
	if ws.Slug == "" {
		return nil, nil
	}
	return ws, nil
}

func parseChatResult(payload []byte) (*ChatResult, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("decode chat response failed: %w", err)
	}
	if apiErr, ok := data["error"].(string); ok && strings.TrimSpace(apiErr) != "" {
		return nil, fmt.Errorf(strings.TrimSpace(apiErr))
	}

	content := ""
	if tr, ok := data["textResponse"]; ok {
		switch typed := tr.(type) {
		case string:
			content = strings.TrimSpace(typed)
		case map[string]interface{}:
			content = strings.TrimSpace(getString(typed, "content"))
		}
	}
	if content == "" {
		content = strings.TrimSpace(getString(data, "response"))
	}

	thoughts := toStringSlice(data["thoughts"])
	sources := toObjectSlice(data["sources"])
	return &ChatResult{
		Content:  content,
		Thoughts: thoughts,
		Sources:  sources,
		Type:     getString(data, "type"),
		Raw:      data,
	}, nil
}

func isAgentKeyError(payload []byte) bool {
	text := strings.ToLower(strings.TrimSpace(string(payload)))
	return strings.Contains(text, "openai api key must be provided to use agents")
}

func getString(data map[string]interface{}, key string) string {
	v, ok := data[key]
	if !ok {
		return ""
	}
	switch typed := v.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func toStringSlice(raw interface{}) []string {
	values, ok := raw.([]interface{})
	if !ok {
		return []string{}
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}

func toObjectSlice(raw interface{}) []map[string]interface{} {
	values, ok := raw.([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, 0, len(values))
	for _, value := range values {
		if m, ok := value.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

func toStringValue(v interface{}) string {
	switch typed := v.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

func toInt64Value(v interface{}) int64 {
	switch typed := v.(type) {
	case int:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	default:
		return 0
	}
}
