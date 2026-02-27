#!/bin/bash

echo "🛑 Stopping RoleCraft AI..."

# 读取 PID 文件
if [ -f /tmp/rolecraft-backend.pid ]; then
    BACKEND_PID=$(cat /tmp/rolecraft-backend.pid)
    echo "   Stopping backend (PID: $BACKEND_PID)..."
    kill $BACKEND_PID 2>/dev/null
    rm /tmp/rolecraft-backend.pid
fi

if [ -f /tmp/rolecraft-frontend.pid ]; then
    FRONTEND_PID=$(cat /tmp/rolecraft-frontend.pid)
    echo "   Stopping frontend (PID: $FRONTEND_PID)..."
    kill $FRONTEND_PID 2>/dev/null
    rm /tmp/rolecraft-frontend.pid
fi

# 或者通过进程名停止
pkill -f "go run cmd/server" 2>/dev/null
pkill -f "npm run dev" 2>/dev/null

echo "✅ Stopped"
