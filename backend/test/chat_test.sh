#!/bin/bash

# RoleCraft AI - AnythingLLM Integration Test Script
# This script tests the refactored chat service

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

echo_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if JWT token is provided
if [ -z "$JWT_TOKEN" ]; then
    echo_error "JWT_TOKEN environment variable is not set"
    echo_info "Please login first and set JWT_TOKEN"
    echo_info "Example: export JWT_TOKEN=your_jwt_token_here"
    exit 1
fi

echo_info "Testing AnythingLLM Integration"
echo_info "Base URL: $BASE_URL"
echo ""

# Test 1: List Chat Sessions
echo_info "Test 1: List Chat Sessions"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/chat-sessions" \
    -H "Authorization: Bearer $JWT_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo_success "List sessions: $HTTP_CODE"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
    echo_error "List sessions failed: $HTTP_CODE"
fi
echo ""

# Test 2: Create Chat Session with AnythingLLM Slug
echo_info "Test 2: Create Chat Session"
ROLE_ID="${ROLE_ID:-test-role}"
SLUG="${SLUG:-test-workspace}"

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/chat-sessions" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"roleId\": \"$ROLE_ID\",
        \"title\": \"Test Chat with AnythingLLM\",
        \"mode\": \"rag\",
        \"anythingLLMSlug\": \"$SLUG\"
    }")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
    echo_success "Create session: $HTTP_CODE"
    SESSION_ID=$(echo "$BODY" | jq -r '.data.id' 2>/dev/null || echo "")
    if [ -n "$SESSION_ID" ] && [ "$SESSION_ID" != "null" ]; then
        echo_info "Session ID: $SESSION_ID"
        export SESSION_ID
    fi
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
    echo_error "Create session failed: $HTTP_CODE"
    echo "$BODY"
fi
echo ""

# Test 3: Send Message (Non-Streaming)
if [ -n "${SESSION_ID:-}" ]; then
    echo_info "Test 3: Send Message (Non-Streaming)"
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/chat/$SESSION_ID/complete" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"content": "你好，请介绍一下你自己"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" = "200" ]; then
        echo_success "Send message: $HTTP_CODE"
        echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    else
        echo_error "Send message failed: $HTTP_CODE"
        echo "$BODY"
    fi
    echo ""

    # Test 4: Send Message (Streaming)
    echo_info "Test 4: Send Message (Streaming)"
    echo "Streaming response:"
    curl -X POST "$BASE_URL/api/v1/chat/$SESSION_ID/stream" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"content": "请用流式方式回复"}' \
        2>/dev/null
    echo ""
    echo ""

    # Test 5: Get Session Details
    echo_info "Test 5: Get Session Details"
    RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/chat-sessions/$SESSION_ID" \
        -H "Authorization: Bearer $JWT_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" = "200" ]; then
        echo_success "Get session: $HTTP_CODE"
        echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    else
        echo_error "Get session failed: $HTTP_CODE"
    fi
    echo ""

    # Test 6: Sync Session
    echo_info "Test 6: Sync Session from AnythingLLM"
    RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/chat-sessions/$SESSION_ID/sync" \
        -H "Authorization: Bearer $JWT_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" = "200" ]; then
        echo_success "Sync session: $HTTP_CODE"
        echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    else
        echo_error "Sync session failed: $HTTP_CODE"
        echo "$BODY"
    fi
    echo ""

    # Test 7: Delete Message
    echo_info "Test 7: Delete Message (Permission Test)"
    # First get a message ID
    MSG_ID=$(echo "$BODY" | jq -r '.data.messages[0].id' 2>/dev/null || echo "test-msg-id")
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/api/v1/chat-sessions/$SESSION_ID/messages/$MSG_ID" \
        -H "Authorization: Bearer $JWT_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "404" ]; then
        echo_success "Delete message: $HTTP_CODE"
        echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    else
        echo_error "Delete message failed: $HTTP_CODE"
        echo "$BODY"
    fi
    echo ""
fi

# Test 8: Permission Control (Access Denied)
echo_info "Test 8: Permission Control (Should return 403)"
FAKE_SESSION_ID="fake-session-id-not-owned-by-user"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/chat/$FAKE_SESSION_ID/complete" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"content": "Test permission"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "404" ]; then
    echo_success "Permission control working: $HTTP_CODE"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
    echo_error "Permission control may not be working correctly: $HTTP_CODE"
fi
echo ""

echo_info "All tests completed!"
