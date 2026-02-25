#!/bin/bash

# RoleCraft AI 前端服务启动脚本
# 用法：./scripts/start-frontend.sh [start|stop|restart|status]

set -e

FRONTEND_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && cd frontend && pwd)"
PID_FILE="$FRONTEND_DIR/.vite.pid"
LOG_FILE="$FRONTEND_DIR/../logs/frontend.log"

# 确保日志目录存在
mkdir -p "$(dirname "$LOG_FILE")"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

start() {
    if [ -f "$PID_FILE" ] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
        log "⚠️  前端服务已在运行 (PID: $(cat "$PID_FILE"))"
        return 0
    fi
    
    log "🚀 启动前端服务..."
    cd "$FRONTEND_DIR"
    
    # 后台启动 vite
    nohup npm run dev > "$LOG_FILE" 2>&1 &
    VITE_PID=$!
    echo $VITE_PID > "$PID_FILE"
    
    # 等待服务启动
    sleep 3
    
    if curl -s http://localhost:5173 > /dev/null 2>&1; then
        log "✅ 前端服务启动成功 (PID: $VITE_PID)"
        log "📍 访问地址：http://localhost:5173"
        return 0
    else
        log "❌ 前端服务启动失败"
        rm -f "$PID_FILE"
        return 1
    fi
}

stop() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 $PID 2>/dev/null; then
            log "🛑 停止前端服务 (PID: $PID)..."
            kill $PID 2>/dev/null || true
            sleep 2
            # 强制终止
            kill -9 $PID 2>/dev/null || true
            log "✅ 前端服务已停止"
        else
            log "⚠️  进程不存在，清理 PID 文件"
        fi
        rm -f "$PID_FILE"
    else
        # 尝试查找并停止 vite 进程
        VITE_PIDS=$(pgrep -f "vite" || true)
        if [ -n "$VITE_PIDS" ]; then
            log "🛑 停止 vite 进程 (PIDs: $VITE_PIDS)..."
            echo $VITE_PIDS | xargs kill 2>/dev/null || true
            sleep 2
            echo $VITE_PIDS | xargs kill -9 2>/dev/null || true
            log "✅ 前端服务已停止"
        else
            log "ℹ️  未找到运行中的前端服务"
        fi
    fi
}

restart() {
    stop
    sleep 2
    start
}

status() {
    if [ -f "$PID_FILE" ] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
        echo "✅ 前端服务运行中 (PID: $(cat "$PID_FILE"))"
        curl -s -o /dev/null -w "📊 响应时间：%{time_total}s\n" http://localhost:5173 2>/dev/null || true
        return 0
    else
        VITE_PIDS=$(pgrep -f "vite" || true)
        if [ -n "$VITE_PIDS" ]; then
            echo "⚠️  服务运行但 PID 文件丢失 (PIDs: $VITE_PIDS)"
            return 0
        else
            echo "❌ 前端服务未运行"
            return 1
        fi
    fi
}

# 主函数
case "${1:-start}" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    *)
        echo "用法：$0 {start|stop|restart|status}"
        exit 1
        ;;
esac
