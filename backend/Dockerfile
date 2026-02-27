# RoleCraft AI Backend - Production Dockerfile
# 多阶段构建优化镜像大小

# ============================================
# 阶段 1: 依赖缓存
# ============================================
FROM golang:1.21-alpine AS deps
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ============================================
# 阶段 2: 构建
# ============================================
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app

# 复制依赖
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译优化
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/bin/server ./cmd/server/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/migrate ./cmd/migrate/main.go

# ============================================
# 阶段 3: 生产运行
# ============================================
FROM alpine:3.19 AS production

# 安装必要工具
RUN apk add --no-cache ca-certificates tzdata curl

# 创建非 root 用户
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder --chown=appuser:appgroup /app/bin/server /app/
COPY --from=builder --chown=appuser:appgroup /app/bin/migrate /app/
COPY --from=builder --chown=appuser:appgroup /app/migrations /app/migrations

# 创建必要目录
RUN mkdir -p /app/uploads /app/logs && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动命令
ENTRYPOINT ["/app/server"]
