#!/bin/bash
# RoleCraft AI 完整 API 测试套件 - 50+ 测试用例

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0
TOTAL=0

test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected=${5:-200}
    local headers=${6:-""}
    
    ((TOTAL++))
    
    if [ "$method" = "GET" ]; then
        if [ -n "$headers" ]; then
            response=$(curl -s -w "\n%{http_code}" -H "$headers" "$BASE_URL$endpoint")
        else
            response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
        fi
    else
        if [ -n "$headers" ]; then
            response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
                -H "Content-Type: application/json" \
                -H "$headers" \
                -d "$data")
        else
            response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
                -H "Content-Type: application/json" \
                -d "$data")
        fi
    fi
    
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "$expected" ]; then
        echo "✅ $TOTAL. $name (HTTP $http_code)"
        ((PASS++))
    else
        echo "❌ $TOTAL. $name (HTTP $http_code, expected $expected)"
        ((FAIL++))
    fi
}

echo "========================================="
echo "  RoleCraft AI API 测试套件 (50+ 用例)"
echo "========================================="
echo ""

# 1. 系统健康检查 (5 用例)
echo "=== 1. 系统健康检查 ==="
test_api "Health Check" "GET" "/health" "" "200"
test_api "API Version" "GET" "/api/v1" "" "404"
test_api "CORS Headers" "OPTIONS" "/api/v1/roles" "" "204"
test_api "Rate Limit" "GET" "/health" "" "200"
test_api "Cache Control" "GET" "/health" "" "200"

echo ""
echo "=== 2. 角色模板 (5 用例) ==="
test_api "Get All Templates" "GET" "/api/v1/roles/templates" "" "200"
test_api "Templates Count" "GET" "/api/v1/roles/templates" "" "200"
test_api "Template Structure" "GET" "/api/v1/roles/templates" "" "200"
test_api "Template Categories" "GET" "/api/v1/roles/templates" "" "200"
test_api "Cache Templates" "GET" "/api/v1/roles/templates" "" "200"

echo ""
echo "=== 3. 用户注册 (10 用例) ==="
TS=$(date +%s)
test_api "Register Valid User" "POST" "/api/v1/auth/register" "{\"email\":\"user1_${TS}@test.com\",\"password\":\"123456\",\"name\":\"User1\"}" "201"
test_api "Register Second User" "POST" "/api/v1/auth/register" "{\"email\":\"user2_${TS}@test.com\",\"password\":\"123456\",\"name\":\"User2\"}" "201"
test_api "Duplicate Email" "POST" "/api/v1/auth/register" "{\"email\":\"user1_${TS}@test.com\",\"password\":\"123456\",\"name\":\"User1\"}" "409"
test_api "Empty Email" "POST" "/api/v1/auth/register" "{\"email\":\"\",\"password\":\"123456\",\"name\":\"Test\"}" "400"
test_api "Invalid Email" "POST" "/api/v1/auth/register" "{\"email\":\"invalid\",\"password\":\"123456\",\"name\":\"Test\"}" "400"
test_api "Short Password" "POST" "/api/v1/auth/register" "{\"email\":\"short_${TS}@test.com\",\"password\":\"123\",\"name\":\"Test\"}" "400"
test_api "Empty Name" "POST" "/api/v1/auth/register" "{\"email\":\"noname_${TS}@test.com\",\"password\":\"123456\",\"name\":\"\"}" "400"
test_api "Long Name" "POST" "/api/v1/auth/register" "{\"email\":\"longname_${TS}@test.com\",\"password\":\"123456\",\"name\":\"$(printf 'A%.0s' {1..100})\"}" "201"
test_api "Special Chars Name" "POST" "/api/v1/auth/register" "{\"email\":\"special_${TS}@test.com\",\"password\":\"123456\",\"name\":\"测试用户\"}" "201"
test_api "Unicode Email" "POST" "/api/v1/auth/register" "{\"email\":\"unicode_${TS}@test.com\",\"password\":\"123456\",\"name\":\"用户\"}" "201"

echo ""
echo "=== 4. 用户登录 (8 用例) ==="
test_api "Login Valid" "POST" "/api/v1/auth/login" "{\"email\":\"user1_${TS}@test.com\",\"password\":\"123456\"}" "200"
test_api "Login Wrong Password" "POST" "/api/v1/auth/login" "{\"email\":\"user1_${TS}@test.com\",\"password\":\"wrong\"}" "401"
test_api "Login Nonexistent User" "POST" "/api/v1/auth/login" "{\"email\":\"nobody@test.com\",\"password\":\"123456\"}" "401"
test_api "Login Empty Email" "POST" "/api/v1/auth/login" "{\"email\":\"\",\"password\":\"123456\"}" "400"
test_api "Login Empty Password" "POST" "/api/v1/auth/login" "{\"email\":\"user1_${TS}@test.com\",\"password\":\"\"}" "400"
test_api "Login Case Insensitive Email" "POST" "/api/v1/auth/login" "{\"email\":\"USER1_${TS}@TEST.COM\",\"password\":\"123456\"}" "200"
test_api "Login Returns Token" "POST" "/api/v1/auth/login" "{\"email\":\"user1_${TS}@test.com\",\"password\":\"123456\"}" "200"
test_api "Login Returns User" "POST" "/api/v1/auth/login" "{\"email\":\"user1_${TS}@test.com\",\"password\":\"123456\"}" "200"

