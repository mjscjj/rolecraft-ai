# AnythingLLM API 参考文档

**版本**: v1.0  
**来源**: https://github.com/Mintplex-Labs/anything-llm  
**更新时间**: 2026-02-26

---

## 📋 目录

1. [Workspace API](#workspace-api)
2. [Document API](#document-api)
3. [OpenAI Compatible API](#openai-compatible-api)
4. [Chat API](#chat-api)

---

## Workspace API

### 1. 创建 Workspace

**端点**: `POST /api/v1/workspace/new`

**请求头**:
```
Authorization: Bearer {API_KEY}
Content-Type: application/json
```

**请求体**:
```json
{
  "name": "My Workspace",
  "similarityThreshold": 0.7,
  "openAiTemp": 0.7,
  "openAiHistory": 20,
  "openAiPrompt": "Custom system prompt",
  "queryRefusalResponse": "Custom refusal message",
  "chatMode": "chat",
  "topN": 4
}
```

**响应 (200)**:
```json
{
  "workspace": {
    "id": 79,
    "name": "My Workspace",
    "slug": "my-workspace",
    "createdAt": "2023-08-17T00:45:03Z",
    "openAiTemp": 0.7,
    "lastUpdatedAt": "2023-08-17T00:45:03Z",
    "openAiHistory": 20,
    "openAiPrompt": "Custom system prompt",
    "similarityThreshold": 0.7,
    "chatMode": "chat",
    "topN": 4
  },
  "message": "Workspace created"
}
```

---

### 2. 获取所有 Workspace

**端点**: `GET /api/v1/workspaces`

**请求头**:
```
Authorization: Bearer {API_KEY}
```

**响应 (200)**:
```json
{
  "workspaces": [
    {
      "id": 79,
      "name": "Sample workspace",
      "slug": "sample-workspace",
      "createdAt": "2023-08-17T00:45:03Z",
      "openAiTemp": null,
      "lastUpdatedAt": "2023-08-17T00:45:03Z",
      "openAiHistory": 20,
      "openAiPrompt": null,
      "documents": [],
      "threads": []
    }
  ]
}
```

---

### 3. 获取 Workspace 详情

**端点**: `GET /api/v1/workspace/:slug`

**路径参数**:
- `slug` (必需): Workspace 的唯一标识符

**请求头**:
```
Authorization: Bearer {API_KEY}
```

**响应 (200)**:
```json
{
  "workspace": [
    {
      "id": 79,
      "name": "My workspace",
      "slug": "my-workspace-123",
      "createdAt": "2023-08-17T00:45:03Z",
      "openAiTemp": null,
      "lastUpdatedAt": "2023-08-17T00:45:03Z",
      "openAiHistory": 20,
      "openAiPrompt": null,
      "documents": [],
      "threads": [],
      "contextWindow": 128000,
      "currentContextTokenCount": 0
    }
  ]
}
```

---

### 4. 删除 Workspace

**端点**: `DELETE /api/v1/workspace/:slug`

**路径参数**:
- `slug` (必需): Workspace 的唯一标识符

**请求头**:
```
Authorization: Bearer {API_KEY}
```

**响应**:
- 200: 删除成功
- 400: Workspace 不存在
- 403: API Key 无效

---

### 5. 更新 Workspace 设置

**端点**: `POST /api/v1/workspace/:slug/update`

**路径参数**:
- `slug` (必需): Workspace 的唯一标识符

**请求头**:
```
Authorization: Bearer {API_KEY}
Content-Type: application/json
```

**请求体**:
```json
{
  "name": "Updated Workspace Name",
  "openAiTemp": 0.2,
  "openAiHistory": 20,
  "openAiPrompt": "Respond to all inquiries in binary",
  "similarityThreshold": 0.5,
  "topN": 5
}
```

**说明**: 所有字段都是可选的，只提供需要更新的字段。

**响应 (200)**:
```json
{
  "workspace": {
    "id": 79,
    "name": "Updated Workspace Name",
    "slug": "my-workspace",
    "openAiTemp": 0.2,
    "lastUpdatedAt": "2023-08-17T01:00:00Z",
    "openAiHistory": 20,
    "openAiPrompt": "Respond to all inquiries in binary"
  },
  "message": "Workspace updated"
}
```

---

## Document API

### 1. 上传文档

**端点**: `POST /api/v1/document/upload`

**请求头**:
```
Authorization: Bearer {API_KEY}
Content-Type: multipart/form-data
```

**请求体 (FormData)**:
- `file` (必需): 要上传的文件
- `addToWorkspaces` (可选): 逗号分隔的 workspace slug 列表
- `metadata` (可选): JSON 对象，包含文档元数据

**metadata 示例**:
```json
{
  "title": "Custom Title",
  "docAuthor": "Author Name",
  "description": "A brief description",
  "docSource": "Source of the document"
}
```

**响应 (200)**:
```json
{
  "success": true,
  "error": null,
  "documents": [
    {
      "location": "custom-documents/file.txt-uuid.json",
      "name": "file.txt-uuid.json",
      "url": "file:///path/to/file.txt",
      "title": "file.txt",
      "docAuthor": "Unknown",
      "description": "Unknown",
      "docSource": "a text file uploaded by the user.",
      "chunkSource": "file.txt",
      "published": "1/16/2024, 3:07:00 PM",
      "wordCount": 93,
      "token_count_estimate": 115
    }
  ]
}
```

---

### 2. 上传文档到指定文件夹

**端点**: `POST /api/v1/document/upload/:folderName`

**路径参数**:
- `folderName`: 目标文件夹名称

**请求体**: 与普通上传相同

**响应**: 与普通上传相同

---

### 3. 获取文档列表

**端点**: `GET /api/v1/documents`

**请求头**:
```
Authorization: Bearer {API_KEY}
```

**响应 (200)**:
```json
{
  "documents": [
    {
      "id": 1,
      "docpath": "custom-documents/file.txt-uuid.json",
      "title": "file.txt",
      "workspaceId": 79
    }
  ]
}
```

---

### 4. 删除文档

**端点**: `DELETE /api/v1/document/:id`

**路径参数**:
- `id`: 文档 ID

**请求头**:
```
Authorization: Bearer {API_KEY}
```

**响应**:
- 200: 删除成功
- 404: 文档不存在

---

## OpenAI Compatible API

### 1. 获取可用模型 (Workspaces)

**端点**: `GET /api/v1/openai/models`

**请求头**:
```
Authorization: Bearer {API_KEY}
```

**响应 (200)**:
```json
{
  "object": "list",
  "data": [
    {
      "id": "workspace-slug-1",
      "object": "model",
      "created": 1686935002,
      "owned_by": "openrouter-gpt-4o-mini"
    },
    {
      "id": "workspace-slug-2",
      "object": "model",
      "created": 1686935003,
      "owned_by": "openrouter-claude-3.5"
    }
  ]
}
```

**说明**: 每个 workspace 都是一个"模型"，使用 slug 作为 model ID。

---

### 2. 聊天 (OpenAI 兼容)

**端点**: `POST /api/v1/openai/chat/completions`

**请求头**:
```
Authorization: Bearer {API_KEY}
Content-Type: application/json
```

**请求体**:
```json
{
  "model": "workspace-slug",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant"},
    {"role": "user", "content": "What is AnythingLLM?"},
    {"role": "assistant", "content": "AnythingLLM is..."},
    {"role": "user", "content": "Follow up question..."}
  ],
  "stream": false,
  "temperature": 0.7
}
```

**参数说明**:
- `model` (必需): Workspace 的 slug
- `messages` (必需): 对话历史数组
- `stream` (可选): 是否流式响应，默认 false
- `temperature` (可选): 温度参数，默认 0.7

**响应 (200, 非流式)**:
```json
{
  "id": "uuid",
  "type": "textResponse",
  "textResponse": "AnythingLLM is a full-stack application...",
  "sources": [
    {
      "text": "Relevant context from documents...",
      "sourceId": "doc-uuid",
      "docId": "doc-uuid",
      "title": "Document Title"
    }
  ],
  "close": true,
  "error": null
}
```

**响应 (流式)**:
```
data: {"id":"uuid","type":"textResponseChunk","textChunk":"Any","sources":[],"error":null}
data: {"id":"uuid","type":"textResponseChunk","textChunk":"thing","sources":[],"error":null}
data: {"id":"uuid","type":"textResponseChunk","textChunk":"LLM","sources":[],"error":null}
data: {"id":"uuid","type":"textResponseChunk","textChunk":" is...","sources":[],"error":null,"close":true}
```

---

## Chat API

### 1. 发送消息

**端点**: `POST /api/v1/workspace/:slug/chat`

**路径参数**:
- `slug`: Workspace 的唯一标识符

**请求头**:
```
Authorization: Bearer {API_KEY}
Content-Type: application/json
```

**请求体**:
```json
{
  "message": "What is AnythingLLM?",
  "mode": "chat",
  "sessionId": "optional-session-id"
}
```

**参数说明**:
- `message` (必需): 用户消息
- `mode` (可选): 聊天模式，`chat` 或 `query`，默认 `chat`
  - `chat`: 多轮对话，包含历史上下文
  - `query`: 单次查询，仅基于文档
- `sessionId` (可选): 会话 ID，用于保持对话历史

**响应 (200)**:
```json
{
  "id": "uuid",
  "type": "textResponse",
  "textResponse": "AnythingLLM is a full-stack application...",
  "sources": [
    {
      "text": "Relevant context...",
      "sourceId": "doc-uuid",
      "docId": "doc-uuid",
      "title": "Document Title"
    }
  ],
  "close": true,
  "error": null
}
```

---

### 2. 流式聊天

**端点**: `POST /api/v1/workspace/:slug/stream-chat`

**路径参数**:
- `slug`: Workspace 的唯一标识符

**请求头**:
```
Authorization: Bearer {API_KEY}
Content-Type: application/json
```

**请求体**: 与普通聊天相同

**响应 (流式 SSE)**:
```
data: {"type":"textResponseChunk","textChunk":"Any","sources":[]}
data: {"type":"textResponseChunk","textChunk":"thing","sources":[]}
data: {"type":"textResponseChunk","textChunk":"LLM","sources":[]}
data: {"type":"textResponseChunk","textChunk":" is...","sources":[],"close":true}
```

---

### 3. 获取聊天历史

**端点**: `GET /api/v1/workspace/:slug/chats`

**路径参数**:
- `slug`: Workspace 的唯一标识符

**查询参数**:
- `limit` (可选): 返回数量限制，默认 20

**请求头**:
```
Authorization: Bearer {API_KEY}
```

**响应 (200)**:
```json
{
  "chats": [
    {
      "id": 1,
      "workspaceId": 79,
      "user_id": "user-uuid",
      "prompt": "What is AnythingLLM?",
      "response": "AnythingLLM is...",
      "createdAt": "2023-08-17T00:45:03Z"
    }
  ]
}
```

---

### 4. 删除聊天历史

**端点**: `DELETE /api/v1/workspace/:slug/chats`

**路径参数**:
- `slug`: Workspace 的唯一标识符

**请求体** (可选):
```json
{
  "chatId": 123
}
```

**说明**:
- 不提供 `chatId`: 删除所有聊天历史
- 提供 `chatId`: 删除指定聊天

**响应**:
- 200: 删除成功

---

## 错误响应

### 通用错误格式

**403 Forbidden - 无效的 API Key**:
```json
{
  "error": "Invalid API Key"
}
```

**400 Bad Request**:
```json
{
  "error": "Bad request parameters"
}
```

**404 Not Found**:
```json
{
  "error": "Resource not found"
}
```

**500 Internal Server Error**:
```json
{
  "error": "Internal server error"
}
```

---

## 使用示例

### 1. 完整工作流

```bash
# 1. 创建 Workspace
curl -X POST "http://localhost:3001/api/v1/workspace/new" \
  -H "Authorization: Bearer sk-xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My AI Assistant",
    "slug": "my-ai-assistant",
    "openAiPrompt": "You are a helpful AI assistant"
  }'

# 2. 上传文档
curl -X POST "http://localhost:3001/api/v1/document/upload" \
  -H "Authorization: Bearer sk-xxx" \
  -F "file=@document.pdf" \
  -F "addToWorkspaces=my-ai-assistant"

# 3. 发送消息
curl -X POST "http://localhost:3001/api/v1/workspace/my-ai-assistant/chat" \
  -H "Authorization: Bearer sk-xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What does the document say?",
    "mode": "chat"
  }'
```

---

## 注意事项

1. **API Key**: 所有请求都需要有效的 API Key
2. **Slug 唯一性**: Workspace slug 必须唯一
3. **文档处理**: 上传后文档需要时间处理，处理完成后才能用于 RAG
4. **聊天模式**: 
   - `chat`: 使用对话历史 + 文档检索
   - `query`: 仅使用文档检索
5. **流式响应**: 使用 SSE (Server-Sent Events) 格式

---

**文档结束**
