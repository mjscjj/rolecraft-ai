package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// LogEntry 结构化日志条目
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	RequestID   string                 `json:"request_id,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Path        string                 `json:"path,omitempty"`
	Status      int                    `json:"status,omitempty"`
	Latency     string                 `json:"latency,omitempty"`
	ClientIP    string                 `json:"client_ip,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// Logger 日志记录器
type Logger struct {
	writer io.Writer
	level  LogLevel
}

// NewLogger 创建新的日志记录器
func NewLogger(level LogLevel, logPath string) *Logger {
	// 使用 lumberjack 实现日志轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	// 同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, lumberjackLogger)

	return &Logger{
		writer: multiWriter,
		level:  level,
	}
}

// log 内部日志方法
func (l *Logger) log(level LogLevel, message string, extra map[string]interface{}) {
	if l.shouldLog(level) {
		entry := LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     level,
			Message:   message,
			Extra:     extra,
		}

		jsonBytes, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
			return
		}

		fmt.Fprintln(l.writer, string(jsonBytes))
	}
}

func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}

	return levels[level] >= levels[l.level]
}

// Debug 记录调试日志
func (l *Logger) Debug(message string, extra map[string]interface{}) {
	l.log(LogLevelDebug, message, extra)
}

// Info 记录信息日志
func (l *Logger) Info(message string, extra map[string]interface{}) {
	l.log(LogLevelInfo, message, extra)
}

// Warn 记录警告日志
func (l *Logger) Warn(message string, extra map[string]interface{}) {
	l.log(LogLevelWarn, message, extra)
}

// Error 记录错误日志
func (l *Logger) Error(message string, extra map[string]interface{}) {
	l.log(LogLevelError, message, extra)
}

// GetCallerInfo 获取调用者信息
func GetCallerInfo() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", 0
	}
	return filepath.Base(file), line
}

// RequestLogger Gin 请求日志中间件
func RequestLogger(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 生成请求 ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			c.Header("X-Request-ID", requestID)
		}
		c.Set("requestID", requestID)

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 记录日志
		extra := map[string]interface{}{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       path,
			"status":     c.Writer.Status(),
			"latency":    latency.String(),
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}

		if query != "" {
			extra["query"] = query
		}

		// 根据状态码选择日志级别
		if c.Writer.Status() >= 500 {
			logger.Error("server error", extra)
		} else if c.Writer.Status() >= 400 {
			logger.Warn("client error", extra)
		} else {
			logger.Info("request completed", extra)
		}
	}
}

// RecoveryLogger 错误恢复中间件（带日志）
func RecoveryLogger(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				stackTrace := string(buf[:n])

				requestID, _ := c.Get("requestID")
				
				// 记录错误日志
				logger.Error("panic recovered", map[string]interface{}{
					"request_id": requestID,
					"error":      fmt.Sprintf("%v", err),
					"stack":      stackTrace,
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
				})

				// 返回错误响应
				c.JSON(500, gin.H{
					"error":      "Internal server error",
					"request_id": requestID,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
