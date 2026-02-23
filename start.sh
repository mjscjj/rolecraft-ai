#!/bin/bash

# RoleCraft AI 一键启动脚本

set -e

echo "🚀 RoleCraft AI 启动脚本"
echo "========================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: Docker 未安装${NC}"
    exit 1
fi

# 检查 Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}错误: Docker Compose 未安装${NC}"
    exit 1
fi

# 检查 Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: Go 未安装${NC}"
    exit 1
fi

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo -e "${YELLOW}警告: Node.js 未安装，前端服务将跳过${NC}"
    SKIP_FRONTEND=true
fi

echo ""
echo "📦 步骤 1: 启动基础设施 (PostgreSQL, Redis, MinIO)..."
docker-compose up -d postgres redis minio

echo ""
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务健康
echo ""
echo "🔍 检查服务状态..."
if docker-compose ps | grep -q "postgres.*Up"; then
    echo -e "${GREEN}✓ PostgreSQL 运行中${NC}"
else
    echo -e "${RED}✗ PostgreSQL 未启动${NC}"
fi

if docker-compose ps | grep -q "redis.*Up"; then
    echo -e "${GREEN}✓ Redis 运行中${NC}"
else
    echo -e "${RED}✗ Redis 未启动${NC}"
fi

echo ""
echo "📊 步骤 2: 初始化数据库..."
cd backend

# 检查 .env 文件
if [ ! -f .env ]; then
    echo -e "${YELLOW}创建 .env 文件...${NC}"
    cp .env.example .env
    echo -e "${YELLOW}请编辑 backend/.env 文件，填入你的 OpenAI API Key${NC}"
fi

# 运行迁移
echo "运行数据库迁移..."
go run cmd/migrate/main.go up

# 填充数据
echo "填充初始数据..."
go run cmd/migrate/main.go seed

echo ""
echo "🔧 步骤 3: 启动后端服务..."
go run cmd/server/main.go &
BACKEND_PID=$!
echo -e "${GREEN}后端服务已启动 (PID: $BACKEND_PID)${NC}"

# 等待后端启动
sleep 5

# 检查后端
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✓ 后端 API 运行中${NC}"
else
    echo -e "${RED}✗ 后端 API 未启动${NC}"
fi

echo ""
echo "🎨 步骤 4: 启动前端服务..."
if [ "$SKIP_FRONTEND" != "true" ]; then
    cd ../frontend
    
    # 安装依赖
    if [ ! -d "node_modules" ]; then
        echo "安装前端依赖..."
        pnpm install || npm install
    fi
    
    # 启动前端
    pnpm dev &
    FRONTEND_PID=$!
    echo -e "${GREEN}前端服务已启动 (PID: $FRONTEND_PID)${NC}"
fi

echo ""
echo "========================"
echo -e "${GREEN}✅ RoleCraft AI 启动完成！${NC}"
echo ""
echo "访问地址:"
echo "  前端:   http://localhost:3000"
echo "  后端:   http://localhost:8080"
echo "  API文档: http://localhost:8080/swagger"
echo "  MinIO:  http://localhost:9001 (minioadmin/minioadmin123)"
echo ""
echo "按 Ctrl+C 停止所有服务"

# 保存 PID
echo $BACKEND_PID > /tmp/rolecraft-backend.pid
[ ! -z "$FRONTEND_PID" ] && echo $FRONTEND_PID > /tmp/rolecraft-frontend.pid

# 等待
wait
