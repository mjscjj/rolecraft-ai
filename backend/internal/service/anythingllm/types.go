package anythingllm

import "time"

// Client represents the AnythingLLM API client
type Client struct {
	BaseURL string
	APIKey  string
}

// Workspace represents an AnythingLLM workspace
type Workspace struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Users         []int64   `json:"users"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	OpenAiApi     string    `json:"openAiApi"`
	OpenAiModel   string    `json:"openAiModel"`
	ChatProvider  string    `json:"chatProvider"`
	ChatModel     string    `json:"chatModel"`
	AgentProvider string    `json:"agentProvider"`
	AgentModel    string    `json:"agentModel"`
	Status        string    `json:"status"`
}

// CreateWorkspaceRequest represents the request to create a workspace
type CreateWorkspaceRequest struct {
	Name         string `json:"name"`
	Slug         string `json:"slug,omitempty"`
	Users        []int  `json:"users,omitempty"`
	SystemPrompt string `json:"systemPrompt,omitempty"`
}

// CreateWorkspaceResponse represents the response from creating a workspace
type CreateWorkspaceResponse struct {
	Workspace *Workspace `json:"workspace,omitempty"`
	Message   string     `json:"message,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// GetWorkspaceResponse represents the response from getting a workspace
type GetWorkspaceResponse struct {
	Workspace *Workspace `json:"workspace,omitempty"`
	Message   string     `json:"message,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Message string `json:"message"`
	Mode    string `json:"mode,omitempty"` // "chat" or "query"
	Stream  bool   `json:"stream,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"` // "textResponse" or "textResponseChunk"
	Response    string        `json:"response"`
	SourceDoc   string        `json:"sourceDoc,omitempty"`
	Attachments []interface{} `json:"attachments,omitempty"`
	Action      string        `json:"action,omitempty"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Response string `json:"response"`
}

// UploadDocumentResponse represents the response from uploading a document
type UploadDocumentResponse struct {
	FileName string `json:"fileName"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Document represents a document in the workspace
type Document struct {
	ID          int64     `json:"id"`
	DocName     string    `json:"docName"`
	DocType     string    `json:"docType"`
	WorkspaceID int64     `json:"workspaceId"`
	S3Key       string    `json:"s3Key"`
	Metadatas   string    `json:"metadatas"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Hash        string    `json:"hash"`
}

// GetDocumentsResponse represents the response from getting documents
type GetDocumentsResponse struct {
	Documents []Document `json:"documents"`
	Message   string     `json:"message,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// DeleteDocumentResponse represents the response from deleting a document
type DeleteDocumentResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// VectorSearchRequest represents a vector search request
type VectorSearchRequest struct {
	Query string `json:"query"`
	TopN  int    `json:"topN"`
}

// VectorSearchResult represents a single search result
type VectorSearchResult struct {
	ID       int64   `json:"id"`
	DocName  string  `json:"docName"`
	Content  string  `json:"content"`
	Score    float64 `json:"score"`
	Metadata string  `json:"metadata"`
}

// VectorSearchResponse represents the response from vector search
type VectorSearchResponse struct {
	Results []VectorSearchResult `json:"results"`
	Message string               `json:"message,omitempty"`
	Error   string               `json:"error,omitempty"`
}

// ChatHistoryItem represents a single chat history item
type ChatHistoryItem struct {
	ID         int64     `json:"id"`
	Workspace  int64     `json:"workspaceId"`
	User       int64     `json:"user_id"`
	Prompt     string    `json:"prompt"`
	Response   string    `json:"response"`
	Attachment string    `json:"attachment,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// GetChatHistoryResponse represents the response from getting chat history
type GetChatHistoryResponse struct {
	History []ChatHistoryItem `json:"history"`
	Message string            `json:"message,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// DeleteChatHistoryResponse represents the response from deleting chat history
type DeleteChatHistoryResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// UpdateWorkspaceRequest represents the request to update a workspace
type UpdateWorkspaceRequest struct {
	Name         string `json:"name,omitempty"`
	SystemPrompt string `json:"systemPrompt,omitempty"`
}

// UpdateWorkspaceResponse represents the response from updating a workspace
type UpdateWorkspaceResponse struct {
	Workspace *Workspace `json:"workspace,omitempty"`
	Message   string     `json:"message,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// APIError represents an API error response
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ClientConfig holds configuration for the AnythingLLM client
type ClientConfig struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

// DefaultClientConfig returns a default client configuration
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}
