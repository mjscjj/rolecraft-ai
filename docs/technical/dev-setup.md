# RoleCraft AI - 开发环境配置指南

> 本地开发环境搭建完整步骤

---

## 目录

1. [环境要求](#1-环境要求)
2. [快速开始](#2-快速开始)
3. [后端配置](#3-后端配置)
4. [前端配置](#4-前端配置)
5. [数据库配置](#5-数据库配置)
6. [常见问题](#6-常见问题)

---

## 1. 环境要求

### 1.1 必需软件

| 软件 | 版本 | 用途 |
|------|------|------|
| Go | 1.21+ | 后端开发 |
| Node.js | 18+ | 前端开发 |
| Git | 最新 | 版本控制 |

### 1.2 可选软件

| 软件 | 用途 |
|------|------|
| Docker | 容器化部署 |
| PostgreSQL | 生产数据库 |
| Redis | 缓存和会话 |

---

## 2. 快速开始

### 2.1 克隆项目

```bash
git clone https://github.com/mjscjj/rolecraft-ai.git
cd rolecraft-ai
```

### 2.2 使用 Docker Compose（推荐）

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

访问：
- 前端：http://localhost:5173
- 后端：http://localhost:8080
- Swagger：http://localhost:8080/swagger

---

## 3. 后端配置

### 3.1 安装 Go 依赖

```bash
cd backend
go mod download
```

### 3.2 配置环境变量

```bash
cp .env.example .env
```

**.env 配置：**
```bash
# 开发环境使用 SQLite
DATABASE_URL=sqlite.db

# 或使用 PostgreSQL
# DATABASE_URL=postgresql://localhost:5432/rolecraft?sslmode=disable

REDIS_URL=redis://localhost:6379
JWT_SECRET=dev-secret-key-change-in-production
OPENAI_API_KEY=sk-your-key-here
SERVER_PORT=8080
```

### 3.3 运行数据库迁移

```bash
cd backend
go run cmd/migrate/main.go up
go run cmd/migrate/main.go seed
```

### 3.4 启动后端服务

```bash
# 开发模式（支持热重载）
go run cmd/server/main.go

# 或编译后运行
go build -o bin/server cmd/server/main.go
./bin/server
```

### 3.5 运行测试

```bash
# 单元测试
go test ./...

# API 测试
./tests/api_test.sh

# 查看测试覆盖率
go test -cover ./...
```

---

## 4. 前端配置

### 4.1 安装 Node.js 依赖

```bash
cd frontend
npm install
# 或使用 pnpm
pnpm install
```

### 4.2 配置环境变量

```bash
cp .env.example .env.local
```

**.env.local 配置：**
```bash
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

### 4.3 启动开发服务器

```bash
npm run dev
# 或
pnpm dev
```

访问：http://localhost:5173

### 4.4 构建生产版本

```bash
npm run build
# 输出到 dist/ 目录
```

### 4.5 运行前端测试

```bash
# 单元测试
npm run test

# E2E 测试
npx playwright test

# 生成测试报告
npx playwright test --reporter=html
```

---

## 5. 数据库配置

### 5.1 SQLite（开发环境）

**优点：**
- 零配置
- 单文件
- 适合开发和测试

**配置：**
```bash
DATABASE_URL=sqlite.db
```

### 5.2 PostgreSQL（生产环境）

**安装（macOS）：**
```bash
brew install postgresql
brew services start postgresql
```

**安装（Ubuntu）：**
```bash
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

**创建数据库：**
```bash
sudo -u postgres psql
CREATE DATABASE rolecraft;
CREATE USER rolecraft_user WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE rolecraft TO rolecraft_user;
```

**配置：**
```bash
DATABASE_URL=postgresql://rolecraft_user:password@localhost:5432/rolecraft?sslmode=disable
```

### 5.3 Redis（可选）

**安装：**
```bash
# macOS
brew install redis
brew services start redis

# Ubuntu
sudo apt install redis-server
sudo systemctl start redis
```

**配置：**
```bash
REDIS_URL=redis://localhost:6379
```

---

## 6. 常见问题

### Q1: Go 依赖下载失败？

**解决方案：**
```bash
# 使用国内镜像
export GOPROXY=https://goproxy.cn,direct
go mod download
```

### Q2: Node.js 版本不兼容？

**解决方案：**
```bash
# 使用 nvm 管理 Node 版本
nvm install 18
nvm use 18
```

### Q3: 端口被占用？

**解决方案：**
```bash
# 修改端口
# 后端 .env: SERVER_PORT=8081
# 前端 .env.local: VITE_API_URL=http://localhost:8081
```

### Q4: 数据库迁移失败？

**解决方案：**
```bash
# 重置数据库
rm sqlite.db
go run cmd/migrate/main.go up
go run cmd/migrate/main.go seed
```

### Q5: 前端构建失败？

**解决方案：**
```bash
# 清理缓存
rm -rf node_modules package-lock.json
npm install
npm run build
```

---

## 📚 相关文档

- [API 参考文档](./api-reference.md)
- [数据库设计文档](./database-design.md)
- [系统架构图](./architecture.md)

---

*最后更新：2026-02-27*
