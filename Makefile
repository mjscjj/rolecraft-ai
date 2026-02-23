.PHONY: dev build test clean docker-up docker-down migrate

# 开发环境启动
dev:
	cd backend && go run cmd/server/main.go

# 前端开发
dev-frontend:
	cd frontend && pnpm dev

# 构建后端
build-backend:
	cd backend && go build -o bin/server cmd/server/main.go

# 构建前端
build-frontend:
	cd frontend && pnpm build

# 构建全部
build: build-backend build-frontend

# 运行测试
test-backend:
	cd backend && go test ./... -v

test-frontend:
	cd frontend && pnpm test

test: test-backend test-frontend

# 清理构建产物
clean:
	rm -rf backend/bin
	rm -rf frontend/dist
	rm -rf uploads/*

# Docker 操作
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-reset: docker-down
	docker-compose down -v
	docker-compose up -d

# 数据库迁移
migrate-up:
	cd backend && go run cmd/migrate/main.go up

migrate-down:
	cd backend && go run cmd/migrate/main.go down

# 代码质量
lint:
	cd backend && golangci-lint run
	cd frontend && pnpm lint

fmt:
	cd backend && go fmt ./...
	cd frontend && pnpm format

# 安装依赖
install-backend:
	cd backend && go mod download

install-frontend:
	cd frontend && pnpm install

install: install-backend install-frontend

# 完整开发环境设置
setup: install docker-up
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Development environment is ready!"
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:3000"
	@echo "MinIO Console: http://localhost:9001"
