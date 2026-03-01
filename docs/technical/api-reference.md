# RoleCraft AI - API 参考文档

> 完整的 RESTful API 接口文档

---

## 目录

1. [概述](#1-概述)
2. [认证](#2-认证)
3. [用户 API](#3-用户-api)
4. [角色 API](#4-角色-api)
5. [对话 API](#5-对话-api)
6. [文档 API](#6-文档-api)
7. [分析 API](#7-分析-api)
8. [公司与工作区 API（MVP）](#8-公司与工作区-apimvp)
9. [错误处理](#9-错误处理)

---

## 1. 概述

### 1.1 API 地址

**生产环境：**
```
https://api.rolecraft.ai/v1
```

**测试环境：**
```
https://api-test.rolecraft.ai/v1
```

**本地开发：**
```
http://localhost:8080/api/v1
```

### 1.2 请求格式

- **Content-Type:** `application/json`
- **字符编码:** `UTF-8`
- **数据格式:** JSON

### 1.3 响应格式

所有响应均为 JSON 格式：

```json
{
  "success": true,
  "data": { ... },
  "message": "操作成功",
  "timestamp": "2026-02-27T10:00:00Z"
}
```

### 1.4 分页参数

列表接口支持分页：

```
GET /api/v1/roles?page=1&page_size=20
```

**响应包含：**
```json
{
  "success": true,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  }
}
```

---

## 2. 认证

### 2.1 认证方式

使用 Bearer Token 认证：

```
Authorization: Bearer <your_access_token>
```

### 2.2 获取 Token

**登录接口：**

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "your_password"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 7200,
    "token_type": "Bearer"
  }
}
```

### 2.3 Token 刷新

Token 过期前使用 refresh_token 刷新：

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 7200
  }
}
```

### 2.4 API 密钥认证

服务端对服务端调用使用 API 密钥：

```
Authorization: Bearer <api_key>
```

**创建 API 密钥：**
1. 登录控制台
2. 进入 API 平台
3. 创建新密钥
4. 复制并保存（只显示一次）

---

## 3. 用户 API

### 3.1 获取当前用户信息

**请求：**
```http
GET /api/v1/users/me
Authorization: Bearer <token>
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "user_123",
    "email": "user@example.com",
    "name": "张三",
    "avatar": "https://...",
    "workspace_id": "ws_456",
    "created_at": "2026-01-01T00:00:00Z",
    "subscription": {
      "plan": "professional",
      "status": "active",
      "expires_at": "2026-12-31T23:59:59Z"
    }
  }
}
```

### 3.2 更新用户信息

**请求：**
```http
PUT /api/v1/users/me
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "李四",
  "avatar": "https://..."
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "user_123",
    "name": "李四",
    ...
  },
  "message": "更新成功"
}
```

---

## 4. 角色 API

### 4.1 获取角色列表

**请求：**
```http
GET /api/v1/roles
Authorization: Bearer <token>
```

**查询参数：**
- `page` - 页码（默认 1）
- `page_size` - 每页数量（默认 20，最大 100）
- `category` - 分类筛选
- `search` - 关键词搜索

**响应：**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "role_123",
        "name": "营销专家",
        "description": "专业的营销策划助手",
        "avatar": "https://...",
        "category": "营销",
        "is_template": false,
        "is_public": false,
        "created_at": "2026-02-01T10:00:00Z",
        "updated_at": "2026-02-27T10:00:00Z"
      }
    ],
    "total": 15,
    "page": 1,
    "page_size": 20
  }
}
```

### 4.2 获取角色详情

**请求：**
```http
GET /api/v1/roles/:id
Authorization: Bearer <token>
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "role_123",
    "name": "营销专家",
    "description": "专业的营销策划助手",
    "avatar": "https://...",
    "category": "营销",
    "system_prompt": "你是一位经验丰富的营销专家...",
    "welcome_message": "你好！我是你的营销助手👋",
    "model_config": {
      "model": "gpt-4",
      "temperature": 0.7,
      "max_tokens": 1000,
      "top_p": 0.9
    },
    "skills": ["web_search", "file_processing"],
    "knowledge_bases": ["kb_123", "kb_456"],
    "is_template": false,
    "is_public": false,
    "created_at": "2026-02-01T10:00:00Z",
    "updated_at": "2026-02-27T10:00:00Z"
  }
}
```

### 4.3 创建角色

**请求：**
```http
POST /api/v1/roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "我的写作助手",
  "description": "帮助撰写和优化文章",
  "avatar": "https://...",
  "category": "通用",
  "system_prompt": "你是一位专业的写作助手...",
  "welcome_message": "你好！我是你的写作助手",
  "model_config": {
    "model": "gpt-4",
    "temperature": 0.8,
    "max_tokens": 1500
  },
  "skills": ["file_processing"],
  "knowledge_bases": [],
  "is_public": false
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "role_789",
    "name": "我的写作助手",
    ...
  },
  "message": "角色创建成功"
}
```

### 4.4 更新角色

**请求：**
```http
PUT /api/v1/roles/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "高级写作助手",
  "system_prompt": "你是一位资深写作专家...",
  "model_config": {
    "temperature": 0.9
  }
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "role_789",
    "name": "高级写作助手",
    ...
  },
  "message": "角色更新成功"
}
```

### 4.5 删除角色

**请求：**
```http
DELETE /api/v1/roles/:id
Authorization: Bearer <token>
```

**响应：**
```json
{
  "success": true,
  "message": "角色删除成功"
}
```

### 4.6 获取角色模板

**请求：**
```http
GET /api/v1/roles/templates
Authorization: Bearer <token>
```

**响应：**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "template_1",
        "name": "智能助理",
        "description": "全能型办公助手",
        "category": "通用",
        "avatar": "🤖",
        "system_prompt": "你是一位智能办公助手...",
        "welcome_message": "你好！我是你的智能助理"
      }
    ]
  }
}
```

