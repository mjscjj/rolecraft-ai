# RoleCraft AI Backend Scripts

## 数据库迁移 (Database Migration)

### v2 迁移 - AnythingLLM Workspace 关联

迁移脚本：`migrate_v2.go`

#### 主要变更

1. **User 模型**
   - 新增 `anything_llm_slug` 字段：关联 AnythingLLM Workspace
   - 添加索引优化查询

2. **Role 模型** - 简化
   - 移除 `workspace_id` 字段（废弃）
   - 新增 `user_id` 字段：直接关联用户
   - 移除 `skills` 和 `documents` 多对多关联
   - 添加索引：`user_id`, `is_template`, `is_public`

3. **Document 模型** - 添加 AnythingLLM 关联
   - 新增 `anything_llm_hash` 字段：AnythingLLM 文档 hash
   - 新增 `user_id` 字段：替代 `workspace_id`
   - 自动迁移现有数据的 `workspace_id` → `user_id`
   - 添加索引：`user_id`, `anything_llm_hash`, `status`

4. **ChatSession 模型** - 添加关联
   - 新增 `anything_llm_slug` 字段：关联 AnythingLLM Workspace
   - 新增 `user_id` 字段（如果不存在）
   - 添加索引：`user_id`, `role_id`, `anything_llm_slug`

5. **索引优化**
   - 为所有外键字段添加索引
   - 创建复合索引优化常用查询
   - 清理废弃的关联表（`role_skills`, `role_documents`）

#### 运行迁移

```bash
cd backend

# 方式 1: 使用默认数据库路径
go run scripts/migrate_v2.go

# 方式 2: 指定数据库路径
DB_PATH=/path/to/rolecraft.db go run scripts/migrate_v2.go
```

#### 迁移特性

- ✅ 幂等操作：可重复运行，不会重复添加列
- ✅ 数据保留：自动迁移现有数据
- ✅ 索引优化：自动创建性能索引
- ✅ 清理废弃：删除不再需要的关联表
- ✅ 详细日志：显示每个迁移步骤

## 测试 (Testing)

### 运行模型测试

```bash
cd backend

# 运行所有测试
go test -v ./scripts/...

# 运行特定测试
go test -v -run TestUserCRUD ./scripts/...
go test -v -run TestRoleCRUD ./scripts/...
go test -v -run TestDocumentCRUD ./scripts/...
go test -v -run TestChatSessionCRUD ./scripts/...
go test -v -run TestIndexes ./scripts/...
go test -v -run TestAnythingLLMIntegration ./scripts/...
```

### 测试覆盖

- ✅ **User CRUD**: 创建、读取、更新、删除
- ✅ **Role CRUD**: 简化后的模型操作
- ✅ **Document CRUD**: AnythingLLM hash 关联
- ✅ **ChatSession CRUD**: Workspace 关联
- ✅ **索引性能**: 查询性能测试
- ✅ **AnythingLLM 集成**: 完整关联测试
- ✅ **模型验证**: 约束和唯一性测试

## 模型变更总结

### 迁移前 (v1)

```go
type User struct {
    ID            string
    Email         string
    // ... 无 AnythingLLM 关联
}

type Role struct {
    ID             string
    WorkspaceID    string  // ❌ 移除
    // ... 复杂的技能/文档关联
}

type Document struct {
    ID           string
    WorkspaceID  string  // ❌ 移除
    // ... 无 AnythingLLM hash
}
```

### 迁移后 (v2)

```go
type User struct {
    ID              string
    Email           string
    AnythingLLMSlug string  // ✅ 新增
    // ...
}

type Role struct {
    ID             string
    UserID         string  // ✅ 直接关联用户
    // ... 简化模型
}

type Document struct {
    ID              string
    UserID          string  // ✅ 直接关联用户
    AnythingLLMHash string  // ✅ 新增
    // ...
}

type ChatSession struct {
    ID              string
    UserID          string
    AnythingLLMSlug string  // ✅ 新增
    // ...
}
```

## 注意事项

1. **备份数据库**: 运行迁移前建议备份现有数据库
2. **停机时间**: 迁移过程中应停止应用服务
3. **测试环境**: 建议先在测试环境验证迁移
4. **回滚**: 当前迁移不支持自动回滚，需要手动恢复备份

## 性能优化

迁移后创建的索引：

- `idx_users_anything_llm_slug`: 按 Workspace 查询用户
- `idx_roles_user_id`: 按用户查询角色
- `idx_roles_is_template`: 筛选模板角色
- `idx_roles_is_public`: 筛选公开角色
- `idx_documents_user_id`: 按用户查询文档
- `idx_documents_anything_llm_hash`: 按 AnythingLLM hash 查询
- `idx_documents_status`: 按状态筛选文档
- `idx_chat_sessions_user_id`: 按用户查询会话
- `idx_chat_sessions_role_id`: 按角色查询会话
- `idx_chat_sessions_anything_llm_slug`: 按 Workspace 查询会话
- `idx_messages_session_id`: 按会话查询消息
- 复合索引优化常用组合查询

## 下一步

迁移完成后：

1. 更新 API handlers 以使用新模型
2. 更新服务层逻辑
3. 更新前端代码适配新结构
4. 监控查询性能
5. 根据实际使用情况优化索引
