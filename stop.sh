#!/bin/bash

echo "========================================="
echo "  🛑 RoleCraft AI - 停止服务"
echo "========================================="
echo ""

# 停止后端
if [ -f /tmp/rolecraft-backend.pid ]; then
    BACKEND_PID=$(cat /tmp/rolecraft-backend.pid)
    echo "📦 停止后端 (PID: $BACKEND_PID)..."
    kill $BACKEND_PID 2>/dev/null
    rm /tmp/rolecraft-backend.pid
    echo "   ✅ 后端已停止"
else
    echo "ℹ️  后端未运行"
fi

# 停止前端
if [ -f /tmp/rolecraft-frontend.pid ]; then
    FRONTEND_PID=$(cat /tmp/rolecraft-frontend.pid)
    echo "🎨 停止前端 (PID: $FRONTEND_PID)..."
    kill $FRONTEND_PID 2>/dev/null
    rm /tmp/rolecraft-frontend.pid
    echo "   ✅ 前端已停止"
else
    echo "ℹ️  前端未运行"
fi

# 确保所有相关进程都被停止
pkill -f "go run cmd/server" 2>/dev/null
pkill -f "npm run dev" 2>/dev/null

echo ""
echo "✅ 所有服务已停止"
echo "========================================="
