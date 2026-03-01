# Quick Start - AnythingLLM Integration

## 1. Setup Environment

```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/backend

# Create .env file
cat > .env << EOF
ENV=development
PORT=8080
DATABASE_URL=/Users/claw/.openclaw/workspace-work/rolecraft-ai/backend/rolecraft.db
JWT_SECRET=your-secret-key-change-in-production
ANYTHINGLLM_URL=http://43.134.234.4:3001
ANYTHINGLLM_KEY=sk-WaUmgZsMxgeHOpp8SJxK1rmVQxiwfiDJ
EOF
```

## 2. Start Server

```bash
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

## 3. Login and Get JWT Token

```bash
# Register (if needed)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Save the token from response
export JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 4. Create a Chat Session

```bash
curl -X POST http://localhost:8080/api/v1/chat-sessions \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "roleId": "your-role-id",
    "title": "My RAG Chat",
    "mode": "rag",
    "anythingLLMSlug": "your-workspace-slug"
  }'

# Save session ID from response
export SESSION_ID="session-uuid-from-response"
```

## 5. Send a Message

```bash
# Non-streaming
curl -X POST http://localhost:8080/api/v1/chat/$SESSION_ID/complete \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "你好，请根据文档回答问题"}'

# Streaming
curl -X POST http://localhost:8080/api/v1/chat/$SESSION_ID/stream \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "请流式输出答案"}'
```

## 6. Run Automated Tests

```bash
export JWT_TOKEN="your-jwt-token"
export ROLE_ID="your-role-id"
export SLUG="your-workspace-slug"

./test/chat_test.sh
```

## 7. Verify AnythingLLM Connection

```bash
# Test direct AnythingLLM API
curl -X POST "http://43.134.234.4:3001/api/v1/workspace/your-slug/chat" \
  -H "Authorization: Bearer sk-WaUmgZsMxgeHOpp8SJxK1rmVQxiwfiDJ" \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello", "mode": "chat"}'
```

## Common Issues

### Issue: "anythingllm slug not found"
**Solution**: Make sure to include `anythingLLMSlug` when creating the session

### Issue: "invalid token"
**Solution**: Re-login and get a fresh JWT token

### Issue: Connection refused
**Solution**: Check if AnythingLLM service is running at the configured URL

## Next Steps

1. Upload documents to your AnythingLLM workspace
2. Configure embedding settings in AnythingLLM
3. Test RAG retrieval with document-specific questions
4. Integrate with frontend application

## Resources

- Full documentation: `docs/ANYTHINGLLM_INTEGRATION.md`
- API refactoring summary: `REFACTOR_SUMMARY.md`
- Test script: `test/chat_test.sh`
