#!/bin/bash

# 知识库管理增强 - 功能测试脚本
# 版本：v2.0
# 日期：2026-02-27

# 配置
API_BASE="http://localhost:8080/api/v1"
TOKEN="${1:-}"

if [ -z "$TOKEN" ]; then
    echo "❌ 请提供 JWT Token"
    echo "用法：$0 <jwt_token>"
    exit 1
fi

HEADERS="Authorization: Bearer $TOKEN"
CONTENT_TYPE="Content-Type: application/json"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "======================================"
echo "知识库管理增强 - 功能测试"
echo "======================================"
echo ""

# 测试计数器
TOTAL=0
PASSED=0
FAILED=0

# 测试函数
test_api() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    
    TOTAL=$((TOTAL + 1))
    echo -n "测试：$name ... "
    
    if [ "$method" == "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" -X GET \
            -H "$HEADERS" \
            -H "$CONTENT_TYPE" \
            "${API_BASE}${endpoint}")
    elif [ "$method" == "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST \
            -H "$HEADERS" \
            -H "$CONTENT_TYPE" \
            -d "$data" \
            "${API_BASE}${endpoint}")
    elif [ "$method" == "PUT" ]; then
        response=$(curl -s -w "\n%{http_code}" -X PUT \
            -H "$HEADERS" \
            -H "$CONTENT_TYPE" \
            -d "$data" \
            "${API_BASE}${endpoint}")
    elif [ "$method" == "DELETE" ]; then
        response=$(curl -s -w "\n%{http_code}" -X DELETE \
            -H "$HEADERS" \
            -H "$CONTENT_TYPE" \
            "${API_BASE}${endpoint}")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✅ 通过${NC} (HTTP $http_code)"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}❌ 失败${NC} (HTTP $http_code)"
        echo "响应：$body"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

echo "1️⃣  文件夹管理测试"
echo "--------------------------------------"

# 创建文件夹
test_api "创建文件夹" "POST" "/folders" '{"name": "测试文件夹", "parentId": ""}'
FOLDER_ID=$(curl -s -X POST -H "$HEADERS" -H "$CONTENT_TYPE" -d '{"name": "临时文件夹"}' "${API_BASE}/folders" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

# 获取文件夹列表
test_api "获取文件夹列表" "GET" "/folders" ""

if [ -n "$FOLDER_ID" ]; then
    # 删除文件夹
    test_api "删除文件夹" "DELETE" "/folders/$FOLDER_ID" ""
fi

echo ""
echo "2️⃣  文档批量操作测试"
echo "--------------------------------------"

# 注意：文件上传需要实际文件，这里仅测试接口
# test_api "批量上传文档" "POST" "/documents" "-F file=@test.pdf"

# 获取文档列表
test_api "获取文档列表" "GET" "/documents?limit=10" ""

# 高级搜索
test_api "高级搜索" "POST" "/documents/search" '{"query": "", "topN": 10, "filters": {}, "sortBy": "date", "sortOrder": "desc"}'

echo ""
echo "3️⃣  文档预览和下载测试"
echo "--------------------------------------"

# 获取第一个文档 ID
DOC_ID=$(curl -s -X GET -H "$HEADERS" "${API_BASE}/documents?limit=1" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -n "$DOC_ID" ]; then
    # 获取文档详情
    test_api "获取文档详情" "GET" "/documents/$DOC_ID" ""
    
    # 获取文档状态
    test_api "获取文档状态" "GET" "/documents/$DOC_ID/status" ""
    
    # 预览文档 (如果有文件)
    # test_api "预览文档" "GET" "/documents/$DOC_ID/preview" ""
    
    # 下载文档 (如果有文件)
    # test_api "下载文档" "GET" "/documents/$DOC_ID/download" ""
fi

echo ""
echo "4️⃣  批量操作接口测试"
echo "--------------------------------------"

# 批量删除 (空列表)
test_api "批量删除 (空)" "DELETE" "/documents/batch" '{"ids": []}'

# 批量移动 (空列表)
test_api "批量移动 (空)" "PUT" "/documents/batch/move" '{"ids": [], "folderId": ""}'

# 批量标签 (空列表)
test_api "批量标签 (空)" "PUT" "/documents/batch/tags" '{"ids": [], "tags": []}'

echo ""
echo "======================================"
echo "测试结果汇总"
echo "======================================"
echo "总测试数：$TOTAL"
echo -e "通过：${GREEN}$PASSED${NC}"
echo -e "失败：${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 所有测试通过！${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  部分测试失败，请检查日志${NC}"
    exit 1
fi
