# AnythingLLM Integration Guide

## Overview

This document describes the integration between RoleCraft AI and AnythingLLM for RAG-powered conversations.

## Configuration

### Environment Variables

Add the following environment variables to your `.env` file:

```bash
# AnythingLLM Configuration
ANYTHINGLLM_URL=http://150.109.21.115:3001
ANYTHINGLLM_KEY=sk-WaUmgZsMxgeHOpp8SJxK1rmVQxiwfiDJ
```

## API Endpoints

### 1. Create Chat Session with AnythingLLM Workspace

```bash
POST /api/v1/chat-sessions
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "roleId": "role-uuid",
  "title": "My RAG Chat",
  "mode": "rag",
  "anythingLLMSlug": "my-workspace-slug"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "session-uuid",
    "userId": "user-uuid",
    "roleId": "role-uuid",
    "title": "My RAG Chat",
    "mode": "rag",
    "anythingLLMSlug": "my-workspace-slug",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

### 2. Send Message (Non-Streaming)

```bash
POST /api/v1/chat/:id/complete
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "content": "What is in the documents?"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "userMessage": {
      "id": "msg-uuid",
      "sessionId": "session-uuid",
      "role": "user",
      "content": "What is in the documents?",
      "createdAt": "2024-01-01T00:00:00Z"
    },
    "assistantMessage": {
      "id": "msg-uuid",
      "sessionId": "session-uuid",
      "role": "assistant",
      "content": "Based on the documents...",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  }
}
```

### 3. Send Message (Streaming - SSE)

```bash
POST /api/v1/chat/:id/stream
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "content": "Explain the key concepts"
}
```

**Response (Server-Sent Events):**
```
data: {"content":"The","done":false}

data: {"content":" key","done":false}

data: {"content":" concepts","done":false}

data: {"done":true}
```

### 4. Sync Chat History from AnythingLLM

```bash
GET /api/v1/chat-sessions/:id/sync
Authorization: Bearer {jwt_token}
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "history": [
      {
        "message": "User message",
        "response": "Assistant response",
        "timestamp": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 5. Delete Message

```bash
DELETE /api/v1/chat-sessions/:id/messages/:msgId
Authorization: Bearer {jwt_token}
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "deleted": true
  }
}
```

## Workspace Authentication

All chat endpoints now include `WorkspaceAuth()` middleware that ensures:
- Users can only access their own chat sessions
- Unauthorized access returns `403 Forbidden`
- The middleware validates session ownership before processing requests

## Flow Diagram

```
User → RoleCraft API → AnythingLLM → RAG Processing → Response
  │         │              │
  │         │              └─→ Vector Database
  │         └─→ Workspace Slug Validation
  └─→ JWT Authentication
```

## AnythingLLM API Reference

### Chat Endpoint
```bash
curl -X POST "http://150.109.21.115:3001/api/v1/workspace/{slug}/chat" \
  -H "Authorization: Bearer sk-WaUmgZsMxgeHOpp8SJxK1rmVQxiwfiDJ" \
  -H "Content-Type: application/json" \
  -d '{"message": "你好", "mode": "chat"}'
```

### Stream Chat Endpoint
```bash
curl -X POST "http://150.109.21.115:3001/api/v1/workspace/{slug}/stream-chat" \
  -H "Authorization: Bearer sk-WaUmgZsMxgeHOpp8SJxK1rmVQxiwfiDJ" \
  -H "Content-Type: application/json" \
  -d '{"message": "你好", "mode": "chat"}'
```

## Testing

### Test 1: Create Session
```bash
curl -X POST http://localhost:8080/api/v1/chat-sessions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "roleId": "test-role",
    "title": "Test Chat",
    "anythingLLMSlug": "test-workspace"
  }'
```

### Test 2: Send Message
```bash
curl -X POST http://localhost:8080/api/v1/chat/SESSION_ID/complete \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello, tell me about the documents"}'
```

### Test 3: Stream Response
```bash
curl -X POST http://localhost:8080/api/v1/chat/SESSION_ID/stream \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "Explain step by step"}'
```

### Test 4: Permission Control
```bash
# Try to access another user's session (should fail with 403)
curl -X POST http://localhost:8080/api/v1/chat/OTHER_USER_SESSION_ID/complete \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "Test"}'
```

## Migration Notes

### Before (OpenAI Direct)
- Messages sent directly to OpenAI API
- No RAG capabilities
- Limited to model's training data

### After (AnythingLLM)
- Messages sent to AnythingLLM workspace
- Automatic RAG retrieval from uploaded documents
- Context-aware responses based on knowledge base
- Workspace-level access control

## Troubleshooting

### Error: "anythingllm slug not found"
- Ensure the chat session has `anythingLLMSlug` set
- Check if the role has `anythingllm_slug` in `modelConfig`

### Error: "access denied to this workspace"
- Verify JWT token is valid
- Ensure the session belongs to the authenticated user

### Error: "anythingllm API error"
- Check `ANYTHINGLLM_URL` and `ANYTHINGLLM_KEY` configuration
- Verify AnythingLLM service is running
- Check network connectivity

## Security Considerations

1. **API Key Protection**: Store `ANYTHINGLLM_KEY` in environment variables, never commit to code
2. **Workspace Isolation**: Each user session maps to a specific AnythingLLM workspace
3. **JWT Validation**: All endpoints require valid JWT authentication
4. **Ownership Verification**: WorkspaceAuth middleware prevents cross-user access
