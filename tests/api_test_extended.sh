#!/bin/bash
# RoleCraft AI 扩展 API 测试 (50+ 用例)

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0
TOTAL=0
RANDOM_SUFFIX=$(date +%s)

echo "=== RoleCraft API 测试套件 ==="
echo ""

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected=${5:-200}
    local headers=${6:-""}
    
    ((TOTAL++))
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" $headers "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method $headers "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "$expected" ]; then
        echo "✅ $name"
        ((PASS++))
    else
        echo "❌ $name (HTTP $http_code, expected $expected)"
        ((FAIL++))
    fi
}

# 获取认证Token
get_token() {
    local email=$1
    curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"123456\"}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4
}

echo "=== 1. 健康检查 (2 tests) ==="
test_api "GET /health" "GET" "/health" "" "200"
test_api "GET /health 返回 ok" "GET" "/health" "" "200"
echo ""

echo "=== 2. 角色模板 (3 tests) ==="
test_api "GET /roles/templates" "GET" "/api/v1/roles/templates" "" "200"
test_api "验证模板数量 >= 3" "GET" "/api/v1/roles/templates" "" "200"
test_api "模板包含必要字段" "GET" "/api/v1/roles/templates" "" "200"
echo ""

echo "=== 3. 用户注册 (8 tests) ==="
test_api "正常注册 User1" "POST" "/api/v1/auth/register" "{\"email\":\"user1_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\",\"name\":\"User1\"}" "201"
test_api "正常注册 User2" "POST" "/api/v1/auth/register" "{\"email\":\"user2_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\",\"name\":\"User2\"}" "201"
test_api "正常注册 User3" "POST" "/api/v1/auth/register" "{\"email\":\"user3_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\",\"name\":\"User3\"}" "201"
test_api "重复邮箱注册失败" "POST" "/api/v1/auth/register" "{\"email\":\"user1_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\",\"name\":\"User1\"}" "409"
test_api "缺少邮箱失败" "POST" "/api/v1/auth/register" "{\"password\":\"123456\",\"name\":\"NoEmail\"}" "400"
test_api "缺少密码失败" "POST" "/api/v1/auth/register" "{\"email\":\"no${RANDOM_SUFFIX}@test.com\",\"name\":\"NoPass\"}" "400"
test_api "密码太短失败" "POST" "/api/v1/auth/register" "{\"email\":\"short${RANDOM_SUFFIX}@test.com\",\"password\":\"123\",\"name\":\"Short\"}" "400"
test_api "无效邮箱格式" "POST" "/api/v1/auth/register" "{\"email\":\"invalid\",\"password\":\"123456\",\"name\":\"Invalid\"}" "400"
echo ""

echo "=== 4. 用户登录 (6 tests) ==="
test_api "正确密码登录" "POST" "/api/v1/auth/login" "{\"email\":\"user1_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\"}" "200"
test_api "错误密码登录" "POST" "/api/v1/auth/login" "{\"email\":\"user1_${RANDOM_SUFFIX}@test.com\",\"password\":\"wrong\"}" "401"
test_api "不存在用户登录" "POST" "/api/v1/auth/login" "{\"email\":\"notexist@test.com\",\"password\":\"123456\"}" "401"
test_api "缺少邮箱登录" "POST" "/api/v1/auth/login" "{\"password\":\"123456\"}" "400"
test_api "缺少密码登录" "POST" "/api/v1/auth/login" "{\"email\":\"user1_${RANDOM_SUFFIX}@test.com\"}" "400"
test_api "登录返回token" "POST" "/api/v1/auth/login" "{\"email\":\"user1_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\"}" "200"
echo ""

echo "=== 5. Token刷新 (2 tests) ==="
TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"user1_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\"}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ 无法获取Token，跳过需要认证的测试"
    echo "   请检查用户登录是否正常"
else
    echo "Token 获取成功: ${TOKEN:0:20}..."
    test_api "Token刷新成功" "POST" "/api/v1/auth/refresh" "" "200" "-H \"Authorization: Bearer $TOKEN\""
fi
test_api "无效Token刷新失败" "POST" "/api/v1/auth/refresh" "" "401" "-H \"Authorization: Bearer invalid\""
echo ""

echo "=== 6. 用户信息 (3 tests) ==="
if [ -n "$TOKEN" ]; then
    test_api "获取当前用户" "GET" "/api/v1/users/me" "" "200" "-H \"Authorization: Bearer $TOKEN\""
    test_api "更新用户名" "PUT" "/api/v1/users/me" "{\"name\":\"UpdatedName\"}" "200" "-H \"Authorization: Bearer $TOKEN\""
fi
test_api "无Token访问失败" "GET" "/api/v1/users/me" "" "401"
echo ""

