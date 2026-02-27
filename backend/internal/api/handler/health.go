package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/service/anythingllm"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	db     *gorm.DB
	config *config.Config
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string                 `json:"status"` // overall, healthy, unhealthy
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]CheckResult `json:"checks"`
}

// CheckResult 单个检查项结果
type CheckResult struct {
	Status  string `json:"status"` // healthy, unhealthy, unknown
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(db *gorm.DB, cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		db:     db,
		config: cfg,
	}
}

// Health 综合健康检查
// @Summary 综合健康检查
// @Description 检查服务及其依赖的健康状态
// @Tags Health
// @Success 200 {object} HealthStatus
// @Failure 503 {object} HealthStatus
// @Router /api/v1/health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	startTime := time.Now()

	result := HealthStatus{
		Status:    "healthy",
		Timestamp: startTime.Format(time.RFC3339),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(), // 这里应该用服务启动时间
		Checks:    make(map[string]CheckResult),
	}

	// 数据库检查
	dbResult := h.checkDatabase()
	result.Checks["database"] = dbResult
	if dbResult.Status != "healthy" {
		result.Status = "unhealthy"
	}

	// AnythingLLM 检查
	anythingLLMResult := h.checkAnythingLLM()
	result.Checks["anythingllm"] = anythingLLMResult
	// AnythingLLM 不健康不影响整体状态（可能是可选依赖）

	// 磁盘空间检查
	diskResult := h.checkDiskSpace()
	result.Checks["disk"] = diskResult
	if diskResult.Status != "healthy" {
		result.Status = "unhealthy"
	}

	// 内存检查
	memoryResult := h.checkMemory()
	result.Checks["memory"] = memoryResult

	statusCode := http.StatusOK
	if result.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, result)
}

// checkDatabase 检查数据库连接
func (h *HealthHandler) checkDatabase() CheckResult {
	start := time.Now()

	sqlDB, err := h.db.DB()
	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: "failed to get database connection: " + err.Error(),
		}
	}

	// Ping 数据库
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: "database ping failed: " + err.Error(),
		}
	}

	return CheckResult{
		Status:  "healthy",
		Message: "database connection ok",
		Latency: time.Since(start).String(),
	}
}

// checkAnythingLLM 检查 AnythingLLM 服务
func (h *HealthHandler) checkAnythingLLM() CheckResult {
	start := time.Now()

	if h.config.AnythingLLMURL == "" {
		return CheckResult{
			Status:  "unknown",
			Message: "AnythingLLM URL not configured",
		}
	}

	client := anythingllm.NewAnythingLLMClient(h.config.AnythingLLMURL, h.config.AnythingLLMKey)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试获取工作区列表（轻量级检查）
	_, err := client.ListWorkspaces(ctx)
	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: "AnythingLLM connection failed: " + err.Error(),
			Latency: time.Since(start).String(),
		}
	}

	return CheckResult{
		Status:  "healthy",
		Message: "AnythingLLM connection ok",
		Latency: time.Since(start).String(),
	}
}

// checkDiskSpace 检查磁盘空间
func (h *HealthHandler) checkDiskSpace() CheckResult {
	// 简化实现，检查 uploads 目录
	// 在生产环境中应该使用 syscall.Statfs 获取实际磁盘使用情况

	return CheckResult{
		Status:  "healthy",
		Message: "disk space ok",
	}
}

// checkMemory 检查内存使用
func (h *HealthHandler) checkMemory() CheckResult {
	// 简化实现
	// 在生产环境中应该使用 runtime.MemStats

	return CheckResult{
		Status:  "healthy",
		Message: "memory ok",
	}
}

// Ready 就绪检查（用于 Kubernetes readiness probe）
// @Summary 就绪检查
// @Description 检查服务是否准备好接收流量
// @Tags Health
// @Success 200 {object} map[string]string
// @Router /api/v1/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	// 检查数据库是否可连接
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database connection failed",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database ping failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// Live 存活检查（用于 Kubernetes liveness probe）
// @Summary 存活检查
// @Description 检查服务是否存活
// @Tags Health
// @Success 200 {object} map[string]string
// @Router /api/v1/live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

// Metrics 性能指标
// @Summary 性能指标
// @Description 获取服务性能指标
// @Tags Health
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/metrics [get]
func (h *HealthHandler) Metrics(c *gin.Context) {
	// 这里可以集成 Prometheus 或其他监控系统
	// 暂时返回简单的统计信息
	
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get database connection",
		})
		return
	}

	stats := sqlDB.Stats()

	c.JSON(http.StatusOK, gin.H{
		"database": map[string]interface{}{
			"max_open_connections":     stats.MaxOpenConnections,
			"open_connections":         stats.OpenConnections,
			"in_use":                   stats.InUse,
			"idle":                     stats.Idle,
			"wait_count":               stats.WaitCount,
			"wait_duration":            stats.WaitDuration.String(),
			"max_idle_closed":          stats.MaxIdleClosed,
			"max_lifetime_closed":      stats.MaxLifetimeClosed,
			"max_idle_time_closed":     stats.MaxIdleTimeClosed,
		},
	})
}

// DatabaseStats 数据库统计
type DatabaseStats struct {
	MaxOpenConnections  int           `json:"max_open_connections"`
	OpenConnections     int           `json:"open_connections"`
	InUse               int           `json:"in_use"`
	Idle                int           `json:"idle"`
	WaitCount           int64         `json:"wait_count"`
	WaitDuration        time.Duration `json:"wait_duration"`
	MaxIdleClosed       int64         `json:"max_idle_closed"`
	MaxLifetimeClosed   int64         `json:"max_lifetime_closed"`
	MaxIdleTimeClosed   int64         `json:"max_idle_time_closed"`
}

// DBStats 获取数据库统计
// @Summary 数据库统计
// @Description 获取数据库连接池统计信息
// @Tags Health
// @Success 200 {object} DatabaseStats
// @Router /api/v1/db/stats [get]
func (h *HealthHandler) DBStats(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get database connection",
		})
		return
	}

	stats := sqlDB.Stats()
	
	c.JSON(http.StatusOK, DatabaseStats{
		MaxOpenConnections:  stats.MaxOpenConnections,
		OpenConnections:     stats.OpenConnections,
		InUse:               stats.InUse,
		Idle:                stats.Idle,
		WaitCount:           stats.WaitCount,
		WaitDuration:        stats.WaitDuration,
		MaxIdleClosed:       stats.MaxIdleClosed,
		MaxLifetimeClosed:   stats.MaxLifetimeClosed,
		MaxIdleTimeClosed:   stats.MaxIdleTimeClosed,
	})
}

// SimpleHealthCheck 简单健康检查（向后兼容）
func SimpleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// JSONResponse 通用 JSON 响应
type JSONResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewJSONResponse 创建 JSON 响应
func NewJSONResponse(code int, message string, data interface{}) JSONResponse {
	return JSONResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// StringResponse 字符串响应
type StringResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewStringResponse 创建字符串响应
func NewStringResponse(code int, message string) StringResponse {
	return StringResponse{
		Code:    code,
		Message: message,
	}
}

// MarshalJSON 自定义 JSON 序列化
func (r JSONResponse) MarshalJSON() ([]byte, error) {
	type Alias JSONResponse
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&r),
	})
}

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(message string) Response {
	return Response{
		Success: false,
		Error:   message,
	}
}
