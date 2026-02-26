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
	baseURL    = "http://localhost:8080/api/v1"
	authToken  = ""
	documentID = ""
)

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
	docID, err := uploadDocument("./test.pdf")
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
		status, err := getDocumentStatus(documentID)
		if err != nil {
			t.Logf("Failed to get status: %v", err)
			break
		}
		t.Logf("Document status: %s", status)
		
		if status == "completed" {
			t.Log("Document processing completed!")
			break
		} else if status == "failed" {
			t.Fatal("Document processing failed")
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
	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	
	jsonData, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed: %s", string(body))
	}
	
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

// uploadDocument 上传文档
func uploadDocument(filePath string) (string, error) {
	// 创建测试文件
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 创建测试 PDF 文件
		if err := createTestPDF(filePath); err != nil {
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
	
	part, err := writer.CreateFormFile("file", "test.pdf")
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

// createTestPDF 创建测试 PDF 文件
func createTestPDF(path string) error {
	// 简单的 PDF 文件头 (最小有效 PDF)
	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>\nendobj\nxref\n0 4\n0000000000 65535 f\n0000000009 00000 n\n0000000058 00000 n\n0000000115 00000 n\ntrailer\n<< /Size 4 /Root 1 0 R >>\nstartxref\n193\n%%EOF")
	
	return os.WriteFile(path, pdfContent, 0644)
}

// getDocumentStatus 获取文档状态
func getDocumentStatus(docID string) (string, error) {
	req, err := http.NewRequest("GET", baseURL+"/documents/"+docID, nil)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", "Bearer "+authToken)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get status failed: %s", string(respBody))
	}
	
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}
	
	status, ok := data["status"].(string)
	if !ok {
		return "", fmt.Errorf("no status in response")
	}
	
	return status, nil
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
	
	// 测试上传端点
	req, _ := http.NewRequest("GET", "http://150.109.21.115:3001/api/v1/health", nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to connect to AnythingLLM: %v", err)
	}
	defer resp.Body.Close()
	
	t.Logf("AnythingLLM status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		t.Logf("Note: AnythingLLM may not have health endpoint, status: %d", resp.StatusCode)
	}
}