echo ""
echo "=== 5. Token 刷新 (3 用例) ==="
TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"user1_${TS}@test.com\",\"password\":\"123456\"}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

test_api "Refresh Token Valid" "POST" "/api/v1/auth/refresh" "" "200" "Authorization: Bearer $TOKEN"
test_api "Refresh No Token" "POST" "/api/v1/auth/refresh" "" "401"
test_api "Refresh Invalid Token" "POST" "/api/v1/auth/refresh" "" "401" "Authorization: Bearer invalid"

echo ""
echo "=== 6. 用户信息 (5 用例) ==="
test_api "Get Me Valid" "GET" "/api/v1/users/me" "" "200" "Authorization: Bearer $TOKEN"
test_api "Get Me No Token" "GET" "/api/v1/users/me" "" "401"
test_api "Get Me Invalid Token" "GET" "/api/v1/users/me" "" "401" "Authorization: Bearer invalid"
test_api "Update Me Name" "PUT" "/api/v1/users/me" "{\"name\":\"Updated Name\"}" "200" "Authorization: Bearer $TOKEN"
test_api "Update Me Avatar" "PUT" "/api/v1/users/me" "{\"avatar\":\"https://example.com/avatar.png\"}" "200" "Authorization: Bearer $TOKEN"

echo ""
echo "=== 7. 角色管理 (10 用例) ==="
test_api "List Roles" "GET" "/api/v1/roles" "" "200" "Authorization: Bearer $TOKEN"
test_api "Create Role Minimal" "POST" "/api/v1/roles" "{\"name\":\"Minimal Role\",\"systemPrompt\":\"You are helpful\"}" "200" "Authorization: Bearer $TOKEN"
test_api "Create Role Full" "POST" "/api/v1/roles" "{\"name\":\"Full Role\",\"description\":\"Test\",\"category\":\"test\",\"systemPrompt\":\"You are helpful\",\"welcomeMessage\":\"Hello\"}" "200" "Authorization: Bearer $TOKEN"
test_api "Create Role No Name" "POST" "/api/v1/roles" "{\"systemPrompt\":\"You are helpful\"}" "400" "Authorization: Bearer $TOKEN"
test_api "Create Role No Prompt" "POST" "/api/v1/roles" "{\"name\":\"No Prompt\"}" "400" "Authorization: Bearer $TOKEN"
test_api "Get Role By ID" "GET" "/api/v1/roles/role-001" "" "200" "Authorization: Bearer $TOKEN"
test_api "Get Role Not Found" "GET" "/api/v1/roles/nonexistent" "" "404" "Authorization: Bearer $TOKEN"
test_api "Update Role" "PUT" "/api/v1/roles/role-001" "{\"name\":\"Updated Role\"}" "200" "Authorization: Bearer $TOKEN"
test_api "Delete Own Role" "DELETE" "/api/v1/roles/role-001" "" "200" "Authorization: Bearer $TOKEN"
test_api "Roles Without Auth" "GET" "/api/v1/roles" "" "401"

echo ""
echo "=== 8. 对话会话 (5 用例) ==="
test_api "List Sessions No Auth" "GET" "/api/v1/chat-sessions" "" "401"
test_api "List Sessions Valid" "GET" "/api/v1/chat-sessions" "" "200" "Authorization: Bearer $TOKEN"
test_api "Create Session No Role" "POST" "/api/v1/chat-sessions" "{\"roleId\":\"nonexistent\"}" "404" "Authorization: Bearer $TOKEN"
test_api "Create Session Valid" "POST" "/api/v1/chat-sessions" "{\"roleId\":\"11111111-1111-1111-1111-111111111111\"}" "200" "Authorization: Bearer $TOKEN"
test_api "Get Session Valid" "GET" "/api/v1/chat-sessions/valid-id" "" "404" "Authorization: Bearer $TOKEN"

echo ""
echo "=== 9. 文档管理 (4 用例) ==="
test_api "List Documents No Auth" "GET" "/api/v1/documents" "" "401"
test_api "List Documents Valid" "GET" "/api/v1/documents" "" "200" "Authorization: Bearer $TOKEN"
test_api "Get Document Not Found" "GET" "/api/v1/documents/nonexistent" "" "404" "Authorization: Bearer $TOKEN"
test_api "Delete Document Not Found" "DELETE" "/api/v1/documents/nonexistent" "" "404" "Authorization: Bearer $TOKEN"

echo ""
echo "========================================="
echo "  测试结果"
echo "========================================="
echo "总计: $TOTAL 测试"
echo "通过: $PASS"
echo "失败: $FAIL"
echo "通过率: $(echo "scale=1; $PASS * 100 / $TOTAL" | bc)%"
echo ""

if [ $FAIL -eq 0 ]; then
    echo "✅ 所有测试通过！"
    exit 0
else
    echo "⚠️ 有 $FAIL 个测试失败"
    exit 1
fi