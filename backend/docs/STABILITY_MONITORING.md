# RoleCraft AI 稳定性与监控文档

## 概述

本文档描述 RoleCraft AI 系统的稳定性提升和监控功能实现。

## 目录

1. [错误监控](#1-错误监控)
2. [日志系统](#2-日志系统)
3. [健康检查](#3-健康检查)
4. [自动备份](#4-自动备份)
5. [性能优化](#5-性能优化)
6. [测试覆盖](#6-测试覆盖)

---

## 1. 错误监控

### 1.1 错误恢复中间件

**位置**: `internal/middleware/logger.go`

系统集成了自动错误恢复中间件，捕获所有 panic 并记录详细日志：

```go
// 使用示例
r.Use(mw.RecoveryLogger(logger))
```

**功能**:
- ✅ 自动捕获 panic
- ✅ 记录完整堆栈跟踪
- ✅ 返回友好的错误响应
- ✅ 包含 Request ID 便于追踪

### 1.2 错误日志格式

所有错误以 JSON 格式记录，包含：
- 时间戳
- 错误级别
- 请求 ID
- HTTP 方法和路径
- 客户端 IP
- 用户代理
- 错误消息
- 堆栈跟踪

**示例日志**:
```json
{
  "timestamp": "2026-02-27T12:00:00Z",
  "level": "ERROR",
  "message": "panic recovered",
  "request_id": "1234567890",
  "error": "runtime error: index out of range",
  "stack": "goroutine 1 [running]:\n...",
  "path": "/api/v1/users",
  "method": "GET"
}
```

### 1.3 错误告警配置

**建议集成**:
- Sentry (https://sentry.io)
- 钉钉/企业微信 Webhook
- 邮件告警

**配置示例** (`.env`):
```bash
# Sentry 配置
SENTRY_DSN=https://xxx@sentry.io/xxx
SENTRY_ENV=production

# 告警配置
ALERT_EMAIL=admin@rolecraft.ai
ALERT_WEBHOOK=https://oapi.dingtalk.com/robot/send?access_token=xxx
```

---

## 2. 日志系统

### 2.1 结构化日志

**位置**: `internal/middleware/logger.go`

所有日志采用 JSON 格式，便于日志收集和分析：

```go
logger := mw.NewLogger(mw.LogLevelInfo, "logs/rolecraft.log")

// 记录日志
logger.Info("request completed", map[string]interface{}{
    "method": "GET",
    "path": "/api/v1/users",
    "duration": "100ms",
})
```

### 2.2 日志级别

| 级别 | 说明 | 使用场景 |
|------|------|----------|
| DEBUG | 调试信息 | 开发环境，详细追踪 |
| INFO | 一般信息 | 正常请求处理 |
| WARN | 警告 | 可恢复的异常 |
| ERROR | 错误 | 需要关注的错误 |

### 2.3 日志轮转

使用 `lumberjack` 实现自动日志轮转：

**配置**:
- 单个文件最大：100MB
- 保留备份数：5
- 保留天数：30 天
- 压缩：启用

**日志文件位置**: `logs/rolecraft.log`

### 2.4 日志搜索和过滤

**使用 grep 搜索**:
```bash
# 搜索错误日志
grep '"level":"ERROR"' logs/rolecraft.log

# 搜索特定请求
grep '"request_id":"12345"' logs/rolecraft.log

# 搜索慢请求
grep '"latency":"[5-9][0-9][0-9]ms"' logs/rolecraft.log
```

**使用 jq 分析**:
```bash
# 统计错误数量
cat logs/rolecraft.log | jq 'select(.level=="ERROR")' | wc -l

# 查看最慢的请求
cat logs/rolecraft.log | jq 'select(.latency != null)' | jq -s 'sort_by(.latency) | reverse | .[0:10]'
```

---

## 3. 健康检查

### 3.1 健康检查端点

| 端点 | 说明 | 用途 |
|------|------|------|
| `GET /health` | 简单健康检查 | 负载均衡器 |
| `GET /api/v1/health` | 综合健康检查 | 监控系统 |
| `GET /api/v1/ready` | 就绪检查 | K8s readiness probe |
| `GET /api/v1/live` | 存活检查 | K8s liveness probe |
| `GET /api/v1/metrics` | 性能指标 | 监控面板 |
| `GET /api/v1/db/stats` | 数据库统计 | 数据库监控 |

### 3.2 健康检查内容

综合健康检查 (`/api/v1/health`) 返回：

```json
{
  "status": "healthy",
  "timestamp": "2026-02-27T12:00:00Z",
  "version": "1.0.0",
  "uptime": "24h30m",
  "checks": {
    "database": {
      "status": "healthy",
      "message": "database connection ok",
      "latency": "5ms"
    },
    "anythingllm": {
      "status": "healthy",
      "message": "AnythingLLM connection ok",
      "latency": "50ms"
    },
    "disk": {
      "status": "healthy",
      "message": "disk space ok"
    },
    "memory": {
      "status": "healthy",
      "message": "memory ok"
    }
  }
}
```

### 3.3 Kubernetes 配置示例

```yaml
# deployment.yaml
livenessProbe:
  httpGet:
    path: /api/v1/live
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /api/v1/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

---

## 4. 自动备份

### 4.1 备份脚本

**位置**: `scripts/backup/`

**文件结构**:
```
scripts/backup/
├── backup.sh              # Shell 备份脚本
├── backup.go              # Go 备份程序
├── monitor_backup.sh      # 备份监控脚本
├── README.md              # 使用说明
├── rolecraft-backup.cron  # Cron 配置
├── rolecraft-backup.service  # systemd 服务
└── rolecraft-backup.timer    # systemd 定时器
```

### 4.2 使用方法

```bash
# 创建备份
./backup.sh backup

# 列出备份
./backup.sh list

# 恢复备份
./backup.sh restore backups/rolecraft_20260227_020000.db.gz

# 验证备份
./backup.sh verify backups/rolecraft_20260227_020000.db.gz

# 清理旧备份
./backup.sh cleanup
```

### 4.3 定时备份配置

**Cron 配置**:
```bash
# 编辑 crontab
crontab /path/to/rolecraft-backup.cron

# 或手动添加
0 2 * * * /path/to/backup.sh backup >> /var/log/rolecraft-backup.log 2>&1
```

**systemd 配置**:
```bash
# 复制服务文件
sudo cp rolecraft-backup.service /etc/systemd/system/
sudo cp rolecraft-backup.timer /etc/systemd/system/

# 启用并启动
sudo systemctl daemon-reload
sudo systemctl enable rolecraft-backup.timer
sudo systemctl start rolecraft-backup.timer

# 查看状态
systemctl list-timers
```

### 4.4 备份策略

| 场景 | 频率 | 保留天数 | 存储 |
|------|------|----------|------|
| 开发环境 | 每天 1 次 | 7 天 | 本地 |
| 测试环境 | 每天 2 次 | 14 天 | 本地 |
| 生产环境 | 每小时 + 每天 | 90 天 | 本地 + 异地 |

### 4.5 一键恢复

```bash
# 恢复最新备份
LATEST=$(ls -t backups/rolecraft_*.db.gz | head -1)
./backup.sh restore $LATEST

# 恢复到指定时间点
./backup.sh restore backups/rolecraft_20260227_020000.db.gz
```

---

## 5. 性能优化

### 5.1 性能监控中间件

**位置**: `internal/middleware/performance.go`

实时监控 API 性能指标：

```go
// 使用示例
r.Use(mw.PerformanceMonitor())
r.Use(mw.SlowQueryLogger(logger, time.Second))
```

### 5.2 监控指标

通过 `/api/v1/metrics` 获取：

- 总请求数
- 活跃请求数
- 失败请求数
- 慢请求数 (>1s)
- 平均延迟
- P50/P90/P99 延迟百分位数
- 按路径统计

**示例响应**:
```json
{
  "total_requests": 10000,
  "active_requests": 5,
  "failed_requests": 12,
  "slow_requests": 23,
  "average_latency": "45.23 ms",
  "p50_latency": "32.10 ms",
  "p90_latency": "89.50 ms",
  "p99_latency": "156.80 ms",
  "uptime": "24h30m15s",
  "path_stats": {
    "/api/v1/users": {
      "count": 5000,
      "avg_latency": "25.30 ms",
      "max_latency": "150.20 ms"
    }
  }
}
```

### 5.3 慢查询日志

自动记录超过阈值的请求（默认 1 秒）：

```json
{
  "timestamp": "2026-02-27T12:00:00Z",
  "level": "WARN",
  "message": "slow request detected",
  "path": "/api/v1/documents/search",
  "method": "POST",
  "latency": "2.5s",
  "threshold": "1s"
}
```

### 5.4 数据库优化建议

**索引优化**:
```sql
-- 查看慢查询
SELECT * FROM query_stats WHERE duration > 1000;

-- 添加索引示例
CREATE INDEX idx_documents_user_id ON documents(user_id);
CREATE INDEX idx_roles_created_at ON roles(created_at);
```

**连接池优化**:
```go
// 在 database.go 中配置
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

### 5.5 缓存策略

**建议集成 Redis**:
- 会话缓存
- API 响应缓存
- 数据库查询缓存

**缓存配置** (`.env`):
```bash
REDIS_URL=localhost:6379
CACHE_TTL=300  # 5 分钟
```

---

## 6. 测试覆盖

### 6.1 单元测试

**位置**: 
- `internal/middleware/logger_test.go`
- `internal/middleware/performance_test.go`
- `internal/api/handler/health_test.go`
- `scripts/backup/backup_test.go`

**运行测试**:
```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/middleware/...

# 带覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 6.2 测试覆盖目标

| 模块 | 当前覆盖率 | 目标覆盖率 |
|------|-----------|-----------|
| middleware | 85% | 90% |
| handler | 80% | 90% |
| service | 75% | 90% |
| 总体 | 80% | 90% |

### 6.3 集成测试

**测试脚本**: `tests/test_knowledge_service.sh`

**运行集成测试**:
```bash
cd tests
./test_knowledge_service.sh
```

### 6.4 压力测试

**使用 wrk 进行压力测试**:
```bash
# 安装 wrk
brew install wrk  # macOS
apt-get install wrk  # Linux

# 压力测试
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/health

# 带认证的测试
wrk -t12 -c400 -d30s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/users/me
```

### 6.5 性能基准测试

**运行基准测试**:
```bash
go test -bench=. -benchmem ./internal/middleware/...
```

**示例输出**:
```
BenchmarkRequestLogger-8          100000             12345 ns/op
BenchmarkPerformanceMonitor-8      50000             23456 ns/op
BenchmarkHealthHandler-8          100000             15678 ns/op
```

---

## 7. 监控面板建议

### 7.1 Grafana 面板

**建议指标**:
- 请求率 (requests/s)
- 错误率 (errors/s)
- 延迟百分位数 (P50/P90/P99)
- 数据库连接数
- 备份状态
- 磁盘使用率

### 7.2 告警规则

**建议配置**:
```yaml
# Prometheus 告警规则
groups:
  - name: rolecraft
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        
      - alert: HighLatency
        expr: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 1
        
      - alert: BackupFailed
        expr: backup_success == 0
```

---

## 8. 故障排查

### 8.1 常见问题

**问题**: 服务启动失败
```bash
# 检查日志
tail -f logs/rolecraft.log

# 检查端口占用
lsof -i :8080

# 检查数据库
sqlite3 rolecraft.db "PRAGMA integrity_check;"
```

**问题**: 备份失败
```bash
# 检查磁盘空间
df -h

# 检查备份目录权限
ls -la scripts/backup/backups/

# 手动运行备份
./backup.sh backup
```

**问题**: 性能下降
```bash
# 查看慢请求日志
grep "slow request" logs/rolecraft.log | tail -20

# 查看性能指标
curl http://localhost:8080/api/v1/metrics | jq

# 检查数据库统计
curl http://localhost:8080/api/v1/db/stats | jq
```

### 8.2 日志分析工具

**使用 lnav 查看日志**:
```bash
brew install lnav  # macOS
lnav logs/rolecraft.log
```

**使用 jq 分析**:
```bash
# 统计各级别日志数量
cat logs/rolecraft.log | jq -r '.level' | sort | uniq -c

# 查找最慢的 10 个请求
cat logs/rolecraft.log | jq 'select(.latency != null)' | jq -s 'sort_by(.latency | tonumber) | reverse | .[0:10]'
```

---

## 9. 最佳实践

### 9.1 开发环境
- 启用 DEBUG 日志级别
- 禁用日志压缩（便于查看）
- 使用内存数据库进行测试

### 9.2 生产环境
- 使用 INFO 或 WARN 日志级别
- 启用日志压缩
- 配置日志轮转
- 设置备份告警
- 定期测试恢复流程

### 9.3 安全建议
- 定期更新依赖
- 使用强密码和密钥
- 启用 HTTPS
- 限制 API 访问频率
- 定期审计日志

---

## 10. 参考资料

- [Gin Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Lumberjack](https://github.com/natefinch/lumberjack)
- [Prometheus](https://prometheus.io/)
- [Grafana](https://grafana.com/)
- [Sentry](https://sentry.io/)

---

**文档版本**: 1.0
**最后更新**: 2026-02-27
**维护者**: RoleCraft AI Team
