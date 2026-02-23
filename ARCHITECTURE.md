# RoleCraft AI - 技术架构设计

## 1. 系统架构概览

```
┌─────────────────────────────────────────────────────────────────────┐
│                          客户端层                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │   Web App    │  │  Mobile App  │  │   第三方集成  │              │
│  │  React/TS    │  │  React Native│  │   API调用    │              │
│  └──────────────┘  └──────────────┘  └──────────────┘              │
└────────────────────────┬────────────────────────────────────────────┘
                         │ HTTPS/WSS
┌────────────────────────▼────────────────────────────────────────────┐
│                         网关层                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                     API Gateway (Kong/Nginx)                  │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐             │  │
│  │  │  路由分发   │ │  限流控制   │ │  认证中间件  │             │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘             │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────────────┐
│                         服务层                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                   RoleCraft AI 后端服务                        │  │
│  │                         (Go/Gin)                              │  │
│  ├──────────────────────────────────────────────────────────────┤  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │  │
│  │  │用户服务  │ │角色服务  │ │对话服务  │ │知识库服务│        │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │  │
│  │  │技能服务  │ │API服务   │ │计费服务  │ │通知服务  │        │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
┌───────▼───────┐ ┌──────▼────────┐ ┌─────▼──────────┐
│    数据层      │ │    缓存层      │ │   文件存储层   │
│  PostgreSQL   │ │     Redis     │ │    MinIO       │
│  (主数据库)    │ │  (会话/缓存)   │ │  (对象存储)    │
└───────────────┘ └───────────────┘ └────────────────┘
        │                │                │
┌───────▼───────┐ ┌──────▼────────┐ ┌─────▼──────────┐
│   向量数据库   │ │   消息队列     │ │   AI 服务层    │
│   Milvus      │ │    RabbitMQ   │ │  OpenAI/Claude │
│  (语义检索)    │ │  (异步任务)    │ │  Embedding    │
└───────────────┘ └───────────────┘ └────────────────┘
```

---

## 2. 技术栈选型

### 2.1 前端技术栈

| 类别 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 框架 | React | 18.x | 组件化 UI 框架 |
| 语言 | TypeScript | 5.x | 类型安全 |
| 构建 | Vite | 5.x | 快速构建工具 |
| 路由 | React Router | 6.x | 客户端路由 |
| 状态 | Zustand | 4.x | 轻量级状态管理 |
| 数据获取 | TanStack Query | 5.x | 服务端状态管理 |
| UI 组件 | shadcn/ui | latest | 基础组件库 |
| 样式 | Tailwind CSS | 3.x | 原子化 CSS |
| 图标 | Lucide React | latest | 图标库 |
| 编辑器 | Monaco Editor | latest | 代码/文本编辑器 |
| 图表 | Recharts | 2.x | 数据可视化 |

### 2.2 后端技术栈

| 类别 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 语言 | Go | 1.21+ | 高性能后端语言 |
| Web 框架 | Gin | 1.9+ | HTTP Web 框架 |
| ORM | GORM | 1.25+ | 数据库 ORM |
| 验证 | go-playground/validator | latest | 请求验证 |
| JWT | golang-jwt/jwt | 5.x | 认证授权 |
| 配置 | Viper | latest | 配置管理 |
| 日志 | Zap | latest | 高性能日志 |
| 文档 | Swaggo | latest | API 文档生成 |
| 测试 | Testify | latest | 单元测试 |

### 2.3 基础设施

| 类别 | 技术 | 说明 |
|------|------|------|
| 数据库 | PostgreSQL 15+ | 主数据库 |
| 缓存 | Redis 7+ | 会话、缓存、限流 |
| 向量 DB | Milvus / Pinecone | 语义检索 |
| 消息队列 | RabbitMQ | 异步任务 |
| 对象存储 | MinIO / AWS S3 | 文件存储 |
| 搜索 | Elasticsearch | 全文检索（可选）|
| 监控 | Prometheus + Grafana | 指标监控 |
| 日志 | ELK Stack | 日志收集分析 |
| 网关 | Kong / Nginx | API 网关 |
| 容器 | Docker + K8s | 容器编排 |

