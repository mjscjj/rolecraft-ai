package anythingllm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewAnythingLLMClient tests client initialization
func TestNewAnythingLLMClient(t *testing.T) {
	baseURL := "http://localhost:3001/api/v1"
	apiKey := "test-api-key"
	
	client := NewAnythingLLMClient(baseURL, apiKey)
	
	if client.BaseURL != baseURL {
		t.Errorf("expected BaseURL %s, got %s", baseURL, client.BaseURL)
	}
	
	if client.APIKey != apiKey {
		t.Errorf("expected APIKey %s, got %s", apiKey, client.APIKey)
	}
}

// TestGetWorkspaceSlug tests workspace slug generation
func TestGetWorkspaceSlug(t *testing.T) {
	client := NewAnythingLLMClient("http://localhost:3001/api/v1", "test-key")
	
	testCases := []struct {
		userId   string
		expected string
	}{
		{"123", "user_123"},
		{"abc", "user_abc"},
		{"user@example.com", "user_user@example.com"},
	}
	
	for _, tc := range testCases {
		result := client.getWorkspaceSlug(tc.userId)
		if result != tc.expected {
			t.Errorf("expected slug %s for userId %s, got %s", tc.expected, tc.userId, result)
		}
	}
}

// TestCreateWorkspace tests workspace creation
func TestCreateWorkspace(t *testing.T) {
	expectedWorkspace := &Workspace{
		ID:    1,
		Name:  "Test Workspace",
		Slug:  "user_123",
		Status: "active",
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		
		if r.URL.Path != "/api/v1/workspace/new" {
			t.Errorf("expected path /api/v1/workspace/new, got %s", r.URL.Path)
		}
		
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-key" {
			t.Errorf("expected Bearer test-api-key, got %s", auth)
		}
		
		var req CreateWorkspaceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		
		if req.Name != "Test Workspace" {
			t.Errorf("expected name 'Test Workspace', got %s", req.Name)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CreateWorkspaceResponse{
			Workspace: expectedWorkspace,
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	workspace, err := client.CreateWorkspace("123", "Test Workspace", "Test system prompt")
	if err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}
	
	if workspace.ID != expectedWorkspace.ID {
		t.Errorf("expected workspace ID %d, got %d", expectedWorkspace.ID, workspace.ID)
	}
	
	if workspace.Name != expectedWorkspace.Name {
		t.Errorf("expected workspace name %s, got %s", expectedWorkspace.Name, workspace.Name)
	}
}

// TestCreateWorkspaceError tests error handling in workspace creation
func TestCreateWorkspaceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateWorkspaceResponse{
			Error: "Workspace already exists",
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	_, err := client.CreateWorkspace("123", "Test Workspace", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	
	if !strings.Contains(err.Error(), "Workspace already exists") {
		t.Errorf("expected error about existing workspace, got: %v", err)
	}
}

// TestGetWorkspace tests workspace retrieval
func TestGetWorkspace(t *testing.T) {
	expectedWorkspace := &Workspace{
		ID:    1,
		Name:  "Test Workspace",
		Slug:  "user_123",
		Status: "active",
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		
		if r.URL.Path != "/api/v1/workspace/user_123" {
			t.Errorf("expected path /api/v1/workspace/user_123, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GetWorkspaceResponse{
			Workspace: expectedWorkspace,
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	workspace, err := client.GetWorkspace("123")
	if err != nil {
		t.Fatalf("GetWorkspace failed: %v", err)
	}
	
	if workspace.ID != expectedWorkspace.ID {
		t.Errorf("expected workspace ID %d, got %d", expectedWorkspace.ID, workspace.ID)
	}
}

// TestGetWorkspaceNotFound tests 404 handling
func TestGetWorkspaceNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(GetWorkspaceResponse{
			Error: "Workspace not found",
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	_, err := client.GetWorkspace("999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestChat tests chat functionality
func TestChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		
		if req.Message != "Hello" {
			t.Errorf("expected message 'Hello', got %s", req.Message)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{
			Response: "Hi there!",
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	response, err := client.Chat("123", "Hello", "chat")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	
	if response != "Hi there!" {
		t.Errorf("expected response 'Hi there!', got %s", response)
	}
}

// TestStreamChat tests streaming chat functionality
func TestStreamChat(t *testing.T) {
	chunks := []StreamChunk{
		{Response: "Hello"},
		{Response: " "},
		{Response: "world"},
		{Response: "!"},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		for _, chunk := range chunks {
			encoder.Encode(chunk)
			w.(http.Flusher).Flush()
		}
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	var received []string
	callback := func(chunk string) {
		received = append(received, chunk)
	}
	
	err := client.StreamChat("123", "Test message", "chat", callback)
	if err != nil {
		t.Fatalf("StreamChat failed: %v", err)
	}
	
	expected := []string{"Hello", " ", "world", "!"}
	if len(received) != len(expected) {
		t.Errorf("expected %d chunks, got %d", len(expected), len(received))
	}
}

// TestUploadDocument tests document upload
func TestUploadDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("expected multipart/form-data, got %s", r.Header.Get("Content-Type"))
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UploadDocumentResponse{
			FileName: "test.pdf",
			Message:  "Document uploaded successfully",
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	fileData := []byte("test content")
	response, err := client.UploadDocument("123", "test.pdf", fileData)
	if err != nil {
		t.Fatalf("UploadDocument failed: %v", err)
	}
	
	if response.FileName != "test.pdf" {
		t.Errorf("expected filename 'test.pdf', got %s", response.FileName)
	}
}

// TestGetDocuments tests document listing
func TestGetDocuments(t *testing.T) {
	expectedDocs := []Document{
		{ID: 1, DocName: "doc1.pdf", DocType: "pdf"},
		{ID: 2, DocName: "doc2.txt", DocType: "txt"},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GetDocumentsResponse{
			Documents: expectedDocs,
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	docs, err := client.GetDocuments("123")
	if err != nil {
		t.Fatalf("GetDocuments failed: %v", err)
	}
	
	if len(docs) != len(expectedDocs) {
		t.Errorf("expected %d documents, got %d", len(expectedDocs), len(docs))
	}
}

// TestDeleteDocument tests document deletion
func TestDeleteDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DeleteDocumentResponse{
			Message: "Document deleted successfully",
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	err := client.DeleteDocument("123", "doc-hash-123")
	if err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}
}

// TestVectorSearch tests vector search functionality
func TestVectorSearch(t *testing.T) {
	expectedResults := []VectorSearchResult{
		{ID: 1, DocName: "doc1.pdf", Content: "relevant content", Score: 0.95},
		{ID: 2, DocName: "doc2.pdf", Content: "more content", Score: 0.85},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req VectorSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		
		if req.Query != "test query" {
			t.Errorf("expected query 'test query', got %s", req.Query)
		}
		
		if req.TopN != 5 {
			t.Errorf("expected topN 5, got %d", req.TopN)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(VectorSearchResponse{
			Results: expectedResults,
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	results, err := client.VectorSearch("123", "test query", 5)
	if err != nil {
		t.Fatalf("VectorSearch failed: %v", err)
	}
	
	if len(results) != len(expectedResults) {
		t.Errorf("expected %d results, got %d", len(expectedResults), len(results))
	}
}

// TestGetChatHistory tests chat history retrieval
func TestGetChatHistory(t *testing.T) {
	expectedHistory := []ChatHistoryItem{
		{ID: 1, Prompt: "Hello", Response: "Hi"},
		{ID: 2, Prompt: "How are you?", Response: "I'm good"},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GetChatHistoryResponse{
			History: expectedHistory,
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	history, err := client.GetChatHistory("123", 10)
	if err != nil {
		t.Fatalf("GetChatHistory failed: %v", err)
	}
	
	if len(history) != len(expectedHistory) {
		t.Errorf("expected %d history items, got %d", len(expectedHistory), len(history))
	}
}

// TestRetryLogic tests that the client retries on server errors
func TestRetryLogic(t *testing.T) {
	attempts := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GetWorkspaceResponse{
			Workspace: &Workspace{ID: 1, Name: "Test"},
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	_, err := client.GetWorkspace("123")
	if err != nil {
		t.Fatalf("GetWorkspace failed after retries: %v", err)
	}
	
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

// TestContextCancellation tests that context cancellation is respected
func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GetWorkspaceResponse{
			Workspace: &Workspace{ID: 1},
		})
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	// Note: Our current implementation doesn't use the passed context
	// This is a placeholder for future implementation
	_ = ctx
	_ = client
}

// TestInvalidJSON tests handling of invalid JSON responses
func TestInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json{"))
	}))
	defer server.Close()
	
	client := NewAnythingLLMClient(server.URL+"/api/v1", "test-api-key")
	
	_, err := client.GetWorkspace("123")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to decode") {
		t.Errorf("expected decode error, got: %v", err)
	}
}

// TestBaseURLTrailingSlash tests that trailing slashes are handled correctly
func TestBaseURLTrailingSlash(t *testing.T) {
	client1 := NewAnythingLLMClient("http://localhost:3001/api/", "key")
	client2 := NewAnythingLLMClient("http://localhost:3001/api/v1", "key")
	
	if client1.BaseURL != client2.BaseURL {
		t.Errorf("expected same BaseURL, got %s vs %s", client1.BaseURL, client2.BaseURL)
	}
}
