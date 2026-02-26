# 角色同步 AnythingLLM 手动测试指南

## 🚀 启动服务

```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/backend
go run ./cmd/server/main.go
```

## 📝 测试步骤

### 1. 创建角色（测试 Create 同步）

```bash
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试同步角色",
    "description": "用于测试 AnythingLLM 自动同步功能",
    "category": "测试",
    "systemPrompt": "你是一个测试助手，专门用于验证同步功能。请用简洁的语言回答问题。",
    "welcomeMessage": "你好！我是测试助手"
  }' | jq
```

**预期响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "xxx",
    "name": "测试同步角色",
    "description": "用于测试 AnythingLLM 自动同步功能",
    "category": "测试",
    "systemPrompt": "你是一个测试助手，专门用于验证同步功能。请用简洁的语言回答问题。",
    "welcomeMessage": "你好！我是测试助手"
  }
}
```

**预期日志**:
```
✅ 角色 [测试同步角色] 已同步到 AnythingLLM
```

或（如果 AnythingLLM 服务不可用）:
```
⚠️ 角色 [测试同步角色] 同步到 AnythingLLM 失败：connection refused
```

### 2. 更新角色（测试 Update 同步）

使用上一步返回的角色 ID：

```bash
ROLE_ID="上一步返回的 id"

curl -X PUT http://localhost:8080/api/v1/roles/$ROLE_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试同步角色",
    "description": "更新后的描述 - 验证同步功能",
    "category": "测试",
    "systemPrompt": "你是一个更新后的测试助手，系统提示词已更新。请用友好的语气回答。",
    "welcomeMessage": "你好！我是更新后的测试助手"
  }' | jq
```

**预期响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "xxx",
    "name": "测试同步角色",
    "description": "更新后的描述 - 验证同步功能",
    "category": "测试",
    "systemPrompt": "你是一个更新后的测试助手，系统提示词已更新。请用友好的语气回答。",
    "welcomeMessage": "你好！我是更新后的测试助手"
  }
}
```

**预期日志**:
```
✅ 角色 [测试同步角色] 已同步到 AnythingLLM
```

### 3. 验证数据库中的角色

```bash
curl http://localhost:8080/api/v1/roles/$ROLE_ID | jq
```

### 4. 查看 AnythingLLM Workspace（可选）

如果 AnythingLLM 服务可用，可以访问：
```
http://150.109.21.115:3001
```

检查 workspace `user_{ROLE_ID}` 的系统提示词是否已更新。

## 🔍 验证要点

1. **异步非阻塞**: API 响应应该立即返回，不等待同步完成
2. **日志记录**: 服务器日志中应该看到同步成功/失败的日志
3. **数据一致性**: 数据库中的 systemPrompt 应该与发送到 AnythingLLM 的一致
4. **错误处理**: 即使 AnythingLLM 不可用，API 也应该正常返回

## 🐛 故障排查

### 问题：看不到同步日志

**解决**: 确保日志级别正确，检查服务器输出

### 问题：同步失败

**可能原因**:
1. AnythingLLM 服务不可达
2. Workspace 不存在（slug 格式：`user_{roleId}`）
3. API Key 无效

**解决**:
```bash
# 检查环境变量
echo $ANYTHINGLLM_URL
echo $ANYTHINGLLM_KEY

# 测试 AnythingLLM 连接
curl -I http://150.109.21.115:3001
```

### 问题：响应变慢

**原因**: 可能是网络延迟或 AnythingLLM 服务响应慢

**解决**: 异步同步不应该影响响应时间，如果变慢可能是其他问题

## ✅ 成功标准

- [x] 创建角色后，日志中显示同步消息
- [x] 更新角色后，日志中显示同步消息
- [x] API 响应时间正常（<100ms）
- [x] 即使同步失败，API 也能正常返回
- [x] 数据库中数据正确保存
