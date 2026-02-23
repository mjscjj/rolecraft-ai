# RoleCraft AI - 快速启动指南

## 环境准备

### 1. 安装依赖
```bash
# 后端依赖
cd backend
go mod download

# 前端依赖
cd ../frontend
pnpm install
```

### 2. 配置环境变量
```bash
cp backend/.env.example backend/.env
# 编辑 .env 文件，填入你的 OpenAI API Key
```

### 3. 启动基础设施
```bash
# 启动 PostgreSQL, Redis, Milvus, MinIO
docker-compose up -d postgres redis minio

# 等待服务启动 (约 30 秒)
```

### 4. 初始化数据库
```bash
cd backend
go run cmd/migrate/main.go up
go run cmd/migrate/main.go seed
```

## 启动服务

### 开发模式
```bash
# 终端 1 - 后端
cd backend
go run cmd/server/main.go

# 终端 2 - 前端
cd frontend
pnpm dev
```

### 使用 Docker
```bash
docker-compose up -d
```

## 访问地址

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:3000 |
| 后端 API | http://localhost:8080 |
| API 文档 | http://localhost:8080/swagger |
| MinIO Console | http://localhost:9001 |

## 测试 API

```bash
# 健康检查
curl http://localhost:8080/health

# 注册用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456","name":"Test User"}'

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'

# 获取角色模板
curl http://localhost:8080/api/v1/roles/templates
```

## 常用命令

```bash
make dev          # 启动开发服务器
make build        # 构建项目
make test         # 运行测试
make docker-up    # 启动 Docker 环境
make docker-down  # 停止 Docker 环境
```
