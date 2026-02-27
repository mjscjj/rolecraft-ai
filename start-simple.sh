#!/bin/bash

echo "🚀 Starting RoleCraft AI (Simple Mode)..."
echo ""

# 检查 Go
if ! command -v go &> /dev/null; then
    echo "❌ Go not found. Please install Go 1.21+"
    exit 1
fi

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js not found. Please install Node.js 18+"
    exit 1
fi

echo "✅ Dependencies check passed"
echo ""

# 启动后端
echo "📦 Starting backend..."
cd backend
go run cmd/server/main.go &
BACKEND_PID=$!
echo "   ✅ Backend started (PID: $BACKEND_PID)"

# 等待后端启动
echo "   ⏳ Waiting for backend..."
sleep 3

# 检查后端健康
if curl -s "http://localhost:8080/health" | grep -q "ok"; then
    echo "   ✅ Backend is healthy"
else
    echo "   ❌ Backend failed to start"
    kill $BACKEND_PID
    exit 1
fi

echo ""

# 启动前端
echo "🎨 Starting frontend..."
cd ../frontend
npm run dev &
FRONTEND_PID=$!
echo "   ✅ Frontend started (PID: $FRONTEND_PID)"

# 等待前端启动
echo "   ⏳ Waiting for frontend..."
sleep 2

# 检查前端
if curl -s "http://localhost:5173" | grep -q "html"; then
    echo "   ✅ Frontend is healthy"
else
    echo "   ❌ Frontend failed to start"
    kill $FRONTEND_PID
    exit 1
fi

echo ""
echo "========================================="
echo "  ✅ RoleCraft AI is running!"
echo "========================================="
echo ""
echo "  🌐 Frontend:  http://localhost:5173"
echo "  🔧 Backend:   http://localhost:8080"
echo "  📊 Health:    http://localhost:8080/health"
echo ""
echo "  Press Ctrl+C to stop all services"
echo ""

# 保存 PID 到文件
echo $BACKEND_PID > /tmp/rolecraft-backend.pid
echo $FRONTEND_PID > /tmp/rolecraft-frontend.pid

# 等待中断信号
trap "echo ''; echo '🛑 Stopping services...'; kill $BACKEND_PID $FRONTEND_PID; rm -f /tmp/rolecraft-*.pid; echo '✅ Stopped'; exit 0" EXIT
wait
