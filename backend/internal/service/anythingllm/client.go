package anythingllm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// NewAnythingLLMClient creates a new AnythingLLM client with default configuration
func NewAnythingLLMClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
	}
}

// NewAnythingLLMClientWithConfig creates a new AnythingLLM client with custom configuration
func NewAnythingLLMClientWithConfig(config ClientConfig) *Client {
	return &Client{
		BaseURL: strings.TrimRight(config.BaseURL, "/"),
		APIKey:  config.APIKey,
	}
}

// GetWorkspaceSlug generates the workspace slug for a user
func (c *Client) GetWorkspaceSlug(userId string) string {
	return fmt.Sprintf("user_%s", userId)
}

// getWorkspaceSlug generates the workspace slug for a user (deprecated, use GetWorkspaceSlug)
func (c *Client) getWorkspaceSlug(userId string) string {
	return c.GetWorkspaceSlug(userId)
}

// doRequest performs an HTTP request with retry logic
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		
		// Don't retry on client errors (4xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return resp, nil
		}
		
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		
		return resp, nil
	}
	
	return nil, fmt.Errorf("request failed after retries: %w", lastErr)
}

// CreateWorkspace creates a new workspace for a user
func (c *Client) CreateWorkspace(userId, name, systemPrompt string) (*Workspace, error) {
	slug := c.getWorkspaceSlug(userId)
	
	reqBody := CreateWorkspaceRequest{
		Name:       name,
		Slug:       slug,
		SystemPrompt: systemPrompt,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodPost, "/workspace/new", bytes.NewReader(jsonData), "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result CreateWorkspaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("api error: %s", result.Error)
	}
	
	if result.Workspace == nil {
		return nil, fmt.Errorf("no workspace returned")
	}
	
	return result.Workspace, nil
}

// GetWorkspace retrieves a user's workspace
func (c *Client) GetWorkspace(userId string) (*Workspace, error) {
	slug := c.getWorkspaceSlug(userId)
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/workspace/%s", slug), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result GetWorkspaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("api error: %s", result.Error)
	}
	
	if result.Workspace == nil {
		return nil, fmt.Errorf("workspace not found")
	}
	
	return result.Workspace, nil
}

// Chat sends a chat message and gets a response
func (c *Client) Chat(userId, message, mode string) (string, error) {
	slug := c.getWorkspaceSlug(userId)
	
	reqBody := ChatRequest{
		Message: message,
		Mode:    mode,
		Stream:  false,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/chat", slug), bytes.NewReader(jsonData), "application/json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Response, nil
}

// StreamChat sends a chat message and streams the response via SSE
func (c *Client) StreamChat(userId, message, mode string, callback func(chunk string)) error {
	slug := c.getWorkspaceSlug(userId)
	
	reqBody := ChatRequest{
		Message: message,
		Mode:    mode,
		Stream:  true,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	url := fmt.Sprintf("%s/workspace/%s/chat", c.BaseURL, slug)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{
		Timeout: 120 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
	}
	
	// Read SSE stream
	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk StreamChunk
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode stream: %w", err)
		}
		
		if chunk.Response != "" {
			callback(chunk.Response)
		}
	}
	
	return nil
}

// UploadDocument uploads a document to the user's workspace
func (c *Client) UploadDocument(userId string, fileName string, fileData []byte) (*UploadDocumentResponse, error) {
	slug := c.getWorkspaceSlug(userId)
	
	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("document", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}
	
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/upload-document", slug), body, writer.FormDataContentType())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result UploadDocumentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("api error: %s", result.Error)
	}
	
	return &result, nil
}

// GetDocuments retrieves all documents in the user's workspace
func (c *Client) GetDocuments(userId string) ([]Document, error) {
	slug := c.getWorkspaceSlug(userId)
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/workspace/%s/documents", slug), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result GetDocumentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("api error: %s", result.Error)
	}
	
	return result.Documents, nil
}

// DeleteDocument deletes a document from the user's workspace
func (c *Client) DeleteDocument(userId, docHash string) error {
	slug := c.getWorkspaceSlug(userId)
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/workspace/%s/documents/%s", slug, docHash), nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var result DeleteDocumentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return fmt.Errorf("api error: %s", result.Error)
	}
	
	return nil
}

// VectorSearch performs a vector search in the user's workspace
func (c *Client) VectorSearch(userId, query string, topN int) ([]VectorSearchResult, error) {
	slug := c.getWorkspaceSlug(userId)
	
	reqBody := VectorSearchRequest{
		Query: query,
		TopN:  topN,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/vector-search", slug), bytes.NewReader(jsonData), "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result VectorSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("api error: %s", result.Error)
	}
	
	return result.Results, nil
}

// GetChatHistory retrieves chat history for the user's workspace
func (c *Client) GetChatHistory(userId string, limit int) ([]ChatHistoryItem, error) {
	slug := c.getWorkspaceSlug(userId)
	
	path := fmt.Sprintf("/workspace/%s/chats", slug)
	if limit > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, limit)
	}
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result GetChatHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("api error: %s", result.Error)
	}
	
	return result.History, nil
}

// DeleteChatHistory deletes chat history from the user's workspace
// Note: AnythingLLM doesn't support deleting individual messages, only entire chat history
func (c *Client) DeleteChatHistory(userId string) error {
	slug := c.getWorkspaceSlug(userId)
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/workspace/%s/chats", slug), nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var result DeleteChatHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return fmt.Errorf("api error: %s", result.Error)
	}
	
	return nil
}

// UpdateWorkspaceSystemPrompt updates the system prompt for a workspace
// Uses POST /v1/workspace/:slug/update endpoint
func (c *Client) UpdateWorkspaceSystemPrompt(slug, systemPrompt string) error {
	// First, try to get the workspace
	_, err := c.GetWorkspaceBySlug(slug)
	if err != nil {
		// Workspace doesn't exist, create it
		_, err := c.CreateWorkspaceBySlug(slug, slug, systemPrompt)
		return err
	}
	
	// Workspace exists, update settings via POST /workspace/:slug/update
	reqBody := UpdateWorkspaceRequest{
		SystemPrompt: systemPrompt,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/workspace/%s/update", slug), bytes.NewReader(jsonData), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var result UpdateWorkspaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return fmt.Errorf("api error: %s", result.Error)
	}
	
	return nil
}

// GetWorkspaceBySlug retrieves a workspace by its slug
func (c *Client) GetWorkspaceBySlug(slug string) (*Workspace, error) {
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/workspace/%s", slug), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result GetWorkspaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("api error: %s", result.Error)
	}
	
	if result.Workspace == nil {
		return nil, fmt.Errorf("workspace not found")
	}
	
	return result.Workspace, nil
}

// CreateWorkspaceBySlug creates a new workspace with specific slug
func (c *Client) CreateWorkspaceBySlug(slug, name, systemPrompt string) (*Workspace, error) {
	reqBody := CreateWorkspaceRequest{
		Name:         name,
		Slug:         slug,
		SystemPrompt: systemPrompt,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	ctx := context.Background()
	resp, err := c.doRequest(ctx, http.MethodPost, "/workspace/new", bytes.NewReader(jsonData), "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result CreateWorkspaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("api error: %s", result.Error)
	}
	
	if result.Workspace == nil {
		return nil, fmt.Errorf("no workspace returned")
	}
	
	return result.Workspace, nil
}

// ListWorkspaces lists all workspaces (for health check)
func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/workspaces", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Workspaces []Workspace `json:"workspaces"`
		Message    string      `json:"message,omitempty"`
		Error      string      `json:"error,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("api error: %s", result.Error)
	}
	
	return result.Workspaces, nil
}
