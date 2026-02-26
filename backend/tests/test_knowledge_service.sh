#!/bin/bash

# RoleCraft AI 知识库服务测试脚本

BASE_URL="http://localhost:8080/api/v1"
TOKEN=""

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 1. 登录
login() {
    echo_info "=== Step 1: Login ==="
    
    response=$(curl -s -X POST "${BASE_URL}/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"test@example.com","password":"password123"}')
    
    TOKEN=$(echo $response | jq -r '.data.token')
    
    if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
        echo_error "Login failed: $response"
        return 1
    fi
    
    echo_info "Login successful, token: ${TOKEN:0:30}..."
    return 0
}

# 2. 上传文档
upload_document() {
    local file=$1
    
    echo_info "=== Step 2: Upload Document ==="
    
    if [ ! -f "$file" ]; then
        echo_warn "File not found, creating test file..."
        echo "Test document content for vector search testing." > /tmp/test_doc.txt
        file="/tmp/test_doc.txt"
    fi
    
    response=$(curl -s -X POST "${BASE_URL}/documents" \
        -H "Authorization: Bearer $TOKEN" \
        -F "file=@$file")
    
    echo "Upload response: $response" | jq .
    
    DOC_ID=$(echo $response | jq -r '.data.id')
    
    if [ -z "$DOC_ID" ] || [ "$DOC_ID" = "null" ]; then
        echo_error "Upload failed"
        return 1
    fi
    
    echo_info "Document uploaded, ID: $DOC_ID"
    echo $DOC_ID
    return 0
}

# 3. 检查文档状态
check_status() {
    local doc_id=$1
    
    echo_info "=== Step 3: Check Document Status ==="
    
    response=$(curl -s -X GET "${BASE_URL}/documents/$doc_id" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "Status response: $response" | jq .
    
    status=$(echo $response | jq -r '.data.status')
    echo_info "Document status: $status"
    
    echo $status
    return 0
}

# 4. 向量搜索
vector_search() {
    local query=$1
    local topn=${2:-4}
    
    echo_info "=== Step 4: Vector Search ==="
    echo_info "Query: $query, TopN: $topn"
    
    response=$(curl -s -X POST "${BASE_URL}/documents/search" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\\\"query\\\":\\\"$query\\\",\\\"topN\\\":$topn}")
    
    echo "Search response: $response" | jq .
    return 0
}

# 5. 删除文档
delete_document() {
    local doc_id=$1
    
    echo_info "=== Step 5: Delete Document ==="
    
    response=$(curl -s -X DELETE "${BASE_URL}/documents/$doc_id" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "Delete response: $response" | jq .
    
    if [ $? -eq 0 ]; then
        echo_info "Document deleted successfully"
    else
        echo_error "Delete failed"
    fi
    return 0
}

# 等待文档处理完成
wait_for_processing() {
    local doc_id=$1
    local max_attempts=${2:-10}
    local delay=${3:-3}
    
    echo_info "=== Waiting for document processing ==="
    
    for i in $(seq 1 $max_attempts); do
        status=$(check_status $doc_id)
        
        if [ "$status" = "completed" ]; then
            echo_info "Document processing completed!"
            return 0
        elif [ "$status" = "failed" ]; then
            echo_error "Document processing failed"
            return 1
        fi
        
        echo_info "Attempt $i/$max_attempts, status: $status, waiting ${delay}s..."
        sleep $delay
    done
    
    echo_warn "Timeout waiting for processing"
    return 1
}

# 主测试流程
run_full_test() {
    echo_info "========================================="
    echo_info "RoleCraft AI Knowledge Service Test"
    echo_info "========================================="
    
    # 登录
    if ! login; then
        echo_error "Login failed, aborting test"
        exit 1
    fi
    
    # 上传文档
    doc_id=$(upload_document "$1")
    if [ -z "$doc_id" ]; then
        echo_error "Upload failed, aborting test"
        exit 1
    fi
    
    # 等待处理
    if ! wait_for_processing $doc_id; then
        echo_warn "Processing may still be in progress, continuing..."
    fi
    
    # 向量搜索
    vector_search "测试" 4
    
    # 删除文档
    delete_document $doc_id
    
    echo_info "========================================="
    echo_info "Test completed!"
    echo_info "========================================="
}

# 显示帮助
show_help() {
    echo "RoleCraft AI Knowledge Service Test Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  full [file]     Run full test flow (default)"
    echo "  login           Login and get token"
    echo "  upload [file]   Upload a document"
    echo "  search [query]  Vector search"
    echo "  status [id]     Check document status"
    echo "  delete [id]     Delete a document"
    echo "  help            Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 full ./test.pdf"
    echo "  $0 search \"人工智能\""
    echo "  $0 status abc-123-def"
}

# 命令行参数处理
case "${1:-full}" in
    full)
        run_full_test "$2"
        ;;
    login)
        login
        ;;
    upload)
        login && upload_document "$2"
        ;;
    search)
        login && vector_search "${2:-测试}" 4
        ;;
    status)
        login && check_status "$2"
        ;;
    delete)
        login && delete_document "$2"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
