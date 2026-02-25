# Swagger API 文档配置指南

## 概述

RoleCraft AI 项目已集成 Swagger UI，提供交互式 API 文档。

## 访问地址

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Swagger JSON**: http://localhost:8080/swagger/swagger.json
- **健康检查**: http://localhost:8080/health

## 启动服务

```bash
# 启动后端服务
cd backend
go run cmd/server/main.go

# 或者使用构建的二进制文件
./bin/server
```

## 更新 API 文档

当修改 API 处理器时，需要更新 Swagger 注解并重新生成文档：

### 1. 添加/更新注解

在处理器函数上方添加 Swagger 注解：

```go
// @Summary 简短描述
// @Description 详细描述
// @Tags 标签名称
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "参数描述"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Router /api/v1/endpoint [method]
func (h *Handler) Endpoint(c *gin.Context) {
    // ...
}
```

### 2. 重新生成文档

```bash
# 安装 swag 工具 (首次)
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
cd backend
~/go/bin/swag init --parseDependency --parseInternal --parseDepth 6 -o ./docs -g cmd/server/main.go
```

### 3. 重启服务

```bash
# 重新编译并重启
go build -o bin/server ./cmd/server/main.go
./bin/server
```

## 已文档化的 API

### 认证 API (3 个)
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - Token 刷新

### 角色 API (7 个)
- `GET /api/v1/roles/templates` - 获取角色模板
- `GET /api/v1/roles` - 角色列表
- `GET /api/v1/roles/:id` - 角色详情
- `POST /api/v1/roles` - 创建角色
- `PUT /api/v1/roles/:id` - 更新角色
- `DELETE /api/v1/roles/:id` - 删除角色
- `POST /api/v1/roles/:id/chat` - 与角色对话

## 认证说明

需要认证的端点使用 JWT Bearer Token：

```
Authorization: Bearer {your_jwt_token}
```

在 Swagger UI 中：
1. 点击顶部的 "Authorize" 按钮
2. 输入 `Bearer {your_token}`
3. 点击 "Authorize"
4. 现在可以测试需要认证的端点

## 测试

```bash
# 运行 API 测试
./tests/api_test.sh

# 运行 E2E 测试
./tests/e2e_test.sh

# 运行完整测试套件
./tests/run_all_tests.sh
```

## 故障排除

### Swagger UI 显示 404
- 确保 `backend/docs/docs.go` 已生成
- 检查 `main.go` 中是否正确导入 docs 包
- 确认 `swag init` 成功执行

### API 端点未显示
- 检查函数上方是否有完整的 Swagger 注解
- 确认 `@Router` 路径与实际路由一致
- 重新运行 `swag init` 生成文档

## 参考

- [Swag 文档](https://github.com/swaggo/swag)
- [Gin-Swagger](https://github.com/swaggo/gin-swagger)
- [Swagger 规范](https://swagger.io/specification/)
