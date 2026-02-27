#!/bin/bash

# RoleCraft AI 数据分析平台 - 快速启动脚本

echo "🚀 RoleCraft AI 数据分析平台 - 快速启动"
echo "========================================"
echo ""

# 检查后端编译
echo "📦 检查后端编译..."
if [ ! -f "bin/server" ]; then
    echo "编译后端..."
    cd backend
    go build -o ../bin/server ./cmd/server/main.go
    if [ $? -eq 0 ]; then
        echo "✅ 后端编译成功"
    else
        echo "❌ 后端编译失败"
        exit 1
    fi
    cd ..
else
    echo "✅ 后端已编译"
fi

# 检查前端依赖
echo ""
echo "📦 检查前端依赖..."
cd frontend
if [ ! -d "node_modules" ]; then
    echo "安装前端依赖..."
    npm install
    if [ $? -eq 0 ]; then
        echo "✅ 前端依赖安装完成"
    else
        echo "❌ 前端依赖安装失败"
        exit 1
    fi
else
    echo "✅ 前端依赖已安装"
fi

# 启动后端
echo ""
echo "🚀 启动后端服务..."
cd ..
./bin/server &
BACKEND_PID=$!
echo "✅ 后端服务已启动 (PID: $BACKEND_PID)"

# 等待后端启动
echo "等待后端启动..."
sleep 3

# 启动前端
echo ""
echo "🚀 启动前端开发服务器..."
cd frontend
npm run dev &
FRONTEND_PID=$!
echo "✅ 前端服务已启动 (PID: $FRONTEND_PID)"

echo ""
echo "========================================"
echo "🎉 服务启动成功!"
echo ""
echo "📊 数据分析平台地址:"
echo "   http://localhost:5173/analytics"
echo ""
echo "🔧 后端 API 地址:"
echo "   http://localhost:8080/api/v1/analytics"
echo ""
echo "📖 API 文档:"
echo "   http://localhost:8080/swagger/index.html"
echo ""
echo "按 Ctrl+C 停止所有服务"
echo "========================================"

# 等待用户中断
wait
