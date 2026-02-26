#!/bin/bash

echo "========================================="
echo "  RoleCraft AI E2E 测试"
echo "========================================="
echo ""

cd "$(dirname "$0")/.."

# 1. 检查依赖
echo "1️⃣  检查 Playwright 安装..."
if ! command -v npx &> /dev/null; then
    echo "   ❌ Node.js/npm 未安装"
    exit 1
fi

if [ ! -d "node_modules/@playwright" ]; then
    echo "   安装 Playwright..."
    npm install -D @playwright/test
    npx playwright install chromium
fi
echo "   ✅ Playwright 已安装"
echo ""

# 2. 检查服务
echo "2️⃣  检查服务状态..."

# 检查后端
if curl -s http://localhost:8080/health | grep -q "ok"; then
    echo "   ✅ 后端服务运行中 (8080)"
else
    echo "   ❌ 后端服务未运行"
    echo "   启动：cd backend && ./bin/server"
    exit 1
fi

# 检查前端
if curl -s http://localhost:5173 | grep -q "html"; then
    echo "   ✅ 前端服务运行中 (5173)"
else
    echo "   ❌ 前端服务未运行"
    echo "   启动：cd frontend && npm run dev"
    exit 1
fi
echo ""

# 3. 运行测试
echo "3️⃣  运行 E2E 测试..."
echo ""

npx playwright test --reporter=list

# 4. 显示结果
echo ""
echo "========================================="
if [ $? -eq 0 ]; then
    echo "  ✅ 所有测试通过！"
else
    echo "  ⚠️  部分测试失败"
fi
echo "========================================="
echo ""

# 5. 生成 HTML 报告
echo "📊 生成 HTML 报告..."
npx playwright show-report --host 0.0.0.0 &
echo "   报告地址：http://localhost:9323"
echo ""