---

## 3. 项目结构

```
rolecraft-ai/
├── frontend/                          # 前端项目
│   ├── public/
│   ├── src/
│   │   ├── api/                       # API 请求封装
│   │   │   ├── client.ts              # Axios 实例
│   │   │   ├── auth.ts                # 认证相关
│   │   │   ├── roles.ts               # 角色管理
│   │   │   ├── documents.ts           # 文档管理
│   │   │   └── chat.ts                # 对话相关
│   │   ├── components/                # 组件
│   │   │   ├── ui/                    # 基础 UI 组件
│   │   │   ├── layout/                # 布局组件
│   │   │   ├── role/                  # 角色相关组件
│   │   │   ├── chat/                  # 对话相关组件
│   │   │   └── document/              # 文档相关组件
│   │   ├── hooks/                     # 自定义 Hooks
│   │   ├── pages/                     # 页面组件
│   │   │   ├── auth/                  # 认证页面
│   │   │   ├── dashboard/             # 仪表板
│   │   │   ├── roles/                 # 角色管理
│   │   │   ├── documents/             # 知识库
│   │   │   ├── chat/                  # 对话界面
│   │   │   └── settings/              # 设置页面
│   │   ├── stores/                    # 状态管理
│   │   ├── types/                     # TypeScript 类型
│   │   ├── utils/                     # 工具函数
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   ├── tailwind.config.js
│   └── vite.config.ts
│
├── backend/                           # 后端项目
│   ├── cmd/
│   │   └── server/                    # 应用入口
│   │       └── main.go
│   ├── internal/
│   │   ├── api/                       # API 层
│   │   │   ├── handler/               # 请求处理器
│   │   │   │   ├── auth.go
│   │   │   │   ├── user.go
│   │   │   │   ├── workspace.go
│   │   │   │   ├── role.go
│   │   │   │   ├── document.go
│   │   │   │   └── chat.go
│   │   │   ├── middleware/            # 中间件
│   │   │   │   ├── auth.go
│   │   │   │   ├── cors.go
│   │   │   │   └── rate_limit.go
│   │   │   └── router.go              # 路由配置
│   │   ├── config/                    # 配置
│   │   │   └── config.go
│   │   ├── models/                    # 数据模型
│   │   │   ├── user.go
│   │   │   ├── workspace.go
│   │   │   ├── role.go
│   │   │   ├── document.go
│   │   │   └── chat.go
│   │   ├── repository/                # 数据访问层
│   │   │   ├── user_repo.go
│   │   │   ├── role_repo.go
│   │   │   └── ...
│   │   ├── service/                   # 业务逻辑层
│   │   │   ├── auth_service.go
│   │   │   ├── role_service.go
│   │   │   ├── document_service.go
│   │   │   ├── chat_service.go
│   │   │   └── ai/                    # AI 相关服务
│   │   │       ├── openai.go
│   │   │       ├── embedding.go
│   │   │       └── rag.go
│   │   ├── pkg/                       # 公共包
│   │   │   ├── utils/
│   │   │   ├── errors/
│   │   │   └── logger/
│   │   └── database/                  # 数据库
│   │       ├── postgres.go
│   │       ├── redis.go
│   │       └── migrate.go
│   ├── pkg/                           # 外部可引用包
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── docker-compose.yml                 # 本地开发环境
├── Makefile                           # 构建脚本
└── README.md
```

---

## 4. 核心功能实现

### 4.1 RAG (检索增强生成) 流程

```
用户提问
    │
    ▼
┌─────────────┐
│  查询理解    │
│  Query理解   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  向量检索    │──────► 向量数据库 (Milvus)
│  Embedding  │        相似度 Top-K
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  重排序     │
│  Rerank    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  上下文构建  │
│  Prompt构建  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  LLM 生成   │──────► OpenAI/Claude API
│  流式输出   │
└──────┬──────┘
       │
       ▼
    返回结果
```

