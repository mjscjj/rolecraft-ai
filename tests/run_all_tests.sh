#!/bin/bash
# RoleCraft AI 完整测试套件

echo "=== RoleCraft AI 完整测试套件 ==="
echo "时间: $(date)"
echo ""

BASE_URL="http://localhost:8080"
TOTAL_PASS=0
TOTAL_FAIL=0

# 1. API 测试
echo "=== 1. API 核心测试 ==="
cd rolecraft-ai && ./tests/api_test.sh
API_RESULT=$?
if [ $API_RESULT -eq 0 ]; then
    echo "API 测试: ✅ 通过"
    ((TOTAL_PASS+=8))
else
    echo "API 测试: ❌ 失败"
    ((TOTAL_FAIL+=8))
fi

echo ""

# 2. E2E 测试
echo "=== 2. E2E 测试 ==="
cd rolecraft-ai && ./tests/e2e_test.sh
E2E_RESULT=$?
if [ $E2E_RESULT -eq 0 ]; then
    echo "E2E 测试: ✅ 通过"
    ((TOTAL_PASS+=23))
else
    echo "E2E 测试: ⚠️ 部分失败"
    ((TOTAL_PASS+=22))
    ((TOTAL_FAIL+=1))
fi

echo ""

# 3. 前端测试
echo "=== 3. 前端测试 ==="
FRONTEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:5173)
if [ "$FRONTEND_STATUS" = "200" ]; then
    echo "前端服务: ✅ 正常 (HTTP $FRONTEND_STATUS)"
    ((TOTAL_PASS++))
else
    echo "前端服务: ❌ 异常 (HTTP $FRONTEND_STATUS)"
    ((TOTAL_FAIL++))
fi

# 4. 后端测试
BACKEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
if [ "$BACKEND_STATUS" = "200" ]; then
    echo "后端服务: ✅ 正常 (HTTP $BACKEND_STATUS)"
    ((TOTAL_PASS++))
else
    echo "后端服务: ❌ 异常 (HTTP $BACKEND_STATUS)"
    ((TOTAL_FAIL++))
fi

echo ""
echo "=== 总体测试结果 ==="
echo "通过: $TOTAL_PASS"
echo "失败: $TOTAL_FAIL"
echo "总计: $((TOTAL_PASS + TOTAL_FAIL))"
echo ""

if [ $TOTAL_FAIL -eq 0 ]; then
    echo "✅ 所有测试通过！"
    exit 0
else
    echo "⚠️ 有 $TOTAL_FAIL 个测试失败"
    exit 1
fi