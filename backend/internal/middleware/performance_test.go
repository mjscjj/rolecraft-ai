package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestPerformanceMonitor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 重置指标
	GlobalMetrics.ResetMetrics()
	
	router := gin.New()
	router.Use(PerformanceMonitor())
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if GlobalMetrics.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", GlobalMetrics.TotalRequests)
	}
}

func TestSlowQueryLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	logger := NewLogger(LogLevelDebug, "/tmp/test.log")
	router := gin.New()
	router.Use(SlowQueryLogger(logger, 10*time.Millisecond))
	
	// 快速请求
	router.GET("/fast", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "fast"})
	})
	
	// 慢请求
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.JSON(200, gin.H{"message": "slow"})
	})

	// 测试快速请求
	req, _ := http.NewRequest("GET", "/fast", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 测试慢请求
	req, _ = http.NewRequest("GET", "/slow", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCalculatePercentiles(t *testing.T) {
	metrics := &PerformanceMetrics{
		latencies:  []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		maxSamples: 1000,
	}
	
	metrics.CalculatePercentiles()
	
	// 百分位数计算使用索引，对于 10 个元素：
	// P50: index = 10 * 0.50 = 5 -> 60
	// P90: index = 10 * 0.90 = 9 -> 100
	// P99: index = 10 * 0.99 = 9 (截断) -> 100
	if metrics.P50Latency != 60 {
		t.Errorf("Expected P50=60, got %f", metrics.P50Latency)
	}
	if metrics.P90Latency != 100 {
		t.Errorf("Expected P90=100, got %f", metrics.P90Latency)
	}
	if metrics.P99Latency != 100 {
		t.Errorf("Expected P99=100, got %f", metrics.P99Latency)
	}
}

func TestGetMetricsJSON(t *testing.T) {
	metrics := &PerformanceMetrics{
		TotalRequests:  100,
		ActiveRequests: 5,
		FailedRequests: 2,
		SlowRequests:   10,
		latencies:      []float64{50, 100, 150, 200},
		PathStats: map[string]*PathStats{
			"/api/test": {
				Count:      10,
				AvgLatency: 100.5,
				MaxLatency: 200.0,
			},
		},
	}
	
	result := metrics.GetMetricsJSON()
	
	if result["total_requests"] != int64(100) {
		t.Errorf("Expected total_requests=100, got %v", result["total_requests"])
	}
	if result["active_requests"] != int64(5) {
		t.Errorf("Expected active_requests=5, got %v", result["active_requests"])
	}
}

func TestResetMetrics(t *testing.T) {
	metrics := &PerformanceMetrics{
		TotalRequests:  100,
		ActiveRequests: 5,
		FailedRequests: 2,
		latencies:      []float64{10, 20, 30},
		PathStats: map[string]*PathStats{
			"/test": {Count: 5},
		},
	}
	
	metrics.ResetMetrics()
	
	if metrics.TotalRequests != 0 {
		t.Errorf("Expected TotalRequests=0, got %d", metrics.TotalRequests)
	}
	if metrics.ActiveRequests != 0 {
		t.Errorf("Expected ActiveRequests=0, got %d", metrics.ActiveRequests)
	}
	if len(metrics.latencies) != 0 {
		t.Errorf("Expected empty latencies, got %d items", len(metrics.latencies))
	}
	if len(metrics.PathStats) != 0 {
		t.Errorf("Expected empty PathStats, got %d items", len(metrics.PathStats))
	}
}

func TestPathStats(t *testing.T) {
	metrics := &PerformanceMetrics{
		PathStats:  make(map[string]*PathStats),
		maxSamples: 1000,
	}
	
	// 模拟多个请求
	for i := 0; i < 5; i++ {
		metrics.mu.Lock()
		if _, exists := metrics.PathStats["/test"]; !exists {
			metrics.PathStats["/test"] = &PathStats{}
		}
		stats := metrics.PathStats["/test"]
		stats.Count++
		stats.TotalLatency += 100.0
		stats.AvgLatency = stats.TotalLatency / float64(stats.Count)
		if 100.0 > stats.MaxLatency {
			stats.MaxLatency = 100.0
		}
		metrics.mu.Unlock()
	}
	
	stats := metrics.PathStats["/test"]
	if stats.Count != 5 {
		t.Errorf("Expected count=5, got %d", stats.Count)
	}
	if stats.AvgLatency != 100.0 {
		t.Errorf("Expected avg=100.0, got %f", stats.AvgLatency)
	}
}

func BenchmarkPerformanceMonitor(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	GlobalMetrics.ResetMetrics()
	router := gin.New()
	router.Use(PerformanceMonitor())
	
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

func TestConcurrentMetrics(t *testing.T) {
	GlobalMetrics.ResetMetrics()
	
	done := make(chan bool, 100)
	
	for i := 0; i < 100; i++ {
		go func() {
			GlobalMetrics.mu.Lock()
			GlobalMetrics.TotalRequests++
			GlobalMetrics.mu.Unlock()
			done <- true
		}()
	}
	
	for i := 0; i < 100; i++ {
		<-done
	}
	
	if GlobalMetrics.TotalRequests != 100 {
		t.Errorf("Expected 100 requests, got %d", GlobalMetrics.TotalRequests)
	}
}