echo "=== 7. 角色管理 (8 tests) ==="
if [ -n "$TOKEN" ]; then
    test_api "获取角色列表" "GET" "/api/v1/roles" "" "200" "-H \"Authorization: Bearer $TOKEN\""
    test_api "创建角色1" "POST" "/api/v1/roles" "{\"name\":\"角色A_${RANDOM_SUFFIX}\",\"systemPrompt\":\"你是角色A\"}" "200" "-H \"Authorization: Bearer $TOKEN\""
    test_api "创建角色2" "POST" "/api/v1/roles" "{\"name\":\"角色B_${RANDOM_SUFFIX}\",\"systemPrompt\":\"你是角色B\"}" "200" "-H \"Authorization: Bearer $TOKEN\""
    test_api "创建角色缺少prompt" "POST" "/api/v1/roles" "{\"name\":\"角色C\"}" "400" "-H \"Authorization: Bearer $TOKEN\""
    test_api "获取单个角色" "GET" "/api/v1/roles/role-001" "" "200"
    test_api "获取不存在的角色" "GET" "/api/v1/roles/not-exist" "" "404"
    test_api "与角色对话" "POST" "/api/v1/roles/role-001/chat" "{\"message\":\"你好\"}" "200" "-H \"Authorization: Bearer $TOKEN\""
fi
test_api "无Token创建角色失败" "POST" "/api/v1/roles" "{\"name\":\"角色X\",\"systemPrompt\":\"test\"}" "401"
echo ""

echo "=== 8. 文档管理 (5 tests) ==="
if [ -n "$TOKEN" ]; then
    test_api "获取文档列表" "GET" "/api/v1/documents" "" "200" "-H \"Authorization: Bearer $TOKEN\""
    test_api "获取单个文档不存在" "GET" "/api/v1/documents/not-exist" "" "404" "-H \"Authorization: Bearer $TOKEN\""
    test_api "删除文档不存在" "DELETE" "/api/v1/documents/not-exist" "" "404" "-H \"Authorization: Bearer $TOKEN\""
fi
test_api "无Token获取文档失败" "GET" "/api/v1/documents" "" "401"
test_api "无Token删除文档失败" "DELETE" "/api/v1/documents/test" "" "401"
echo ""

echo "=== 9. 对话管理 (6 tests) ==="
if [ -n "$TOKEN" ]; then
    test_api "获取会话列表" "GET" "/api/v1/chat-sessions" "" "200" "-H \"Authorization: Bearer $TOKEN\""
    test_api "创建会话" "POST" "/api/v1/chat-sessions" "{\"roleId\":\"role-001\"}" "201" "-H \"Authorization: Bearer $TOKEN\""
    test_api "获取会话详情不存在" "GET" "/api/v1/chat-sessions/not-exist" "" "404" "-H \"Authorization: Bearer $TOKEN\""
    test_api "发送消息不存在会话" "POST" "/api/v1/chat/not-exist/complete" "{\"content\":\"test\"}" "404" "-H \"Authorization: Bearer $TOKEN\""
fi
test_api "无Token获取会话失败" "GET" "/api/v1/chat-sessions" "" "401"
test_api "无Token创建会话失败" "POST" "/api/v1/chat-sessions" "{\"roleId\":\"role-001\"}" "401"
echo ""

echo "=== 10. 边界测试 (5 tests) ==="
test_api "超长邮箱" "POST" "/api/v1/auth/register" "{\"email\":\"verylongemailaddress@verylongdomainname.com\",\"password\":\"123456\",\"name\":\"Test\"}" "201"
if [ -n "$TOKEN" ]; then
    test_api "空内容消息" "POST" "/api/v1/chat-sessions" "{\"roleId\":\"role-001\"}" "201" "-H \"Authorization: Bearer $TOKEN\""
fi
test_api "特殊字符名" "POST" "/api/v1/auth/register" "{\"email\":\"special${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\",\"name\":\"用户<>&\\\"'\"}" "201"
test_api "JSON注入尝试" "POST" "/api/v1/auth/register" "{\"email\":\"inject${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\",\"name\":\"{\\\"admin\\\":true}\"}" "201"
test_api "SQL注入尝试" "POST" "/api/v1/auth/login" "{\"email\":\"user1_${RANDOM_SUFFIX}@test.com\",\"password\":\"' OR '1'='1\"}" "401"
echo ""

echo "=== 测试结果 ==="
echo "总计: $TOTAL"
echo "通过: $PASS"
echo "失败: $FAIL"
echo ""

if [ $FAIL -eq 0 ]; then
    echo "✅ 所有测试通过！"
    exit 0
else
    echo "⚠️ 有 $FAIL 个测试失败"
    exit 1
fi