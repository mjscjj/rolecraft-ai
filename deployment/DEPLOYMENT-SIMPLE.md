# RoleCraft AI - 极简部署方案

**原则**: 最少依赖、架构整洁、快速启动  
**更新**: 2026-02-27

---

## 🎯 核心架构（精简版）

### 最小依赖
```
┌─────────────────────────────────────┐
│                                     │
│  前端 (React + Vite)                │
│  Port: 5173 (dev) / 3000 (prod)     │
│                                     │
│  ↓ HTTP                             │
│                                     │
│  后端 (Go + Gin)                    │
│  Port: 8080                         │
│                                     │
│  ↓ SQLite (嵌入式数据库)             │
│                                     │
│  rolecraft.db (单文件)               │
│                                     │
└─────────────────────────────────────┘
```

**移除的依赖**（可选）:
- ❌ PostgreSQL → 改用 SQLite（开发/小规模）
- ❌ Redis → 内存缓存（小规模不需要）
- ❌ Milvus → 暂不需要向量搜索
- ❌ MinIO → 本地文件存储

**保留的核心**:
- ✅ SQLite（单文件数据库，零配置）
- ✅ Mock AI（无需 OpenAI Key）
- ✅ AnythingLLM（可选，已有远程服务）

---

## 🚀 一、快速启动（开发环境）

### 方式 1：最简启动（推荐）

**步骤 1: 启动后端**
```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/backend

# 直接运行（使用 SQLite）
go run cmd/server/main.go
```

**步骤 2: 启动前端**
```bash
# 另开终端
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/frontend

npm run dev
```

**访问**:
- 前端：http://localhost:5173
- 后端：http://localhost:8080

---

### 方式 2：使用启动脚本

```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai

# 一键启动前后端
./start.sh
```

**start.sh 内容**:
```bash
#!/bin/bash

echo "🚀 Starting RoleCraft AI..."

# 启动后端
echo "Starting backend..."
cd backend
go run cmd/server/main.go &
BACKEND_PID=$!

# 等待后端启动
sleep 2

# 启动前端
echo "Starting frontend..."
cd ../frontend
npm run dev &
FRONTEND_PID=$!

echo ""
echo "✅ Services started!"
echo "   Frontend: http://localhost:5173"
echo "   Backend:  http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop"

# 等待中断信号
trap "kill $BACKEND_PID $FRONTEND_PID" EXIT
wait
```

---

## 📦 二、生产部署（简化版）

### 方案 A：单机部署（无 Docker）

**步骤 1: 构建后端**
```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/backend

# 编译为单个二进制文件
go build -o bin/server cmd/server/main.go

# 验证
./bin/server --version
```

**步骤 2: 构建前端**
```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/frontend

# 构建静态文件
npm run build

# 输出到 dist/ 目录
ls -la dist/
```

**步骤 3: 配置 Nginx**
```bash
# 安装 Nginx
brew install nginx  # macOS
# apt-get install nginx  # Linux

# 配置 Nginx
sudo vim /usr/local/etc/nginx/servers/rolecraft.conf
```

**Nginx 配置**:
```nginx
server {
    listen 80;
    server_name localhost;

    # 前端静态文件
    location / {
        root /Users/claw/.openclaw/workspace-work/rolecraft-ai/frontend/dist;
        try_files $uri $uri/ /index.html;
    }

    # 后端 API 代理
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 静态资源
    location /assets/ {
        root /Users/claw/.openclaw/workspace-work/rolecraft-ai/frontend/dist;
        expires 30d;
    }
}
```

**步骤 4: 启动服务**
```bash
# 启动 Nginx
sudo nginx -s reload

# 启动后端（后台运行）
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/backend
./bin/server &

# 或使用 nohup
nohup ./bin/server > server.log 2>&1 &

# 验证
curl http://localhost/health
```

---

### 方案 B：Docker 部署（单容器）

