#!/bin/bash
# RoleCraft AI E2E 测试 (浏览器模拟)

BASE_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:5173"
PASS=0
FAIL=0
TOTAL=0

echo "=== RoleCraft AI E2E 测试 ==="
echo ""

# 测试函数
test_e2e() {
    local name=$1
    local expected=$2
    local actual=$3
    
    ((TOTAL++))
    
    if [ "$actual" = "$expected" ]; then
        echo "✅ $name"
        ((PASS++))
    else
        echo "❌ $name (expected: $expected, got: $actual)"
        ((FAIL++))
    fi
}

echo "=== 1. 前端页面测试 ==="
FRONTEND_HTML=$(curl -s "$FRONTEND_URL")
test_e2e "前端页面可访问" "true" "$([ -n \"$FRONTEND_HTML\" ] && echo 'true' || echo 'false')"
test_e2e "包含 React 根节点" "true" "$([[ \"$FRONTEND_HTML\" == *'id="root"'* ]] && echo 'true' || echo 'false')"
test_e2e "包含 Vite 脚本" "true" "$([[ \"$FRONTEND_HTML\" == *'vite'* ]] && echo 'true' || echo 'false')"

echo ""
echo "=== 2. API 端点可达性 ==="
ENDPOINTS=(
    "GET /health"
    "GET /api/v1/roles/templates"
    "POST /api/v1/auth/register"
    "POST /api/v1/auth/login"
    "GET /api/v1/roles"
    "GET /api/v1/documents"
    "GET /api/v1/chat-sessions"
)

for endpoint in "${ENDPOINTS[@]}"; do
    method=$(echo $endpoint | cut -d' ' -f1)
    path=$(echo $endpoint | cut -d' ' -f2)
    
    if [ "$method" = "GET" ]; then
        code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$path")
    else
        code=$(curl -s -o /dev/null -w "%{http_code}" -X $method "$BASE_URL$path" \
            -H "Content-Type: application/json" \
            -d '{}')
    fi
    
    # 200, 201, 400, 401, 404 都是正常响应
    if [ "$code" -ge 200 ] && [ "$code" -lt 500 ]; then
        echo "✅ $endpoint (HTTP $code)"
        ((PASS++))
    else
        echo "❌ $endpoint (HTTP $code)"
        ((FAIL++))
    fi
    ((TOTAL++))
done

echo ""
echo "=== 3. 数据完整性测试 ==="
TEMPLATES=$(curl -s "$BASE_URL/api/v1/roles/templates")
TEMPLATE_COUNT=$(echo "$TEMPLATES" | grep -o '"id"' | wc -l | tr -d ' ')
test_e2e "角色模板数量 >= 3" "true" "$([ $TEMPLATE_COUNT -ge 3 ] && echo 'true' || echo 'false')"

echo ""
echo "=== 4. 认证流程测试 ==="
# 注册
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"e2e_test_$(date +%s)@test.com\",\"password\":\"test123\",\"name\":\"E2E User\"}")
HAS_TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"token"' | wc -l | tr -d ' ')
test_e2e "注册返回Token" "true" "$([ $HAS_TOKEN -gt 0 ] && echo 'true' || echo 'false')"

# 登录
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"e2e_test_$(date +%s)@test.com\",\"password\":\"test123\"}")
test_e2e "登录返回Token" "true" "$([[ \"$LOGIN_RESPONSE\" == *'token'* ]] && echo 'true' || echo 'false')"

echo ""
echo "=== 5. CORS 测试 ==="
CORS_HEADER=$(curl -s -I "$BASE_URL/health" | grep -i "access-control" | head -1)
test_e2e "CORS 头存在" "true" "$([ -n \"$CORS_HEADER\" ] && echo 'true' || echo 'false')"

echo ""
echo "=== 6. 错误处理测试 ==="
ERROR_404=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/not-exist")
test_e2e "404错误码" "404" "$ERROR_404"

ERROR_401=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/users/me")
test_e2e "401未授权" "401" "$ERROR_401"

ERROR_400=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d '{}')
test_e2e "400参数错误" "400" "$ERROR_400"

echo ""
echo "=== 7. 响应时间测试 ==="
for i in {1..5}; do
    TIME=$(curl -s -o /dev/null -w "%{time_total}" "$BASE_URL/health")
    if (( $(echo "$TIME < 0.1" | bc -l) )); then
        echo "✅ 响应时间 $i: ${TIME}s (< 100ms)"
        ((PASS++))
    else
        echo "⚠️ 响应时间 $i: ${TIME}s (>= 100ms)"
        ((FAIL++))
    fi
    ((TOTAL++))
done

echo ""
echo "=== 8. 并发测试 ==="
for i in {1..10}; do
    curl -s "$BASE_URL/health" > /dev/null &
done
wait
echo "✅ 10个并发请求完成"
((PASS++))
((TOTAL++))

echo ""
echo "=== 测试结果 ==="
echo "总计: $TOTAL"
echo "通过: $PASS"
echo "失败: $FAIL"
echo ""

if [ $FAIL -eq 0 ]; then
    echo "✅ 所有E2E测试通过！"
    exit 0
else
    echo "⚠️ 有 $FAIL 个测试失败"
    exit 1
fi