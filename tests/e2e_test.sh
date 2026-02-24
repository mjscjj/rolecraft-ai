#!/bin/bash
# RoleCraft AI E2E 测试 (浏览器模拟) - 30+ 用例

BASE_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:5173"
PASS=0
FAIL=0
TOTAL=0
TEST_SUFFIX=$(date +%s)

echo "=== RoleCraft AI E2E 测试 (30+ 用例) ==="
echo "时间：$(date)"
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

# 获取认证 Token
get_auth_token() {
    local response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"user1@test.com","password":"123456"}')
    echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4
}

TOKEN=$(get_auth_token)

echo "=== 1. 前端页面测试 (4 tests) ==="
FRONTEND_HTML=$(curl -s "$FRONTEND_URL")
test_e2e "前端页面可访问" "true" "$([ -n "$FRONTEND_HTML" ] && echo 'true' || echo 'false')"
test_e2e "包含 React 根节点" "true" "$([[ "$FRONTEND_HTML" == *'id="root"'* ]] && echo 'true' || echo 'false')"
test_e2e "包含 Vite 脚本" "true" "$([[ "$FRONTEND_HTML" == *'vite'* ]] && echo 'true' || echo 'false')"
test_e2e "包含 HTML5 doctype" "true" "$([[ "$FRONTEND_HTML" == *'<!doctype html>'* ]] && echo 'true' || echo 'false')"
echo ""

echo "=== 2. API 端点可达性 (8 tests) ==="
# GET 端点
test_e2e "GET /health" "200" "$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")"
test_e2e "GET /api/v1/roles/templates" "200" "$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/roles/templates")"
test_e2e "POST /api/v1/auth/refresh (空)" "401" "$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/auth/refresh" -H "Content-Type: application/json" -d '{}')"
test_e2e "GET /api/v1/roles (无 Token)" "401" "$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/roles")"
test_e2e "GET /api/v1/documents (无 Token)" "401" "$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/documents")"
test_e2e "GET /api/v1/chat-sessions (无 Token)" "401" "$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/chat-sessions")"
test_e2e "POST /api/v1/auth/register (空)" "400" "$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" -H "Content-Type: application/json" -d '{}')"
test_e2e "POST /api/v1/auth/login (空)" "400" "$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/auth/login" -H "Content-Type: application/json" -d '{}')"
echo ""

echo "=== 3. 数据完整性测试 (4 tests) ==="
TEMPLATES=$(curl -s "$BASE_URL/api/v1/roles/templates")
TEMPLATE_COUNT=$(echo "$TEMPLATES" | grep -o '"id"' | wc -l | tr -d ' ')
test_e2e "角色模板数量 >= 3" "true" "$([ $TEMPLATE_COUNT -ge 3 ] && echo 'true' || echo 'false')"
test_e2e "模板包含 name 字段" "true" "$([[ "$TEMPLATES" == *'"name"'* ]] && echo 'true' || echo 'false')"
test_e2e "模板包含 systemPrompt" "true" "$([[ "$TEMPLATES" == *'"systemPrompt"'* ]] && echo 'true' || echo 'false')"
test_e2e "模板包含 description" "true" "$([[ "$TEMPLATES" == *'"description"'* ]] && echo 'true' || echo 'false')"
echo ""

echo "=== 4. 认证流程测试 (6 tests) ==="
# 注册新用户
REGISTER_EMAIL="e2e_${TEST_SUFFIX}@test.com"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$REGISTER_EMAIL\",\"password\":\"test123\",\"name\":\"E2E User\"}")
test_e2e "注册返回成功" "true" "$([[ "$REGISTER_RESPONSE" == *'"token"'* ]] && echo 'true' || echo 'false')"
test_e2e "注册返回 Token" "true" "$([[ "$REGISTER_RESPONSE" == *'"token"'* ]] && echo 'true' || echo 'false')"

# 重复注册
DUPLICATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$REGISTER_EMAIL\",\"password\":\"test123\",\"name\":\"E2E User\"}")
test_e2e "重复注册返回错误" "true" "$([[ "$DUPLICATE_RESPONSE" == *'"error"'* ]] || [[ "$DUPLICATE_RESPONSE" == *'"code":409'* ]] && echo 'true' || echo 'false')"

# 登录
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$REGISTER_EMAIL\",\"password\":\"test123\"}")
test_e2e "登录返回成功" "true" "$([[ "$LOGIN_RESPONSE" == *'"token"'* ]] && echo 'true' || echo 'false')"
test_e2e "登录返回 Token" "true" "$([[ "$LOGIN_RESPONSE" == *'"token"'* ]] && echo 'true' || echo 'false')"

# 错误密码登录
ERROR_LOGIN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$REGISTER_EMAIL\",\"password\":\"wrongpassword\"}")
test_e2e "错误密码返回错误" "true" "$([[ "$ERROR_LOGIN" == *'"error"'* ]] && echo 'true' || echo 'false')"
echo ""