**后端 Dockerfile**（已存在）:
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./server"]
```

**前端 Dockerfile**（已存在）:
```dockerfile
FROM node:18-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**简化版 docker-compose.yml**:
```yaml
version: '3.8'

services:
  backend:
    build: ./backend
    container_name: rolecraft-backend
    environment:
      - DATABASE_URL=sqlite://./rolecraft.db
      - JWT_SECRET=your-secret-key
      - PORT=8080
    volumes:
      - ./backend/data:/root/data
    ports:
      - "8080:8080"
    restart: unless-stopped

  frontend:
    build: ./frontend
    container_name: rolecraft-frontend
    ports:
      - "3000:80"
    depends_on:
      - backend
    restart: unless-stopped
```

**启动**:
```bash
docker-compose up -d
```

---

## 🔧 三、配置说明

### 环境变量（最小配置）

**创建 `.env` 文件**:
```bash
# 基础配置
ENV=production
PORT=8080

# 数据库（使用 SQLite）
DATABASE_URL=sqlite://./rolecraft.db

# JWT 密钥（生产环境务必修改）
JWT_SECRET=your-super-secret-key-change-me

# 可选：AnythingLLM（已有远程服务）
ANYTHINGLLM_URL=http://150.109.21.115:3001
ANYTHINGLLM_KEY=sk-WaUmgZsMxgeHOpp8SJxK1rmVQxiwfiDJ

# 可选：OpenAI（如果需要真实 AI）
# OPENAI_API_KEY=sk-xxx
```

### 数据库配置

**SQLite（推荐用于开发/小规模）**:
```bash
# 自动创建，无需配置
# 文件位置：backend/rolecraft.db
```

**PostgreSQL（可选，用于生产）**:
```bash
# 如需要 PostgreSQL，设置环境变量
DATABASE_URL=postgres://user:password@localhost:5432/rolecraft?sslmode=disable
```

---

## 📊 四、目录结构（精简后）

```
rolecraft-ai/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # 后端入口
│   ├── internal/
│   │   ├── api/                 # API 处理器
│   │   ├── config/              # 配置加载
│   │   ├── models/              # 数据模型
│   │   ├── service/             # 业务逻辑
│   │   └── middleware/          # 中间件
│   ├── data/
│   │   └── rolecraft.db         # SQLite 数据库
│   ├── uploads/                 # 上传文件
│   └── go.mod                   # Go 依赖
│
├── frontend/
│   ├── src/
│   │   ├── pages/               # 页面组件
│   │   ├── components/          # 通用组件
│   │   ├── api/                 # API 调用
│   │   └── App.tsx              # 应用入口
│   ├── dist/                    # 构建输出
│   └── package.json             # Node 依赖
│
├── start.sh                     # 启动脚本
├── stop.sh                      # 停止脚本
└── .env                         # 环境变量
```

---

## ✅ 五、快速验证

### 测试脚本（简化版）

**创建 `test-simple.sh`**:
```bash
#!/bin/bash

echo "========================================="
echo "  RoleCraft AI 快速测试"
echo "========================================="

# 1. 检查后端
echo "1️⃣  检查后端..."
if curl -s "http://localhost:8080/health" | grep -q "ok"; then
    echo "   ✅ 后端正常"
else
    echo "   ❌ 后端异常"
    exit 1
fi

# 2. 检查前端
echo "2️⃣  检查前端..."
if curl -s "http://localhost:5173" | grep -q "html"; then
    echo "   ✅ 前端正常"
else
    echo "   ❌ 前端异常"
    exit 1
fi

# 3. 测试 API
echo "3️⃣  测试 API..."
if curl -s "http://localhost:8080/api/v1/roles" | grep -q "data"; then
    echo "   ✅ API 正常"
else
    echo "   ❌ API 异常"
    exit 1
fi

echo ""
echo "========================================="
echo "  ✅ 所有测试通过！"
echo "========================================="
```

**运行测试**:
```bash
chmod +x test-simple.sh
./test-simple.sh
```

---

## 🎯 六、核心功能验证