---

## 5. 对话 API

### 5.1 创建对话会话

**请求：**
```http
POST /api/v1/chat-sessions
Authorization: Bearer <token>
Content-Type: application/json

{
  "role_id": "role_123",
  "title": "产品推广方案讨论",
  "mode": "task"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "session_456",
    "role_id": "role_123",
    "title": "产品推广方案讨论",
    "mode": "task",
    "created_at": "2026-02-27T10:00:00Z"
  }
}
```

### 5.2 获取会话列表

**请求：**
```http
GET /api/v1/chat-sessions
Authorization: Bearer <token>
```

**查询参数：**
- `page` - 页码
- `page_size` - 每页数量
- `role_id` - 按角色筛选

**响应：**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "session_456",
        "role_id": "role_123",
        "role_name": "营销专家",
        "title": "产品推广方案讨论",
        "mode": "task",
        "last_message": "好的，让我为你制定一份推广方案...",
        "updated_at": "2026-02-27T10:30:00Z"
      }
    ],
    "total": 25
  }
}
```

### 5.3 获取会话详情

**请求：**
```http
GET /api/v1/chat-sessions/:id
Authorization: Bearer <token>
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "session_456",
    "role_id": "role_123",
    "title": "产品推广方案讨论",
    "mode": "task",
    "messages": [
      {
        "id": "msg_1",
        "role": "user",
        "content": "我需要写一份产品推广方案",
        "created_at": "2026-02-27T10:00:00Z"
      },
      {
        "id": "msg_2",
        "role": "assistant",
        "content": "你好！很高兴帮你制定产品推广方案...",
        "sources": [],
        "created_at": "2026-02-27T10:00:05Z"
      }
    ],
    "created_at": "2026-02-27T10:00:00Z",
    "updated_at": "2026-02-27T10:30:00Z"
  }
}
```

### 5.4 发送消息（普通）

**请求：**
```http
POST /api/v1/chat/:id/complete
Authorization: Bearer <token>
Content-Type: application/json

{
  "message": "我的产品是一款智能手表，目标用户是 25-35 岁的都市白领",
  "mode": "task"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "message_id": "msg_3",
    "content": "非常好！针对智能手表产品，我为你制定以下推广方案...\n\n## 目标用户分析\n25-35 岁都市白领具有以下特点：...",
    "role_id": "role_123",
    "session_id": "session_456",
    "usage": {
      "tokens": 350,
      "cost": 0.005
    },
    "created_at": "2026-02-27T10:05:00Z"
  }
}
```

### 5.5 发送消息（流式）

**请求：**
```http
POST /api/v1/chat/:id/stream
Authorization: Bearer <token>
Content-Type: application/json
Accept: text/event-stream

{
  "message": "我的产品是一款智能手表...",
  "mode": "task"
}
```

**响应：** SSE (Server-Sent Events)

```
data: {"content": "非常好！针"}
data: {"content": "对智能手表"}
data: {"content": "产品，我为你"}
data: {"content": "制定以下推广方案..."}
data: {"done": true, "usage": {"tokens": 350}}
```

### 5.6 删除会话

**请求：**
```http
DELETE /api/v1/chat-sessions/:id
Authorization: Bearer <token>
```

**响应：**
```json
{
  "success": true,
  "message": "会话删除成功"
}
```

---

## 6. 文档 API

### 6.1 获取文档列表

**请求：**
```http
GET /api/v1/documents
Authorization: Bearer <token>
```

**响应：**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "doc_123",
        "name": "产品手册.pdf",
        "file_type": "pdf",
        "file_size": 2048000,
        "status": "completed",
        "chunk_count": 25,
        "created_at": "2026-02-20T10:00:00Z"
      }
    ],
    "total": 10
  }
}
```

### 6.2 上传文档

