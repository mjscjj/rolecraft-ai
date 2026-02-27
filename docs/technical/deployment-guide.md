# RoleCraft AI - 部署指南

> 生产环境部署完整步骤

---

## 目录

1. [部署前准备](#1-部署前准备)
2. [服务器要求](#2-服务器要求)
3. [Docker 部署（推荐）](#3-docker 部署推荐)
4. [源码部署](#4-源码部署)
5. [配置说明](#5-配置说明)
6. [运维管理](#6-运维管理)

---

## 1. 部署前准备

### 1.1 检查清单

- [ ] 服务器资源到位
- [ ] 域名和 SSL 证书准备
- [ ] 数据库服务可用
- [ ] 环境变量配置
- [ ] 备份策略制定

### 1.2 获取部署文件

```bash
# 克隆项目
git clone https://github.com/mjscjj/rolecraft-ai.git
cd rolecraft-ai

# 切换到稳定版本
git checkout v1.0.0
```

---

## 2. 服务器要求

### 2.1 最低配置

| 组件 | 配置 |
|------|------|
| CPU | 2 核 |
| 内存 | 4GB |
| 磁盘 | 50GB SSD |
| 带宽 | 5Mbps |

### 2.2 推荐配置

| 组件 | 配置 |
|------|------|
| CPU | 4 核 |
| 内存 | 8GB |
| 磁盘 | 100GB SSD |
| 带宽 | 10Mbps |

### 2.3 软件要求

- Ubuntu 20.04+ / CentOS 7+
- Docker 20.10+
- Docker Compose 2.0+
- Nginx 1.20+（反向代理）

---

## 3. Docker 部署（推荐）

### 3.1 安装 Docker

```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# 验证安装
docker --version
docker-compose --version
```

### 3.2 配置环境变量

```bash
cp .env.example .env
vim .env
```

**.env 配置示例：**
```bash
# 数据库
DATABASE_URL=postgresql://user:password@localhost:5432/rolecraft

# Redis
REDIS_URL=redis://localhost:6379

# JWT 配置
JWT_SECRET=your-secret-key-change-this

# OpenAI API
OPENAI_API_KEY=sk-...

# 文件存储
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# 服务端口
BACKEND_PORT=8080
FRONTEND_PORT=3000
```

### 3.3 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 检查状态
docker-compose ps
```

### 3.4 初始化数据库

```bash
# 进入后端容器
docker-compose exec backend sh

# 运行数据库迁移
go run cmd/migrate/main.go up

# 初始化基础数据
go run cmd/migrate/main.go seed
```

### 3.5 配置 Nginx

**/etc/nginx/sites-available/rolecraft:**
```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 后端 API
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Swagger 文档
    location /swagger {
        proxy_pass http://localhost:8080/swagger;
    }
}
```

```bash
# 启用配置
sudo ln -s /etc/nginx/sites-available/rolecraft /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 3.6 配置 HTTPS

```bash
# 使用 Let's Encrypt
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

---

## 4. 源码部署

### 4.1 安装依赖

```bash
# Go 环境
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Node.js 环境
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# PostgreSQL
sudo apt install postgresql postgresql-contrib

# Redis
sudo apt install redis-server
```

### 4.2 编译后端

```bash
cd backend
go mod download
go build -o bin/server cmd/server/main.go
```

### 4.3 编译前端

```bash
cd frontend
npm install
npm run build
```

### 4.4 配置 Systemd

**/etc/systemd/system/rolecraft-backend.service:**
```ini
[Unit]
Description=RoleCraft AI Backend
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/rolecraft-ai/backend
ExecStart=/opt/rolecraft-ai/backend/bin/server
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
# 启用服务
sudo systemctl daemon-reload
sudo systemctl enable rolecraft-backend
sudo systemctl start rolecraft-backend
```

---

## 5. 配置说明

### 5.1 后端配置

| 配置项 | 说明 | 默认值 |
|-------|------|--------|
| `DATABASE_URL` | 数据库连接 | - |
| `REDIS_URL` | Redis 连接 | - |
| `JWT_SECRET` | JWT 密钥 | - |
| `OPENAI_API_KEY` | OpenAI 密钥 | - |
| `SERVER_PORT` | 服务端口 | 8080 |

### 5.2 前端配置

| 配置项 | 说明 | 默认值 |
|-------|------|--------|
| `VITE_API_URL` | API 地址 | http://localhost:8080 |
| `VITE_WS_URL` | WebSocket 地址 | ws://localhost:8080 |

---

## 6. 运维管理

### 6.1 日志管理

```bash
# 查看后端日志
docker-compose logs backend

# 查看前端日志
docker-compose logs frontend

# 实时日志
docker-compose logs -f
```

### 6.2 备份策略

**数据库备份：**
```bash
# 备份脚本
#!/bin/bash
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d).sql
# 上传到对象存储
```

**定时任务：**
```bash
# crontab -e
0 2 * * * /opt/rolecraft-ai/scripts/backup.sh
```

### 6.3 监控告警

**健康检查：**
```bash
curl http://localhost:8080/health
```

**监控指标：**
- CPU 使用率
- 内存使用率
- 磁盘空间
- 响应时间
- 错误率

### 6.4 更新升级

```bash
# 拉取新版本
git pull origin main

# 停止服务
docker-compose down

# 重新构建
docker-compose build

# 启动服务
docker-compose up -d

# 运行迁移
docker-compose exec backend go run cmd/migrate/main.go up
```

---

## 📞 部署支持

遇到问题？联系技术支持：
- 📧 support@rolecraft.ai
- 📖 [GitHub Issues](https://github.com/mjscjj/rolecraft-ai/issues)

---

*最后更新：2026-02-27*