### 1. 用户认证
```bash
# 注册用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456","name":"Test"}'

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
```

### 2. 角色管理
```bash
# 获取角色列表
curl http://localhost:8080/api/v1/roles

# 创建角色
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"name":"测试助手","description":"描述","systemPrompt":"你是一个助手"}'
```

### 3. 对话功能
```bash
# 创建会话
curl -X POST http://localhost:8080/api/v1/chat-sessions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"roleId":"xxx","mode":"quick"}'

# 发送消息（Mock AI）
curl -X POST http://localhost:8080/api/v1/chat/SESSION_ID/complete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"content":"你好"}'
```

### 4. 深度思考
```bash
# 带思考的对话
curl -X POST http://localhost:8080/api/v1/chat/SESSION_ID/stream-with-thinking \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"content":"如何学习 Go 语言？"}'
```

---

## 📝 七、常见问题

### Q1: 数据库文件在哪？
```bash
# SQLite 数据库位置
ls -la backend/rolecraft.db

# 或使用绝对路径
DATABASE_URL=sqlite:///Users/claw/.openclaw/workspace-work/rolecraft-ai/backend/rolecraft.db
```

### Q2: 如何重置数据库？
```bash
# 删除数据库文件
rm backend/rolecraft.db

# 重启后端，自动创建新数据库
go run cmd/server/main.go
```

### Q3: 上传文件存在哪？
```bash
# 默认上传目录
ls -la backend/uploads/

# 可配置环境变量
UPLOAD_DIR=/path/to/uploads
```

### Q4: 如何查看日志？
```bash
# 后端日志（直接运行）
# 日志输出到终端

# 后端日志（nohup 运行）
tail -f backend/server.log

# Nginx 日志
tail -f /usr/local/var/log/nginx/access.log
tail -f /usr/local/var/log/nginx/error.log
```

### Q5: 如何停止服务？
```bash
# 使用停止脚本
./stop.sh

# 或手动停止
# Ctrl+C（终端运行）

# 或 kill 进程
pkill -f "go run cmd/server"
pkill -f "npm run dev"

# Docker 方式
docker-compose down
```

---

## 🎉 八、性能优化建议

### 1. 启用 Gzip 压缩

**Nginx 配置**:
```nginx
gzip on;
gzip_vary on;
gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
gzip_min_length 1000;
```

### 2. 静态资源缓存

**Nginx 配置**:
```nginx
location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

### 3. 数据库优化

```sql
-- 添加索引（自动执行）
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
```

### 4. 启用 HTTPS（生产环境）

```bash
# 使用 Caddy（自动 HTTPS）
brew install caddy

# Caddyfile
example.com {
    reverse_proxy localhost:8080
    root * /path/to/frontend/dist
    file_server
}
```

---

## 📊 九、架构对比

### 原方案（复杂）
```
前端 → Nginx → 后端 → PostgreSQL
              ↓ Redis
              ↓ Milvus
              ↓ MinIO
```
**依赖**: 5 个服务  
**启动时间**: ~5 分钟  
**内存占用**: ~2GB  
**适用**: 大规模生产环境

---

### 新方案（极简）
```
前端 → Nginx → 后端 → SQLite
```
**依赖**: 2 个服务（前后端）  
**启动时间**: ~10 秒  
**内存占用**: ~200MB  
**适用**: 开发/测试/小规模生产

---

## 🚀 十、快速开始清单

- [ ] 1. 克隆项目
- [ ] 2. 安装 Go 1.21+
- [ ] 3. 安装 Node.js 18+
- [ ] 4. `cd backend && go run cmd/server/main.go`
- [ ] 5. `cd frontend && npm run dev`
- [ ] 6. 访问 http://localhost:5173
- [ ] 7. 注册账号并测试

**总计时间**: < 5 分钟  
**依赖数量**: 2（Go + Node.js）  
**数据库**: SQLite（零配置）

---

**极简部署，快速启动！** 🎉

需要我帮你执行具体的启动操作吗？
