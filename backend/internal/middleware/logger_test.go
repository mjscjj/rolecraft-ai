package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger(LogLevelDebug, "/tmp/test.log")
	if logger == nil {
		t.Fatal("Expected logger to be created")
	}
	if logger.level != LogLevelDebug {
		t.Errorf("Expected level Debug, got %v", logger.level)
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		testLevel LogLevel
		shouldLog bool
	}{
		{"Debug logs Debug", LogLevelDebug, LogLevelDebug, true},
		{"Info logs Info", LogLevelInfo, LogLevelInfo, true},
		{"Warn logs Warn", LogLevelWarn, LogLevelWarn, true},
		{"Error logs Error", LogLevelError, LogLevelError, true},
		{"Info does not log Debug", LogLevelInfo, LogLevelDebug, false},
		{"Warn does not log Info", LogLevelWarn, LogLevelInfo, false},
		{"Error does not log Warn", LogLevelError, LogLevelWarn, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level, "/tmp/test.log")
			result := logger.shouldLog(tt.testLevel)
			if result != tt.shouldLog {
				t.Errorf("Expected shouldLog=%v, got %v", tt.shouldLog, result)
			}
		})
	}
}

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	logger := NewLogger(LogLevelDebug, "/tmp/test.log")
	router := gin.New()
	router.Use(RequestLogger(logger))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRecoveryLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	logger := NewLogger(LogLevelDebug, "/tmp/test.log")
	router := gin.New()
	router.Use(RecoveryLogger(logger))
	
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestLogEntryJSON(t *testing.T) {
	entry := LogEntry{
		Timestamp: "2026-02-27T12:00:00Z",
		Level:     LogLevelInfo,
		Message:   "Test message",
		RequestID: "12345",
		Method:    "GET",
		Path:      "/test",
		Status:    200,
		Latency:   "100ms",
		ClientIP:  "127.0.0.1",
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal log entry: %v", err)
	}

	var decoded LogEntry
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}

	if decoded.Message != entry.Message {
		t.Errorf("Expected message %s, got %s", entry.Message, decoded.Message)
	}
}

func TestGetCallerInfo(t *testing.T) {
	file, line := GetCallerInfo()
	if file == "unknown" {
		t.Error("Expected to get caller info")
	}
	if line <= 0 {
		t.Error("Expected positive line number")
	}
	t.Logf("Caller: %s:%d", file, line)
}

func BenchmarkRequestLogger(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	logger := NewLogger(LogLevelInfo, "/tmp/test.log")
	router := gin.New()
	router.Use(RequestLogger(logger))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func TestLoggerConcurrent(t *testing.T) {
	logger := NewLogger(LogLevelDebug, "/tmp/test.log")
	
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("concurrent test", map[string]interface{}{
				"goroutine": id,
			})
			done <- true
		}(i)
	}
	
	for i := 0; i < 10; i++ {
		<-done
	}
}
