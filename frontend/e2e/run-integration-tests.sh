#!/bin/bash

# RoleCraft AI - E2E Integration Tests Runner
# 测试完整用户流程：注册 → 创建角色 → 上传文档 → 对话

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKEND_ROOT="$(dirname "$PROJECT_ROOT")"

echo "========================================="
echo "  RoleCraft AI E2E Integration Tests"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 清理函数
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"
    
    # 杀死后台进程
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
        echo "Backend process stopped (PID: $BACKEND_PID)"
    fi
    
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
        echo "Frontend process stopped (PID: $FRONTEND_PID)"
    fi
    
    # 清理临时文件
    rm -rf /tmp/uploads 2>/dev/null || true
}

# 设置陷阱，确保退出时清理
trap cleanup EXIT

# 1. 检查依赖
echo -e "${YELLOW}Step 1: Checking dependencies...${NC}"

if ! command -v pnpm &> /dev/null; then
    echo -e "${RED}Error: pnpm is not installed${NC}"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Dependencies OK${NC}"
echo ""

# 2. 安装依赖
echo -e "${YELLOW}Step 2: Installing dependencies...${NC}"

cd "$BACKEND_ROOT/backend"
go mod download
echo "✓ Backend dependencies installed"

cd "$PROJECT_ROOT"
pnpm install
echo "✓ Frontend dependencies installed"
echo ""

# 3. 安装 Playwright 浏览器
echo -e "${YELLOW}Step 3: Installing Playwright browsers...${NC}"
pnpm exec playwright install chromium
echo -e "${GREEN}✓ Playwright browsers installed${NC}"
echo ""

# 4. 构建项目
echo -e "${YELLOW}Step 4: Building projects...${NC}"

cd "$BACKEND_ROOT/backend"
go build -o bin/server cmd/server/main.go
echo "✓ Backend built"

cd "$PROJECT_ROOT"
pnpm build
echo "✓ Frontend built"
echo ""

# 5. 启动后端
echo -e "${YELLOW}Step 5: Starting backend server...${NC}"

export DATABASE_URL="postgres://test:test@localhost:5432/rolecraft_e2e?sslmode=disable"
export JWT_SECRET="e2e-test-jwt-secret-key-for-testing-only"
export UPLOAD_DIR="/tmp/uploads"
export ANYTHINGLLM_BASE_URL="http://localhost:3001/api/v1"
export ANYTHINGLLM_API_KEY="test-api-key"
export ANYTHINGLLM_WORKSPACE="e2e_test_workspace"

mkdir -p /tmp/uploads

cd "$BACKEND_ROOT/backend"
./bin/server > /tmp/backend.log 2>&1 &
BACKEND_PID=$!
echo "Backend started (PID: $BACKEND_PID)"

# 等待后端启动
echo "Waiting for backend to be ready..."
for i in {1..15}; do
    if curl -s http://localhost:8080/api/v1/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Backend is ready${NC}"
        break
    fi
    
    if [ $i -eq 15 ]; then
        echo -e "${RED}✗ Backend failed to start${NC}"
        echo "Backend logs:"
        cat /tmp/backend.log
        exit 1
    fi
    
    echo "  Waiting... ($i/15)"
    sleep 2
done
echo ""

# 6. 启动前端
echo -e "${YELLOW}Step 6: Starting frontend server...${NC}"

cd "$PROJECT_ROOT"
pnpm serve dist > /tmp/frontend.log 2>&1 &
FRONTEND_PID=$!
echo "Frontend started (PID: $FRONTEND_PID)"

# 等待前端启动
echo "Waiting for frontend to be ready..."
for i in {1..15}; do
    if curl -s http://localhost:5173 > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Frontend is ready${NC}"
        break
    fi
    
    if [ $i -eq 15 ]; then
        echo -e "${RED}✗ Frontend failed to start${NC}"
        echo "Frontend logs:"
        cat /tmp/frontend.log
        exit 1
    fi
    
    echo "  Waiting... ($i/15)"
    sleep 2
done
echo ""

# 7. 运行 E2E 测试
echo -e "${YELLOW}Step 7: Running E2E Integration Tests...${NC}"
echo "========================================="

cd "$PROJECT_ROOT"
pnpm exec playwright test e2e/integration.spec.ts --reporter=list

TEST_EXIT_CODE=$?

echo "========================================="

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
else
    echo -e "${RED}✗ Some tests failed${NC}"
fi

# 8. 生成报告
echo ""
echo -e "${YELLOW}Step 8: Generating HTML report...${NC}"
pnpm exec playwright show-report --host=0.0.0.0 &
echo "HTML report available at: http://localhost:9323"
echo ""

# 保持脚本运行，让用户可以查看报告
echo -e "${YELLOW}Press Ctrl+C to stop the report server and cleanup${NC}"
wait

exit $TEST_EXIT_CODE