**请求：**
```http
POST /api/v1/documents
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <binary>
name: "产品手册"
description: "产品功能说明手册"
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "doc_123",
    "name": "产品手册",
    "file_type": "pdf",
    "file_size": 2048000,
    "status": "processing",
    "message": "文档上传成功，正在处理中"
  }
}
```

### 6.3 获取文档详情

**请求：**
```http
GET /api/v1/documents/:id
Authorization: Bearer <token>
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "doc_123",
    "name": "产品手册",
    "file_type": "pdf",
    "file_size": 2048000,
    "status": "completed",
    "chunk_count": 25,
    "metadata": {
      "pages": 50,
      "words": 15000
    },
    "created_at": "2026-02-20T10:00:00Z",
    "completed_at": "2026-02-20T10:05:00Z"
  }
}
```

### 6.4 删除文档

**请求：**
```http
DELETE /api/v1/documents/:id
Authorization: Bearer <token>
```

**响应：**
```json
{
  "success": true,
  "message": "文档删除成功"
}
```

---

## 7. 分析 API

### 7.1 获取使用统计

**请求：**
```http
GET /api/v1/analytics/usage
Authorization: Bearer <token>
```

**查询参数：**
- `start_date` - 开始日期（YYYY-MM-DD）
- `end_date` - 结束日期（YYYY-MM-DD）
- `granularity` - 粒度（day/week/month）

**响应：**
```json
{
  "success": true,
  "data": {
    "total_requests": 1500,
    "total_tokens": 250000,
    "total_cost": 5.50,
    "by_date": [
      {
        "date": "2026-02-27",
        "requests": 150,
        "tokens": 25000,
        "cost": 0.55
      }
    ],
    "by_role": [
      {
        "role_id": "role_123",
        "role_name": "营销专家",
        "requests": 500,
        "tokens": 80000
      }
    ]
  }
}
```

### 7.2 获取对话统计

**请求：**
```http
GET /api/v1/analytics/conversations
Authorization: Bearer <token>
```

**响应：**
```json
{
  "success": true,
  "data": {
    "total_conversations": 200,
    "active_conversations": 50,
    "avg_messages_per_session": 15,
    "top_roles": [
      {
        "role_id": "role_123",
        "role_name": "营销专家",
        "conversations": 80
      }
    ]
  }
}
```

---

## 8. 公司与工作区 API（MVP）

> 2026-03 新增，支持“工作区”异步任务和公司成果交付聚合。

### 8.1 获取公司列表

```http
GET /api/v1/companies
Authorization: Bearer <token>
```

### 8.2 获取公司详情（含聚合成果）

```http
GET /api/v1/companies/:id
Authorization: Bearer <token>
```

响应中包含：

- `stats.workspaceCount`
- `stats.outcomeCount`
- `recentOutcomes[]`

### 8.3 工作区任务列表

```http
GET /api/v1/workspaces?companyId=&status=&triggerType=&asyncStatus=
Authorization: Bearer <token>
```

### 8.4 创建工作区任务

```http
POST /api/v1/workspaces
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "每天 09:00 生成运营汇报",
  "companyId": "xxx",
  "type": "report",
  "triggerType": "daily",
  "triggerValue": "09:00",
  "timezone": "Asia/Shanghai",
  "reportRule": "汇总昨日数据并输出摘要"
}
```

### 8.5 更新工作区任务

```http
PUT /api/v1/workspaces/:id
Authorization: Bearer <token>
Content-Type: application/json
```

### 8.6 立即执行任务（MVP）

```http
POST /api/v1/workspaces/:id/run
Authorization: Bearer <token>
```

### 8.7 兼容旧路径

- `/api/v1/works` 与 `/api/v1/workspaces` 等价（向后兼容）

---

## 9. 错误处理

### 8.1 错误响应格式

```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "请求参数无效",
    "details": {
      "field": "email",
      "reason": "邮箱格式不正确"
    }
  },
  "timestamp": "2026-02-27T10:00:00Z"
}
```

### 8.2 错误码列表

| 错误码 | HTTP 状态码 | 说明 |
|-------|-----------|------|
| `INVALID_REQUEST` | 400 | 请求参数错误 |
| `UNAUTHORIZED` | 401 | 认证失败 |
| `FORBIDDEN` | 403 | 权限不足 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `RATE_LIMITED` | 429 | 超出速率限制 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |
| `SERVICE_UNAVAILABLE` | 503 | 服务不可用 |

### 8.3 重试策略

**建议重试场景：**
- 5xx 服务器错误
- 网络超时
- 429 速率限制（需等待）

**指数退避：**
```python
import time

def retry_request(func, max_retries=3):
    for i in range(max_retries):
        try:
            return func()
        except ServerError:
            if i == max_retries - 1:
                raise
            wait_time = (2 ** i) + random.random()
            time.sleep(wait_time)
```

---

## 📚 相关文档

- [部署指南](./deployment-guide.md)
- [开发环境配置](./dev-setup.md)
- [数据库设计文档](./database-design.md)

---

*最后更新：2026-02-27*