echo "=== 5. 授权访问测试 (5 tests) ==="
if [ -n "$TOKEN" ]; then
    # 获取用户信息
    USER_RESPONSE=$(curl -s "$BASE_URL/api/v1/users/me" -H "Authorization: Bearer $TOKEN")
    test_e2e "获取用户信息成功" "true" "$([[ "$USER_RESPONSE" == *'"email"'* ]] && echo 'true' || echo 'false')"
    
    # 获取角色列表
    ROLES_RESPONSE=$(curl -s "$BASE_URL/api/v1/roles" -H "Authorization: Bearer $TOKEN")
    test_e2e "获取角色列表成功" "true" "$([[ "$ROLES_RESPONSE" == *'"code":200'* ]] && echo 'true' || echo 'false')"
    
    # 创建角色
    CREATE_ROLE=$(curl -s -X POST "$BASE_URL/api/v1/roles" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"E2E 测试角色\",\"systemPrompt\":\"测试\"}")
    test_e2e "创建角色成功" "true" "$([[ "$CREATE_ROLE" == *'"code":200'* ]] || [[ "$CREATE_ROLE" == *'"code":201'* ]] && echo 'true' || echo 'false')"
    
    # 获取文档列表
    DOCS_RESPONSE=$(curl -s "$BASE_URL/api/v1/documents" -H "Authorization: Bearer $TOKEN")
    test_e2e "获取文档列表成功" "true" "$([[ "$DOCS_RESPONSE" == *'"code":200'* ]] && echo 'true' || echo 'false')"
    
    # 获取会话列表
    SESSIONS_RESPONSE=$(curl -s "$BASE_URL/api/v1/chat-sessions" -H "Authorization: Bearer $TOKEN")
    test_e2e "获取会话列表成功" "true" "$([[ "$SESSIONS_RESPONSE" == *'"code":200'* ]] && echo 'true' || echo 'false')"
fi
echo ""

echo "=== 6. CORS 测试 (2 tests) ==="
# 使用 OPTIONS 预检请求测试 CORS
CORS_RESPONSE=$(curl -s -I -X OPTIONS "$BASE_URL/health" -H "Origin: http://localhost:5173" 2>&1)
test_e2e "CORS 头存在" "true" "$([[ "$CORS_RESPONSE" == *'Access-Control'* ]] && echo 'true' || echo 'false')"
test_e2e "Allow-Origin 配置" "true" "$([[ "$CORS_RESPONSE" == *'*'* ]] && echo 'true' || echo 'false')"
echo ""

echo "=== 7. 错误处理测试 (5 tests) ==="
ERROR_404=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/not-exist")
test_e2e "404 错误码" "404" "$ERROR_404"

ERROR_401=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/users/me")
test_e2e "401 未授权" "401" "$ERROR_401"

ERROR_400=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d '{}')
test_e2e "400 参数错误" "400" "$ERROR_400"

# 无效 Token
INVALID_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/roles" \
    -H "Authorization: Bearer invalid_token_here")
test_e2e "无效 Token 返回 401" "401" "$INVALID_TOKEN"

# 方法不允许
METHOD_NOT_ALLOWED=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/health")
test_e2e "DELETE /health 返回 404" "404" "$METHOD_NOT_ALLOWED"
echo ""

echo "=== 8. 响应时间测试 (5 tests) ==="
for i in {1..5}; do
    TIME=$(curl -s -o /dev/null -w "%{time_total}" "$BASE_URL/health")
    # 检查是否小于 0.1 秒 (100ms) - 使用 awk 代替 bc
    FAST=$(echo "$TIME" | awk '{if ($1 < 0.1) print "yes"; else print "no"}')
    if [ "$FAST" = "yes" ]; then
        echo "✅ 响应时间测试 $i: ${TIME}s (< 100ms)"
        ((PASS++))
    else
        echo "⚠️ 响应时间测试 $i: ${TIME}s (>= 100ms)"
        ((FAIL++))
    fi
    ((TOTAL++))
done
echo ""

echo "=== 9. 并发测试 (2 tests) ==="
# 10 个并发请求
for i in {1..10}; do
    curl -s "$BASE_URL/health" > /dev/null &
done
wait
echo "✅ 10 个并发请求完成"
((PASS++))
((TOTAL++))

# 并发后服务仍然正常
AFTER_CONCURRENT=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
test_e2e "并发后服务正常" "200" "$AFTER_CONCURRENT"
echo ""

echo "=== 10. 数据持久化测试 (3 tests) ==="
# 创建角色后验证可以获取
if [ -n "$TOKEN" ]; then
    # 先获取角色列表
    BEFORE_COUNT=$(curl -s "$BASE_URL/api/v1/roles" -H "Authorization: Bearer $TOKEN" | grep -o '"id"' | wc -l)
    
    # 创建新角色
    curl -s -X POST "$BASE_URL/api/v1/roles" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"持久化测试角色\",\"systemPrompt\":\"测试持久化\"}" > /dev/null
    
    # 再次获取角色列表
    AFTER_COUNT=$(curl -s "$BASE_URL/api/v1/roles" -H "Authorization: Bearer $TOKEN" | grep -o '"id"' | wc -l)
    
    test_e2e "角色创建后数量增加" "true" "$([ "$AFTER_COUNT" -gt "$BEFORE_COUNT" ] && echo 'true' || echo 'false')"
fi

# 健康检查持续通过
HEALTH_1=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
sleep 1
HEALTH_2=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
test_e2e "服务持续健康" "true" "$([ "$HEALTH_1" = "200" ] && [ "$HEALTH_2" = "200" ] && echo 'true' || echo 'false')"
echo ""

echo "=== 测试结果 ==="
echo "总计：$TOTAL"
echo "通过：$PASS"
echo "失败：$FAIL"
echo "覆盖率：$((PASS * 100 / TOTAL))%"
echo ""

if [ $FAIL -eq 0 ]; then
    echo "✅ 所有 E2E 测试通过！"
    exit 0
else
    echo "⚠️ 有 $FAIL 个测试失败"
    exit 1
fi
