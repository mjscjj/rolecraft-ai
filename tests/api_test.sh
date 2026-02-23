#!/bin/bash
# RoleCraft AI 完整 API 测试

echo "=== RoleCraft AI API 测试 ==="
echo ""

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0

# 生成随机用户名避免冲突
RANDOM_SUFFIX=$(date +%s)

test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected=${5:-200}
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "$expected" ]; then
        echo "✅ $name (HTTP $http_code)"
        ((PASS++))
    else
        echo "❌ $name (HTTP $http_code, expected $expected)"
        ((FAIL++))
    fi
}

echo "=== 1. 系统健康检查 ==="
test_api "Health Check" "GET" "/health" "" "200"

echo ""
echo "=== 2. 角色模板 ==="
test_api "Get Templates" "GET" "/api/v1/roles/templates" "" "200"

echo ""
echo "=== 3. 用户认证 ==="
test_api "Register User A" "POST" "/api/v1/auth/register" "{\"email\":\"userA_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\",\"name\":\"UserA\"}" "201"
test_api "Register User B" "POST" "/api/v1/auth/register" "{\"email\":\"userB_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\",\"name\":\"UserB\"}" "201"
test_api "Login Success" "POST" "/api/v1/auth/login" "{\"email\":\"userA_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\"}" "200"
test_api "Login Wrong Password" "POST" "/api/v1/auth/login" "{\"email\":\"userA_${RANDOM_SUFFIX}@test.com\",\"password\":\"wrong\"}" "401"

echo ""
echo "=== 4. 角色管理 (需要认证) ==="
TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"userA_${RANDOM_SUFFIX}@test.com\",\"password\":\"123456\"}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$TOKEN" ]; then
    ROLES_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/roles")
    if [ -n "$ROLES_RESPONSE" ]; then
        echo "✅ Get Roles List (with token)"
        ((PASS++))
    else
        echo "❌ Get Roles List failed"
        ((FAIL++))
    fi

    CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/roles" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{\"name\":\"测试角色_${RANDOM_SUFFIX}\",\"description\":\"测试描述\",\"systemPrompt\":\"你是一个测试角色\"}")
    CREATE_CODE=$(echo "$CREATE_RESPONSE" | tail -n1)
    if [ "$CREATE_CODE" = "200" ] || [ "$CREATE_CODE" = "201" ]; then
        echo "✅ Create Role (HTTP $CREATE_CODE)"
        ((PASS++))
    else
        echo "❌ Create Role (HTTP $CREATE_CODE)"
        ((FAIL++))
    fi
else
    echo "❌ Failed to get auth token"
    ((FAIL+=2))
fi

echo ""
echo "=== 测试结果 ==="
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