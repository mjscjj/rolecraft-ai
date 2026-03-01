package anythingllm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChatAgentModeBuildsSessionPayload(t *testing.T) {
	var seenPath string
	var seenBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&seenBody); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"response": "ok",
			"type":     "textResponse",
		})
	}))
	defer server.Close()

	o := NewOrchestrator(server.URL, "test-key", OrchestratorConfig{})
	result, err := o.Chat(context.Background(), ChatPayload{
		WorkspaceSlug: "abc_123",
		Message:       "please search web",
		Mode:          "agent",
		SessionID:     "sess-001",
	})
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if result.Content != "ok" {
		t.Fatalf("unexpected content: %s", result.Content)
	}
	if seenPath != "/api/v1/workspace/abc_123/chat" {
		t.Fatalf("unexpected path: %s", seenPath)
	}

	if mode, _ := seenBody["mode"].(string); mode != "chat" {
		t.Fatalf("agent mode should map to chat, got: %v", seenBody["mode"])
	}
	message, _ := seenBody["message"].(string)
	if !strings.HasPrefix(strings.ToLower(message), "@agent ") {
		t.Fatalf("agent message should be prefixed, got: %q", message)
	}
	if sid, _ := seenBody["sessionId"].(string); sid != "sess-001" {
		t.Fatalf("expected sessionId sess-001, got: %v", seenBody["sessionId"])
	}
}

func TestSyncWorkspaceRuntimeModelVerifyFailure(t *testing.T) {
	updateCalls := 0
	getCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workspace/sync_slug":
			getCalls++
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"workspace": map[string]interface{}{
					"id":            1,
					"slug":          "sync_slug",
					"chatProvider":  "openrouter",
					"chatModel":     "wrong-model",
					"agentProvider": "openrouter",
					"agentModel":    "wrong-model",
				},
			})
			return
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workspace/sync_slug/update":
			updateCalls++
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"workspace": map[string]interface{}{
					"id":   1,
					"slug": "sync_slug",
				},
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	o := NewOrchestrator(server.URL, "test-key", OrchestratorConfig{DefaultProvider: "openrouter"})
	err := o.SyncWorkspaceRuntimeModel(context.Background(), "sync_slug", "openai/gpt-4o-mini", "openrouter")
	if err == nil {
		t.Fatalf("expected sync verification error")
	}
	if updateCalls != 2 {
		t.Fatalf("expected 2 update calls, got %d", updateCalls)
	}
	if getCalls < 2 {
		t.Fatalf("expected at least 2 get calls, got %d", getCalls)
	}
}

func TestEnsureWorkspaceCreatesOn404(t *testing.T) {
	getCalls := 0
	createCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workspace/proj_1":
			getCalls++
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Workspace not found"})
			return
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workspace/new":
			createCalls++
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"workspace": map[string]interface{}{
					"id":   9,
					"slug": "proj_1",
					"name": "proj_1",
				},
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	o := NewOrchestrator(server.URL, "test-key", OrchestratorConfig{})
	ws, err := o.EnsureWorkspaceBySlug(context.Background(), "proj_1", "proj_1", "")
	if err != nil {
		t.Fatalf("ensure workspace failed: %v", err)
	}
	if ws == nil || ws.Slug != "proj_1" {
		t.Fatalf("unexpected workspace: %#v", ws)
	}
	if getCalls == 0 || createCalls == 0 {
		t.Fatalf("expected both get/create to be called, get=%d create=%d", getCalls, createCalls)
	}
}
