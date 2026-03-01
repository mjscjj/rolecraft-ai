package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	baseURL    = getEnvOrDefault("ROLECRAFT_BASE_URL", "http://localhost:8080/api/v1")
	authToken  = ""
	documentID = ""
)

func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// TestDocumentFlow 测试文档完整流程
func TestDocumentFlow(t *testing.T) {
	// 1. 登录获取 token
	t.Log("=== Step 1: Login ===")
	token, err := login()
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	authToken = token
	t.Logf("Got auth token: %s...", authToken[:20])

	// 2. 上传文档
	t.Log("=== Step 2: Upload Document ===")
	docID, err := uploadDocument("./test.txt")
	if err != nil {
		t.Fatalf("Failed to upload document: %v", err)
	}
	documentID = docID
	t.Logf("Document uploaded, ID: %s", documentID)

	// 3. 等待异步处理完成
	t.Log("=== Step 3: Wait for Processing ===")
	time.Sleep(5 * time.Second) // 等待处理

	// 轮询检查状态
	for i := 0; i < 10; i++ {
		status, detail, err := getDocumentStatus(documentID)
		if err != nil {
			t.Logf("Failed to get status: %v", err)
			break
		}
		t.Logf("Document status: %s (detail: %s)", status, detail)

		if status == "completed" {
			t.Log("Document processing completed!")
			break
		} else if status == "failed" {
			t.Fatalf("Document processing failed: %s", detail)
		}

		time.Sleep(3 * time.Second)
	}

	// 4. 向量搜索测试
	t.Log("=== Step 4: Vector Search ===")
	results, err := vectorSearch("测试关键词", 4)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	t.Logf("Search results: %+v", results)

	// 5. 删除文档
	t.Log("=== Step 5: Delete Document ===")
	err = deleteDocument(documentID)
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}
	t.Log("Document deleted successfully")
}

// login 登录获取 token
func login() (string, error) {
	email := getEnvOrDefault("ROLECRAFT_TEST_EMAIL", "test@example.com")
	password := getEnvOrDefault("ROLECRAFT_TEST_PASSWORD", "password123")
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}

	jsonData, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		data, ok := result["data"].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("invalid response format")
		}

		token, ok := data["token"].(string)
		if !ok {
			return "", fmt.Errorf("no token in response")
		}
		return token, nil
	}

	// 集成测试需要可重复执行：账号不存在时自动注册后重试登录。
	registerReq := map[string]string{
		"email":    email,
		"password": password,
		"name":     "Integration Test User",
	}
	registerBody, _ := json.Marshal(registerReq)
	registerResp, registerErr := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(registerBody))
	if registerErr != nil {
		return "", fmt.Errorf("login failed: %s; register request error: %v", string(body), registerErr)
	}
	defer registerResp.Body.Close()
	if registerResp.StatusCode != http.StatusOK && registerResp.StatusCode != http.StatusCreated {
		registerRespBody, _ := io.ReadAll(registerResp.Body)
		// 账号被占用且密码未知，创建唯一测试账号继续集成测试。
		fallbackEmail := fmt.Sprintf("integration_%d@example.com", time.Now().UnixNano())
		fallbackReq := map[string]string{
			"email":    fallbackEmail,
			"password": password,
			"name":     "Integration Test User",
		}
		fallbackBody, _ := json.Marshal(fallbackReq)
		fallbackResp, fallbackErr := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(fallbackBody))
		if fallbackErr != nil {
			return "", fmt.Errorf("login failed: %s; register failed: %s; fallback register request error: %v", string(body), string(registerRespBody), fallbackErr)
		}
		defer fallbackResp.Body.Close()
		if fallbackResp.StatusCode != http.StatusOK && fallbackResp.StatusCode != http.StatusCreated {
			fallbackRespBody, _ := io.ReadAll(fallbackResp.Body)
			return "", fmt.Errorf("login failed: %s; register failed: %s; fallback register failed: %s", string(body), string(registerRespBody), string(fallbackRespBody))
		}
		jsonData, _ = json.Marshal(map[string]string{
			"email":    fallbackEmail,
			"password": password,
		})
	}

	retryResp, retryErr := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if retryErr != nil {
		return "", fmt.Errorf("retry login request failed: %v", retryErr)
	}
	defer retryResp.Body.Close()
	retryBody, _ := io.ReadAll(retryResp.Body)
	if retryResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("retry login failed: %s", string(retryBody))
	}

	var retryResult map[string]interface{}
	json.Unmarshal(retryBody, &retryResult)
	data, ok := retryResult["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid retry login response format")
	}
	token, ok := data["token"].(string)
	if !ok {
		return "", fmt.Errorf("no token in retry login response")
	}
	return token, nil
}

// uploadDocument 上传文档
func uploadDocument(filePath string) (string, error) {
	// 创建测试文件
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 创建稳定的文本测试文件，避免空 PDF 导致 AnythingLLM 无法解析
		if err := createTestDocument(filePath); err != nil {
			return "", fmt.Errorf("failed to create test file: %w", err)
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/documents", body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("upload failed: %s", string(respBody))
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	id, ok := data["id"].(string)
	if !ok {
		return "", fmt.Errorf("no id in response")
	}

	return id, nil
}

// createTestDocument 创建测试文档
func createTestDocument(path string) error {
	content := []byte("RoleCraft integration test document.\nThis text should be indexed by AnythingLLM.\n")
	return os.WriteFile(path, content, 0644)
}

// getDocumentStatus 获取文档状态
func getDocumentStatus(docID string) (string, string, error) {
	req, err := http.NewRequest("GET", baseURL+"/documents/"+docID, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("get status failed: %s", string(respBody))
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("invalid response format")
	}

	status, ok := data["status"].(string)
	if !ok {
		return "", "", fmt.Errorf("no status in response")
	}

	detail, _ := data["errorMessage"].(string)
	return status, detail, nil
}

// vectorSearch 向量搜索
func vectorSearch(query string, topN int) ([]interface{}, error) {
	reqBody := map[string]interface{}{
		"query": query,
		"topN":  topN,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/documents/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", string(respBody))
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	results, ok := data["results"].([]interface{})
	if !ok {
		return []interface{}{data}, nil
	}

	return results, nil
}

// deleteDocument 删除文档
func deleteDocument(docID string) error {
	req, err := http.NewRequest("DELETE", baseURL+"/documents/"+docID, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s", string(respBody))
	}

	return nil
}

// TestAnythingLLMConnection 测试 AnythingLLM 连接
func TestAnythingLLMConnection(t *testing.T) {
	t.Log("=== Testing AnythingLLM Connection ===")

	base := getEnvOrDefault("ANYTHINGLLM_BASE_URL", getEnvOrDefault("ANYTHINGLLM_URL", "http://43.134.234.4:3001"))
	client := &http.Client{Timeout: 10 * time.Second}
	healthURLs := []string{
		base + "/api/v1/health",
		base + "/health",
	}

	for _, url := range healthURLs {
		req, _ := http.NewRequest("GET", url, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("health check failed on %s: %v", url, err)
			continue
		}
		defer resp.Body.Close()
		t.Logf("AnythingLLM status (%s): %d", url, resp.StatusCode)
		if resp.StatusCode >= 200 && resp.StatusCode < 500 {
			return
		}
	}

	t.Fatalf("AnythingLLM health endpoint not reachable. checked base=%s", base)
}
