# RoleCraft AI Frontend - Production Dockerfile
# 基于 Nginx 的生产配置

# ============================================
# 阶段 1: 构建
# ============================================
FROM node:20-alpine AS builder

WORKDIR /app

# 安装依赖
COPY package*.json ./
RUN npm ci --only=production

# 复制源代码
COPY . .

# 构建参数
ARG VITE_API_URL=/api

# 构建应用
RUN npm run build

# ============================================
# 阶段 2: Nginx 生产服务器
# ============================================
FROM nginx:alpine AS production

# 复制自定义 Nginx 配置
COPY nginx.conf /etc/nginx/nginx.conf
COPY nginx.default.conf /etc/nginx/conf.d/default.conf

# 从构建阶段复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 创建必要目录
RUN mkdir -p /var/cache/nginx /var/log/nginx && \
    chown -R nginx:nginx /var/cache/nginx /var/log/nginx /usr/share/nginx/html

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:80/ || exit 1

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
