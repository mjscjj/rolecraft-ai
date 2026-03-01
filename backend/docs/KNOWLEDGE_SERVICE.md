# 知识库服务重构文档

## 概述

本次重构将知识库服务从本地文件存储升级为集成 AnythingLLM 文档管理和向量检索能力。

## 架构变化

### 旧架构
```
POST /api/v1/documents
  → 保存到本地 ./uploads/{uuid}.pdf
  → 记录到数据库
```

### 新架构
```
POST /api/v1/documents
  → 临时保存到本地
  → 调用 AnythingLLM /v1/document/upload
  → 调用 /v1/workspace/{slug}/update-embeddings
  → 更新本地 Document 状态为 completed
  → 返回结果
```

## 新增功能

### 1. 异步文档处理

文档上传后，系统会：
1. 立即返回 `processing` 状态
2. 后台异步上传到 AnythingLLM
3. 自动更新 embeddings
4. 最终更新状态为 `completed` 或 `failed`

```go
// 异步处理流程
go func() {
    // 1. 上传到 AnythingLLM
    anythingLLMFileId, err := h.uploadToAnythingLLM(tempFilePath, userId)
    
    // 2. 更新 embeddings
    err = h.updateEmbeddings(userId)
    
    // 3. 更新文档状态
    h.updateDocumentStatusWithMetadata(docId, "completed", finalFilePath, metadata)
}()
```

### 2. 向量搜索端点

新增 `POST /api/v1/documents/search` 端点：

**请求示例：**
```bash
curl -X POST "http://localhost:8080/api/v1/documents/search" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "搜索关键词",
    "topN": 4
  }'
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "query": "搜索关键词",
    "results": [
      {
        "text": "相关文档片段内容...",
        "score": 0.95,
        "source": "文档名称.pdf"
      }
    ]
  }
}
```

### 3. 文档删除增强

删除文档时会：
1. 从 AnythingLLM 删除文档
2. 删除本地元数据
3. 清理临时文件

## 配置

### 环境变量

在 `.env` 文件中添加以下配置：

```bash
# AnythingLLM 配置
ANYTHINGLLM_BASE_URL=http://43.134.234.4:3001/api/v1
ANYTHINGLLM_API_KEY=sk-your-api-key-here
ANYTHINGLLM_WORKSPACE=user_001
```

### 默认值

如果未配置，系统使用以下默认值：
- `ANYTHINGLLM_BASE_URL`: `http://43.134.234.4:3001/api/v1`
- `ANYTHINGLLM_WORKSPACE`: `user_001`
- `ANYTHINGLLM_API_KEY`: 必须配置，否则 API 调用会失败

## API 端点

### 上传文档
```
POST /api/v1/documents
Content-Type: multipart/form-data

Parameters:
- file: 文件 (支持 .pdf, .doc, .docx, .txt, .md)

Response:
{
  "code": 200,
  "message": "document uploaded and processing",
  "data": {
    "id": "uuid",
    "name": "filename.pdf",
    "status": "processing",
    ...
  }
}
```

### 向量搜索
```
POST /api/v1/documents/search
Content-Type: application/json

Request Body:
{
  "query": "搜索内容",
  "topN": 4  // 可选，默认 4，最大 20
}

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "query": "搜索内容",
    "results": [...]
  }
}
```

### 获取文档状态
```
GET /api/v1/documents/:id

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "status": "completed",  // processing | completed | failed
    "metadata": {
      "anythingLLMFileId": "xxx"
    },
    ...
  }
}
```

### 删除文档
```
DELETE /api/v1/documents/:id

Response:
{
  "code": 200,
  "message": "success"
}
```

## 测试

### 运行测试

```bash
cd backend
go test -v ./tests -run TestDocumentFlow
```

### 手动测试

1. **测试文档上传流程：**
```bash
# 1. 登录获取 token
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 2. 上传文档
curl -X POST "http://localhost:8080/api/v1/documents" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test.pdf"

# 3. 检查状态
curl -X GET "http://localhost:8080/api/v1/documents/DOCUMENT_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

2. **测试向量搜索：**
```bash
curl -X POST "http://localhost:8080/api/v1/documents/search" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query":"关键词","topN":4}'
```

3. **测试异步处理：**
```bash
# 上传文档后立即查询状态，观察状态变化
# processing → completed (或 failed)
```

## 错误处理

### 常见错误

1. **API Key 未配置**
   - 错误：`anythingLLM upload failed: unauthorized`
   - 解决：设置 `ANYTHINGLLM_API_KEY` 环境变量

2. **AnythingLLM 服务不可达**
   - 错误：`failed to send request: connection refused`
   - 解决：检查 `ANYTHINGLLM_BASE_URL` 配置和网络连接

3. **文档处理失败**
   - 状态：`failed`
   - 查看 `error_message` 字段获取详细错误信息

## 监控和日志

### 文档状态流转
```
pending → processing → completed
                      ↘ failed
```

### 查询文档状态
```bash
# 获取所有处理中的文档
curl "http://localhost:8080/api/v1/documents?status=processing"

# 获取失败的文档
curl "http://localhost:8080/api/v1/documents?status=failed"
```

## 性能优化建议

1. **批量上传**：避免同时上传大量文档，建议限制并发数
2. **超时设置**：大文档上传已设置 5 分钟超时，embedding 更新设置 10 分钟超时
3. **重试机制**：可在 `processDocumentAsync` 中添加重试逻辑
4. **队列处理**：生产环境建议使用消息队列处理异步任务

## 未来改进

- [ ] 添加文档处理进度查询
- [ ] 支持批量文档上传
- [ ] 添加文档处理失败重试机制
- [ ] 支持多个工作空间管理
- [ ] 添加文档版本控制

## 参考

- [AnythingLLM API 文档](http://43.134.234.4:3001/api/v1)
- [AnythingLLM GitHub](https://github.com/Mintplex-Labs/anything-llm)