### 4.2 文档处理流程

```
上传文档
    │
    ▼
┌─────────────┐
│  格式检测    │
│  PDF/DOCX/TXT│
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  文本提取    │
│  OCR(可选)   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  文本清洗    │
│  格式规范化  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  智能分块    │
│  Chunking   │
│  语义边界    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  向量化     │
│  Embedding  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  存储到     │
│  向量数据库  │
└─────────────┘
```

### 4.3 流式对话实现

**前端实现 (EventSource):**
```typescript
const sendMessage = async (content: string) => {
  const response = await fetch('/api/v1/chat/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message: content }),
  });

  const reader = response.body?.getReader();
  const decoder = new TextDecoder();

  while (true) {
    const { done, value } = await reader!.read();
    if (done) break;

    const chunk = decoder.decode(value);
    // 解析 SSE 格式数据并更新 UI
    appendToMessage(chunk);
  }
};
```

**后端实现 (SSE):**
```go
func (h *ChatHandler) StreamChat(c *gin.Context) {
    ctx := c.Request.Context()

    // 设置 SSE 头
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // 获取流式响应
    stream, err := h.chatService.StreamResponse(ctx, req)
    if err != nil {
        c.SSEvent("error", err.Error())
        return
    }

    // 转发流
    for chunk := range stream {
        c.SSEvent("message", chunk)
        c.Writer.Flush()
    }
}
```

---

## 5. 部署架构

### 5.1 生产环境部署

```
                    ┌─────────────┐
                    │   CDN       │
                    │  (静态资源)  │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  Nginx      │
                    │  (负载均衡)  │
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────▼────┐        ┌────▼────┐        ┌────▼────┐
   │Frontend │        │Backend  │        │Backend  │
   │  x2     │        │  x3     │        │  x3     │
   └─────────┘        └────┬────┘        └─────────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
         ┌────▼────┐  ┌────▼────┐  ┌────▼────┐
         │PostgreSQL│  │  Redis  │  │ Milvus  │
         │ Primary │  │ Cluster │  │  Cluster│
         └────┬────┘  └─────────┘  └─────────┘
              │
         ┌────▼────┐
         │  Replica │
         └─────────┘
```

### 5.2 Docker Compose 配置

```yaml
version: '3.8'

services:
  frontend:
    build: ./frontend
    ports:
      - "80:80"
    depends_on:
      - backend

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - MILVUS_HOST=milvus
    depends_on:
      - postgres
      - redis
      - milvus

  postgres:
    image: postgres:15-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=rolecraft
      - POSTGRES_USER=rolecraft
      - POSTGRES_PASSWORD=secret

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  milvus:
    image: milvusdb/milvus:latest
    volumes:
      - milvus_data:/var/lib/milvus

  minio:
    image: minio/minio:latest
    command: server /data
    volumes:
      - minio_data:/data

volumes:
  postgres_data:
  redis_data:
  milvus_data:
  minio_data:
```

---

## 6. 安全设计

### 6.1 认证授权
- JWT Token 认证
- Refresh Token 机制
- Token 黑名单（Redis）
- 接口权限控制

### 6.2 数据安全
- 密码 bcrypt 加密
- 敏感字段加密存储
- HTTPS 传输
- API 限流保护

### 6.3 文件安全
- 文件类型白名单
- 文件大小限制
- 病毒扫描（ClamAV）
- 沙箱处理

---

## 7. 监控与日志

### 7.1 监控指标
- QPS / 响应时间
- 错误率
- AI Token 消耗
- 用户活跃度

### 7.2 日志规范
```
[INFO]  2024-01-20 10:00:00 | request_id: xxx | user_id: xxx | method: POST | path: /api/v1/chat | duration: 1.2s
[ERROR] 2024-01-20 10:00:01 | request_id: xxx | error: failed to connect to AI service
```
