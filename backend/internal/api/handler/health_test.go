package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"rolecraft-ai/internal/config"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database")
	}
	return db
}

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db := setupTestDB()
	cfg := &config.Config{
		AnythingLLMURL: "http://localhost:3001",
		AnythingLLMKey: "test-key",
	}
	
	handler := NewHealthHandler(db, cfg)
	router := gin.New()
	router.GET("/api/v1/health", handler.Health)

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 200 or 503, got %d", w.Code)
	}

	var result HealthStatus
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}
	if result.Checks == nil {
		t.Error("Expected checks to be set")
	}
}

func TestReadyHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db := setupTestDB()
	cfg := &config.Config{}
	
	handler := NewHealthHandler(db, cfg)
	router := gin.New()
	router.GET("/api/v1/ready", handler.Ready)

	req, _ := http.NewRequest("GET", "/api/v1/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestLiveHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db := setupTestDB()
	cfg := &config.Config{}
	
	handler := NewHealthHandler(db, cfg)
	router := gin.New()
	router.GET("/api/v1/live", handler.Live)

	req, _ := http.NewRequest("GET", "/api/v1/live", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMetricsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db := setupTestDB()
	cfg := &config.Config{}
	
	handler := NewHealthHandler(db, cfg)
	router := gin.New()
	router.GET("/api/v1/metrics", handler.Metrics)

	req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDBStatsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db := setupTestDB()
	cfg := &config.Config{}
	
	handler := NewHealthHandler(db, cfg)
	router := gin.New()
	router.GET("/api/v1/db/stats", handler.DBStats)

	req, _ := http.NewRequest("GET", "/api/v1/db/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats DatabaseStats
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// 内存数据库应该有连接
	if stats.OpenConnections < 0 {
		t.Error("Expected non-negative open connections")
	}
}

func TestSimpleHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/health", SimpleHealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status=ok, got %s", result["status"])
	}
}

func TestHealthStatusJSON(t *testing.T) {
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: "2026-02-27T12:00:00Z",
		Version:   "1.0.0",
		Uptime:    "24h",
		Checks: map[string]CheckResult{
			"database": {
				Status:  "healthy",
				Message: "database ok",
				Latency: "10ms",
			},
		},
	}

	jsonBytes, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Failed to marshal health status: %v", err)
	}

	var decoded HealthStatus
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal health status: %v", err)
	}

	if decoded.Status != status.Status {
		t.Errorf("Expected status %s, got %s", status.Status, decoded.Status)
	}
}

func TestCheckResult(t *testing.T) {
	result := CheckResult{
		Status:  "healthy",
		Message: "All checks passed",
		Latency: "5ms",
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal check result: %v", err)
	}

	var decoded CheckResult
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal check result: %v", err)
	}

	if decoded.Status != result.Status {
		t.Errorf("Expected status %s, got %s", result.Status, decoded.Status)
	}
}

func TestNewHealthHandler(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{}
	
	handler := NewHealthHandler(db, cfg)
	
	if handler == nil {
		t.Fatal("Expected handler to be created")
	}
	if handler.db == nil {
		t.Error("Expected db to be set")
	}
	if handler.config == nil {
		t.Error("Expected config to be set")
	}
}

func TestResponseHelpers(t *testing.T) {
	// Test SuccessResponse
	resp := SuccessResponse(map[string]string{"key": "value"})
	if !resp.Success {
		t.Error("Expected Success=true")
	}
	if resp.Error != "" {
		t.Error("Expected Error to be empty")
	}

	// Test ErrorResponse
	errResp := ErrorResponse("something went wrong")
	if errResp.Success {
		t.Error("Expected Success=false")
	}
	if errResp.Error != "something went wrong" {
		t.Errorf("Expected error message, got %s", errResp.Error)
	}
}

func TestJSONResponse(t *testing.T) {
	resp := NewJSONResponse(200, "success", map[string]string{"key": "value"})
	
	if resp.Code != 200 {
		t.Errorf("Expected code 200, got %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("Expected message 'success', got %s", resp.Message)
	}
	
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}
	
	if len(jsonBytes) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestStringResponse(t *testing.T) {
	resp := NewStringResponse(200, "OK")
	
	if resp.Code != 200 {
		t.Errorf("Expected code 200, got %d", resp.Code)
	}
	if resp.Message != "OK" {
		t.Errorf("Expected message 'OK', got %s", resp.Message)
	}
}

func BenchmarkHealthHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	db := setupTestDB()
	cfg := &config.Config{}
	handler := NewHealthHandler(db, cfg)
	
	router := gin.New()
	router.GET("/api/v1/health", handler.Health)

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
