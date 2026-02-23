# RoleCraft AI - Backend

## 项目结构

```
backend/
├── cmd/
│   ├── server/main.go      # 服务入口
│   └── migrate/main.go     # 数据库迁移
├── internal/
│   ├── api/
│   │   ├── handler/        # API 处理器
│   │   │   ├── auth.go     # 认证
│   │   │   ├── user.go     # 用户
│   │   │   ├── role.go     # 角色
│   │   │   ├── document.go # 文档
│   │   │   └── chat.go     # 对话
│   │   └── middleware/     # 中间件
│   ├── config/             # 配置
│   ├── database/           # 数据库
│   ├── models/             # 数据模型
│   └── service/            # 业务服务
│       ├── ai/             # AI 服务
│       │   ├── openai.go   # OpenAI 客户端
│       │   ├── embedding.go # 向量化
│       │   └── rag.go      # RAG 检索
│       └── document/       # 文档处理
│           └── processor.go
├── go.mod
└── go.sum
```

## API 端点

### 认证
- `POST /api/v1/auth/register` - 注册
- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/refresh` - 刷新 Token

### 用户
- `GET /api/v1/users/me` - 获取当前用户
- `PUT /api/v1/users/me` - 更新用户信息

### 角色
- `GET /api/v1/roles` - 获取角色列表
- `GET /api/v1/roles/templates` - 获取模板
- `GET /api/v1/roles/:id` - 获取单个角色
- `POST /api/v1/roles` - 创建角色
- `PUT /api/v1/roles/:id` - 更新角色
- `DELETE /api/v1/roles/:id` - 删除角色
- `POST /api/v1/roles/:id/chat` - 与角色对话

### 文档
- `GET /api/v1/documents` - 文档列表
- `POST /api/v1/documents` - 上传文档
- `GET /api/v1/documents/:id` - 文档详情
- `DELETE /api/v1/documents/:id` - 删除文档

### 对话
- `GET /api/v1/chat-sessions` - 会话列表
- `POST /api/v1/chat-sessions` - 创建会话
- `GET /api/v1/chat-sessions/:id` - 会话详情
- `POST /api/v1/chat/:id/complete` - 发送消息
- `POST /api/v1/chat/:id/stream` - 流式消息

## 运行

```bash
# 开发模式
go run cmd/server/main.go

# 构建
go build -o bin/server cmd/server/main.go

# 数据库迁移
go run cmd/migrate/main.go up
go run cmd/migrate/main.go seed
```
