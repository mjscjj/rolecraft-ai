# 角色与 AnythingLLM 同步实现总结

## ✅ 已完成任务

### 1. 添加 AnythingLLM UpdateWorkspaceSystemPrompt 方法

**文件**: `backend/internal/service/anythingllm/client.go`

- 新增 `UpdateWorkspaceSystemPrompt(slug, systemPrompt string) error` 方法
- 使用 HTTP PATCH 请求更新 workspace 的系统提示词
- 支持重试机制（继承自 `doRequest`）

**文件**: `backend/internal/service/anythingllm/types.go`

- 新增 `UpdateWorkspaceRequest` 类型
- 新增 `UpdateWorkspaceResponse` 类型

### 2. 导出 GetWorkspaceSlug 方法

**文件**: `backend/internal/service/anythingllm/client.go`

- 将 `getWorkspaceSlug` 改为导出方法 `GetWorkspaceSlug`
- 保留原有方法作为兼容层

### 3. 修改 RoleHandler 支持异步同步

**文件**: `backend/internal/api/handler/role.go`

#### 结构变更:
```go
type RoleHandler struct {
    db              *gorm.DB
    anythingllmURL  string
    anythingllmKey  string
}
```

#### 构造函数变更:
```go
func NewRoleHandler(db *gorm.DB, cfg *config.Config) *RoleHandler
```

#### Create 方法新增异步同步:
```go
go func() {
    client := anythingllm.NewAnythingLLMClient(h.anythingllmURL, h.anythingllmKey)
    slug := client.GetWorkspaceSlug(role.ID)
    err := client.UpdateWorkspaceSystemPrompt(slug, role.SystemPrompt)
    if err != nil {
        log.Printf("⚠️ 角色 [%s] 同步到 AnythingLLM 失败：%v", role.Name, err)
    } else {
        log.Printf("✅ 角色 [%s] 已同步到 AnythingLLM", role.Name)
    }
}()
```

#### Update 方法新增异步同步:
同样的异步同步逻辑，在角色更新后触发

### 4. 更新 main.go

**文件**: `backend/cmd/server/main.go`

- 修改 `NewRoleHandler(db, cfg)` 调用，传入配置

### 5. 添加测试用例

**文件**: `backend/tests/role_sync_test.go`

- `TestRoleAnythingLLMSync` 测试函数
- 测试创建角色时的同步
- 测试更新角色时的同步

## 🔍 测试验证

```bash
cd backend
go test -v ./tests/role_sync_test.go -run TestRoleAnythingLLMSync
```

**测试结果**:
```
=== RUN   TestRoleAnythingLLMSync
=== RUN   TestRoleAnythingLLMSync/CreateRole_ShouldSyncToAnythingLLM
    ✅ 角色创建成功，异步同步已触发
=== RUN   TestRoleAnythingLLMSync/UpdateRole_ShouldSyncToAnythingLLM
    ✅ 角色更新成功，异步同步已触发
--- PASS: TestRoleAnythingLLMSync (0.01s)
PASS
```

## 📝 代码特点

1. **异步非阻塞**: 使用 goroutine 异步同步，不阻断主流程
2. **错误处理**: 同步失败时记录日志，不影响主流程
3. **日志记录**: 成功/失败都有详细的日志输出
4. **向后兼容**: 保留原有的 `getWorkspaceSlug` 方法

## 🚀 使用说明

### 创建角色时自动同步:
```bash
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "智能助理",
    "description": "全能型办公助手",
    "category": "通用",
    "systemPrompt": "你是一位智能助理...",
    "welcomeMessage": "你好！"
  }'
```

### 更新角色时自动同步:
```bash
curl -X PUT http://localhost:8080/api/v1/roles/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "智能助理",
    "description": "更新后的描述",
    "category": "通用",
    "systemPrompt": "更新后的系统提示词",
    "welcomeMessage": "更新后的欢迎消息"
  }'
```

## 📊 日志示例

**同步成功**:
```
2026/02/26 20:23:12 ✅ 角色 [智能助理] 已同步到 AnythingLLM
```

**同步失败**:
```
2026/02/26 20:23:12 ⚠️ 角色 [智能助理] 同步到 AnythingLLM 失败：connection refused
```

## ⚠️ 注意事项

1. AnythingLLM 服务必须可访问
2. Workspace 必须已存在（slug 格式：`user_{roleId}`）
3. 异步同步不保证立即完成，适合最终一致性场景
4. 如需强一致性，建议使用同步调用或添加重试机制

## 📁 修改文件清单

1. `backend/internal/service/anythingllm/client.go` - 新增 UpdateWorkspaceSystemPrompt 方法
2. `backend/internal/service/anythingllm/types.go` - 新增请求/响应类型
3. `backend/internal/api/handler/role.go` - 添加异步同步逻辑
4. `backend/cmd/server/main.go` - 更新 RoleHandler 初始化
5. `backend/tests/role_sync_test.go` - 新增测试用例（可选）
