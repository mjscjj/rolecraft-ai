#!/bin/bash

# RoleCraft AI 监控脚本
# 用法：./scripts/monitor.sh [check|start|status|alert]

set -e

# 配置
BACKEND_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:5173"
LOG_DIR="logs"
ALERT_EMAIL=""  # 配置告警邮箱
CHECK_INTERVAL=300  # 检查间隔 (秒)

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log() {
    echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

log_info() {
    log "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    log "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    log "${RED}[ERROR]${NC} $1"
}

# 确保日志目录存在
mkdir -p "$LOG_DIR"

# 检查后端服务
check_backend() {
    log_info "检查后端服务..."
    
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "$BACKEND_URL/health" 2>/dev/null || echo "000")
    
    if [ "$RESPONSE" = "200" ]; then
        log_info "✅ 后端服务正常 (HTTP $RESPONSE)"
        return 0
    else
        log_error "❌ 后端服务异常 (HTTP $RESPONSE)"
        return 1
    fi
}

# 检查前端服务
check_frontend() {
    log_info "检查前端服务..."
    
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "$FRONTEND_URL" 2>/dev/null || echo "000")
    
    if [ "$RESPONSE" = "200" ]; then
        log_info "✅ 前端服务正常 (HTTP $RESPONSE)"
        return 0
    else
        log_error "❌ 前端服务异常 (HTTP $RESPONSE)"
        return 1
    fi
}

# 检查 API 响应时间
check_response_time() {
    log_info "检查 API 响应时间..."
    
    # 使用 curl 的内置时间测量
    TIME_TOTAL=$(curl -s -o /dev/null -w "%{time_total}" --connect-timeout 5 "$BACKEND_URL/health" 2>/dev/null || echo "999")
    TIME_MS=$(echo "$TIME_TOTAL" | awk '{printf "%.0f", $1 * 1000}')
    
    if [ "$TIME_MS" -lt 100 ] 2>/dev/null; then
        log_info "✅ 响应时间：${TIME_MS}ms (< 100ms)"
        return 0
    else
        log_warn "⚠️ 响应时间：${TIME_MS}ms (>= 100ms)"
        return 1
    fi
}

# 检查磁盘空间
check_disk_space() {
    log_info "检查磁盘空间..."
    
    USAGE=$(df -h . | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$USAGE" -lt 80 ]; then
        log_info "✅ 磁盘使用率：${USAGE}% (< 80%)"
        return 0
    elif [ "$USAGE" -lt 90 ]; then
        log_warn "⚠️ 磁盘使用率：${USAGE}% (>= 80%)"
        return 1
    else
        log_error "❌ 磁盘使用率：${USAGE}% (>= 90%)"
        return 2
    fi
}

# 检查进程状态
check_processes() {
    log_info "检查进程状态..."
    
    # 检查后端服务 (通过 HTTP 检查更可靠)
    if curl -s --connect-timeout 2 "$BACKEND_URL/health" > /dev/null 2>&1; then
        log_info "✅ 后端服务可访问"
    else
        log_error "❌ 后端服务不可访问"
        return 1
    fi
    
    # 检查前端进程
    if pgrep -f "vite" > /dev/null 2>&1 || curl -s --connect-timeout 2 "$FRONTEND_URL" > /dev/null 2>&1; then
        log_info "✅ 前端服务可访问"
    else
        log_warn "⚠️ 前端服务不可访问 (可能是生产环境)"
    fi
    
    return 0
}

# 发送告警 (可扩展为邮件/短信/Slack)
send_alert() {
    local MESSAGE="$1"
    log_error "🚨 告警：$MESSAGE"
    
    # 记录到告警日志
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $MESSAGE" >> "$LOG_DIR/alerts.log"
    
    # TODO: 集成邮件/短信/Slack 告警
    # if [ -n "$ALERT_EMAIL" ]; then
    #     echo "$MESSAGE" | mail -s "RoleCraft AI 告警" "$ALERT_EMAIL"
    # fi
}

# 运行所有检查
run_checks() {
    local FAILED=0
    
    log_info "=========================================="
    log_info "RoleCraft AI 健康检查"
    log_info "=========================================="
    
    check_backend || ((FAILED++))
    check_frontend || ((FAILED++))
    check_response_time || ((FAILED++))
    check_disk_space || ((FAILED++))
    check_processes || ((FAILED++))
    
    log_info "=========================================="
    
    if [ "$FAILED" -eq 0 ]; then
        log_info "✅ 所有检查通过！"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 健康检查通过" >> "$LOG_DIR/health.log"
        return 0
    else
        log_error "❌ $FAILED 项检查失败"
        send_alert "$FAILED 项健康检查失败"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 健康检查失败：$FAILED 项" >> "$LOG_DIR/health.log"
        return 1
    fi
}

# 显示服务状态
show_status() {
    echo ""
    echo "=========================================="
    echo "RoleCraft AI 服务状态"
    echo "=========================================="
    echo ""
    
    # 后端状态
    echo "后端服务 ($BACKEND_URL):"
    if curl -s "$BACKEND_URL/health" > /dev/null 2>&1; then
        echo "  状态：✅ 运行中"
        echo "  响应：$(curl -s "$BACKEND_URL/health")"
    else
        echo "  状态：❌ 未运行"
    fi
    echo ""
    
    # 前端状态
    echo "前端服务 ($FRONTEND_URL):"
    if curl -s "$FRONTEND_URL" > /dev/null 2>&1; then
        echo "  状态：✅ 运行中"
    else
        echo "  状态：❌ 未运行"
    fi
    echo ""
    
    # 进程信息
    echo "进程信息:"
    ps aux | grep -E "rolecraft|vite" | grep -v grep | awk '{print "  " $11 " " $12}' || echo "  无相关进程"
    echo ""
    
    # 日志文件
    echo "日志文件:"
    ls -lh "$LOG_DIR"/*.log 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}' || echo "  无日志文件"
    echo ""
    
    # 最近告警
    if [ -f "$LOG_DIR/alerts.log" ]; then
        echo "最近告警:"
        tail -5 "$LOG_DIR/alerts.log" | sed 's/^/  /'
    else
        echo "最近告警：无"
    fi
    echo ""
    echo "=========================================="
}

# 启动监控循环
start_monitoring() {
    log_info "启动监控循环 (间隔：${CHECK_INTERVAL}秒)"
    log_info "按 Ctrl+C 停止"
    
    while true; do
        run_checks || true
        sleep "$CHECK_INTERVAL"
    done
}

# 主函数
case "${1:-check}" in
    check)
        run_checks
        ;;
    start)
        start_monitoring
        ;;
    status)
        show_status
        ;;
    alert)
        send_alert "${2:-测试告警}"
        ;;
    *)
        echo "用法：$0 {check|start|status|alert}"
        echo ""
        echo "命令:"
        echo "  check   - 运行一次健康检查"
        echo "  start   - 启动持续监控"
        echo "  status  - 显示服务状态"
        echo "  alert   - 发送测试告警"
        exit 1
        ;;
esac
