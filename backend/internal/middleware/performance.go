package middleware

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	mu sync.RWMutex

	// 请求统计
	TotalRequests   int64   `json:"total_requests"`
	ActiveRequests  int64   `json:"active_requests"`
	FailedRequests  int64   `json:"failed_requests"`
	AverageLatency  float64 `json:"average_latency_ms"`
	SlowRequests    int64   `json:"slow_requests"` // > 1s 的请求

	// 延迟百分位数
	P50Latency float64 `json:"p50_latency_ms"`
	P90Latency float64 `json:"p90_latency_ms"`
	P99Latency float64 `json:"p99_latency_ms"`

	// 最近延迟样本（用于计算百分位数）
	latencies []float64
	maxSamples int

	// 按路径统计
	PathStats map[string]*PathStats `json:"path_stats,omitempty"`

	// 启动时间
	StartTime time.Time `json:"start_time"`
}

// PathStats 按路径的性能统计
type PathStats struct {
	Count        int64   `json:"count"`
	AvgLatency   float64 `json:"avg_latency_ms"`
	MaxLatency   float64 `json:"max_latency_ms"`
	TotalLatency float64 `json:"-"`
}

// 全局性能指标
var GlobalMetrics = &PerformanceMetrics{
	maxSamples: 1000,
	latencies:  make([]float64, 0, 1000),
	PathStats:  make(map[string]*PathStats),
	StartTime:  time.Now(),
}

// PerformanceMonitor 性能监控中间件
func PerformanceMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 增加活跃请求数
		GlobalMetrics.mu.Lock()
		GlobalMetrics.ActiveRequests++
		GlobalMetrics.mu.Unlock()

		defer func() {
			latency := time.Since(start)
			latencyMs := float64(latency.Nanoseconds()) / 1e6

			GlobalMetrics.mu.Lock()
			defer GlobalMetrics.mu.Unlock()

			GlobalMetrics.ActiveRequests--
			GlobalMetrics.TotalRequests++

			if c.Writer.Status() >= 400 {
				GlobalMetrics.FailedRequests++
			}

			// 记录延迟
			GlobalMetrics.latencies = append(GlobalMetrics.latencies, latencyMs)
			if len(GlobalMetrics.latencies) > GlobalMetrics.maxSamples {
				GlobalMetrics.latencies = GlobalMetrics.latencies[1:]
			}

			// 检查慢请求
			if latencyMs > 1000 {
				GlobalMetrics.SlowRequests++
			}

			// 更新平均延迟
			totalLatency := 0.0
			for _, l := range GlobalMetrics.latencies {
				totalLatency += l
			}
			GlobalMetrics.AverageLatency = totalLatency / float64(len(GlobalMetrics.latencies))

			// 更新路径统计
			path := c.Request.URL.Path
			if _, exists := GlobalMetrics.PathStats[path]; !exists {
				GlobalMetrics.PathStats[path] = &PathStats{}
			}
			stats := GlobalMetrics.PathStats[path]
			stats.Count++
			stats.TotalLatency += latencyMs
			stats.AvgLatency = stats.TotalLatency / float64(stats.Count)
			if latencyMs > stats.MaxLatency {
				stats.MaxLatency = latencyMs
			}
		}()

		c.Next()
	}
}

// SlowQueryLogger 慢查询日志中间件
func SlowQueryLogger(logger *Logger, threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		if latency > threshold {
			requestID, _ := c.Get("requestID")
			logger.Warn("slow request detected", map[string]interface{}{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"latency":    latency.String(),
				"threshold":  threshold.String(),
			})
		}
	}
}

// CalculatePercentiles 计算延迟百分位数
func (m *PerformanceMetrics) CalculatePercentiles() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.latencies) == 0 {
		return
	}

	// 复制并排序
	sorted := make([]float64, len(m.latencies))
	copy(sorted, m.latencies)
	sort.Float64s(sorted)

	// 计算百分位数
	p50Idx := int(float64(len(sorted)) * 0.50)
	p90Idx := int(float64(len(sorted)) * 0.90)
	p99Idx := int(float64(len(sorted)) * 0.99)

	if p50Idx >= len(sorted) {
		p50Idx = len(sorted) - 1
	}
	if p90Idx >= len(sorted) {
		p90Idx = len(sorted) - 1
	}
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}

	m.P50Latency = sorted[p50Idx]
	m.P90Latency = sorted[p90Idx]
	m.P99Latency = sorted[p99Idx]
}

// GetMetricsJSON 获取性能指标 JSON
func (m *PerformanceMetrics) GetMetricsJSON() map[string]interface{} {
	m.CalculatePercentiles()

	m.mu.RLock()
	defer m.mu.RUnlock()

	// 准备路径统计
	pathStats := make(map[string]interface{})
	for path, stats := range m.PathStats {
		pathStats[path] = map[string]interface{}{
			"count":         stats.Count,
			"avg_latency":   fmt.Sprintf("%.2f ms", stats.AvgLatency),
			"max_latency":   fmt.Sprintf("%.2f ms", stats.MaxLatency),
		}
	}

	return map[string]interface{}{
		"total_requests":   m.TotalRequests,
		"active_requests":  m.ActiveRequests,
		"failed_requests":  m.FailedRequests,
		"slow_requests":    m.SlowRequests,
		"average_latency":  fmt.Sprintf("%.2f ms", m.AverageLatency),
		"p50_latency":      fmt.Sprintf("%.2f ms", m.P50Latency),
		"p90_latency":      fmt.Sprintf("%.2f ms", m.P90Latency),
		"p99_latency":      fmt.Sprintf("%.2f ms", m.P99Latency),
		"uptime":           time.Since(m.StartTime).String(),
		"path_stats":       pathStats,
	}
}

// ResetMetrics 重置性能指标（用于测试）
func (m *PerformanceMetrics) ResetMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests = 0
	m.ActiveRequests = 0
	m.FailedRequests = 0
	m.SlowRequests = 0
	m.AverageLatency = 0
	m.P50Latency = 0
	m.P90Latency = 0
	m.P99Latency = 0
	m.latencies = make([]float64, 0, m.maxSamples)
	m.PathStats = make(map[string]*PathStats)
	m.StartTime = time.Now()
}
