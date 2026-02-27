#!/bin/bash

echo "========================================="
echo "  🚀 RoleCraft AI - 一键启动"
echo "========================================="
echo ""

# 检查 Go
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go"
    exit 1
fi

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装，请先安装 Node.js"
    exit 1
fi

echo "✅ 环境检查通过"
echo ""

# 停止旧服务
echo "🛑 停止旧服务..."
pkill -f "go run cmd/server" 2>/dev/null
pkill -f "npm run dev" 2>/dev/null
sleep 2
echo "✅ 已停止旧服务"
echo ""

# 启动后端
echo "📦 启动后端服务..."
cd backend
nohup go run cmd/server/main.go > logs/server.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > /tmp/rolecraft-backend.pid
echo "   ✅ 后端已启动 (PID: $BACKEND_PID)"

# 等待后端启动
echo "   ⏳ 等待后端就绪..."
for i in {1..10}; do
    if curl -s http://localhost:8080/health | grep -q "ok"; then
        echo "   ✅ 后端就绪"
        break
    fi
    sleep 1
done

if ! curl -s http://localhost:8080/health | grep -q "ok"; then
    echo "   ❌ 后端启动失败"
    exit 1
fi

echo ""

# 启动前端
echo "🎨 启动前端服务..."
cd ../frontend
nohup npm run dev > logs/frontend.log 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > /tmp/rolecraft-frontend.pid
echo "   ✅ 前端已启动 (PID: $FRONTEND_PID)"

# 等待前端启动
echo "   ⏳ 等待前端就绪..."
for i in {1..10}; do
    if curl -s http://localhost:5173 | grep -q "html"; then
        echo "   ✅ 前端就绪"
        break
    fi
    sleep 1
done

if ! curl -s http://localhost:5173 | grep -q "html"; then
    echo "   ❌ 前端启动失败"
    exit 1
fi

echo ""
echo "========================================="
echo "  ✅ RoleCraft AI 启动成功！"
echo "========================================="
echo ""
echo "  🌐 访问地址："
echo "     http://localhost:5173"
echo ""
echo "  📊 服务状态："
echo "     后端：http://localhost:8080/health"
echo "     前端：http://localhost:5173"
echo ""
echo "  🛑 停止服务："
echo "     ./stop.sh"
echo ""
echo "========================================="
echo ""
