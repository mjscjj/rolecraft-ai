# RoleCraft AI 重构计划

**基于 ai_base 项目的用户/工作区隔离架构**

---

## 📋 重构目标

使用 **AnythingLLM Workspace** 作为用户数据隔离基础，重构 RoleCraft AI 的对话和知识库服务。

---

## 🏗️ 新架构设计

### 1. 用户隔离模型

```
┌─────────────────────────────────────────────────────────────┐
│                     RoleCraft AI                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  用户 A (user_001)          用户 B (user_002)                │
│       │                          │                          │
│       ▼                          ▼                          │
│  ┌─────────┐                ┌─────────┐                     │
│  │ RoleCraft│                │ RoleCraft│                    │
│  │ Workspace│                │ Workspace│                    │
│  │user_001 │                │user_002 │                     │
│  ├─────────┤                ├─────────┤                     │
│  │ 角色配置 │                │ 角色配置 │                    │
│  │ 知识库   │                │ 知识库   │                    │
│  │ 对话历史 │                │ 对话历史 │                    │
│  └─────────┘                └─────────┘                     │
│       │                          │                          │
│       ▼                          ▼                          │
│  ┌─────────────┐          ┌─────────────┐                   │
│  │user_001.lance│          │user_002.lance│                 │
│  │ (AnythingLLM)│          │ (AnythingLLM)│                 │
│  └─────────────┘          └─────────────┘                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 2. 核心变化

| 模块 | 旧设计 | 新设计 |
|------|--------|--------|
| **用户隔离** | userId 字段 | AnythingLLM Workspace |
| **知识库** | 本地文件 + SQLite | AnythingLLM 向量库 |
| **对话** | 本地 Message 表 | AnythingLLM Chat API |
| **角色** | 本地 Role 表 | Workspace 配置 + 系统提示词 |
| **向量检索** | ❌ 未实现 | ✅ LanceDB (内置) |

---

## 📦 依赖服务

### AnythingLLM (腾讯云 150)

| 配置项 | 值 |
|--------|-----|
| **API Base** | `http://150.109.21.115:3001/api` |
| **API Key** | `sk-WaUmgZsMxgeHOpp8SJxK1rmVQxiwfiDJ` |
| **LLM** | OpenRouter (GPT-4/Claude/Gemini) |
| **Embedding** | text-embedding-3-small |
| **Vector DB** | LanceDB |

---

## 🔧 重构步骤

### 阶段 1: 基础架构 (1-2 天)

#### 1.1 添加 AnythingLLM 客户端
```go
// backend/internal/service/anythingllm/client.go
type AnythingLLMClient struct {
    baseURL string
    apiKey  string
}

// 核心方法
- CreateWorkspace(user_id)
- UploadDocument(user_id, file)
- Chat(user_id, message, mode)
- VectorSearch(user_id, query)
- GetChatHistory(user_id, limit)
```

#### 1.2 数据库模型调整
```go
// 用户模型 - 添加 AnythingLLM 关联
type User struct {
    ID              string
    Email           string
    AnythingLLMSlug string  // 新增：Workspace slug
    CreatedAt       time.Time
}

// 角色模型 - 简化，绑定到 Workspace
type Role struct {
    ID            string
    UserID        string
    Name          string
    SystemPrompt  string  // 同步到 AnythingLLM
    ModelConfig   JSON
}

// 文档模型 - 只存元数据，实际存储在 AnythingLLM
type Document struct {
    ID              string
    UserID          string
    Name            string
    AnythingLLMHash string  // 新增：AnythingLLM 文档 hash
    Status          string  // processing/completed/failed
}
```

#### 1.3 迁移脚本
```bash
# 为现有用户创建 AnythingLLM Workspace
for user in users:
    anythingllm.create_workspace(user.id)
    user.anythingllm_slug = f"user_{user.id}"
```

---

### 阶段 2: 对话服务重构 (1-2 天)

#### 2.1 对话 API 改造
```go
// 旧的：本地存储对话
POST /api/v1/chat/:id/complete
→ 读取本地 Message 历史
→ 调用 OpenAI API
→ 保存到本地数据库

// 新的：使用 AnythingLLM
POST /api/v1/chat/:id/complete
→ 读取本地会话配置 (角色/模式)
→ 调用 AnythingLLM /v1/workspace/{slug}/chat
→ 保存会话元数据到本地 (可选)
```

#### 2.2 新增端点
```go
// 同步 AnythingLLM 对话历史
GET /api/v1/chat-sessions/:id/sync
→ 从 AnythingLLM 获取完整历史
→ 缓存到本地 (可选)

// 清除对话历史
DELETE /api/v1/chat-sessions/:id/messages
→ 调用 AnythingLLM 删除消息
```

---

### 阶段 3: 知识库重构 (2-3 天)

#### 3.1 文档上传改造
```go
// 旧的：保存到本地文件系统
POST /api/v1/documents
→ 保存到 ./uploads/{uuid}.pdf
→ 记录到数据库

// 新的：上传到 AnythingLLM
POST /api/v1/documents
→ 临时保存到本地
→ 调用 AnythingLLM /v1/document/upload
→ 调用 /v1/workspace/{slug}/update-embeddings
→ 更新本地 Document 状态
```

#### 3.2 RAG 检索
```go
// AnythingLLM 自动处理 RAG
// 对话时自动检索用户 Workspace 中的相关文档

// 可选：手动向量搜索
POST /api/v1/documents/search
→ 调用 AnythingLLM /v1/workspace/{slug}/vector-search
→ 返回相关片段
```

---

### 阶段 4: 角色系统增强 (1-2 天)

#### 4.1 角色与 Workspace 绑定
```go
// 创建角色时同步更新 AnythingLLM
POST /api/v1/roles
→ 创建本地 Role 记录
→ 调用 AnythingLLM /v1/workspace/{slug}/update
   更新 system_prompt

// 切换角色
POST /api/v1/roles/:id/activate
→ 更新用户的当前角色
→ 调用 AnythingLLM 更新 Workspace 配置
```

#### 4.2 多角色支持
```
用户 Workspace
├── 主角色 (默认系统提示词)
├── 角色 A (营销专家)
├── 角色 B (法务顾问)
└── 角色 C (技术支持)

// 实现方式：
- 方案 1: 每个角色一个 Workspace (资源消耗大)
- 方案 2: 对话时动态覆盖 system_prompt (推荐)
```

---

## 📊 API 变更对比

### 用户认证 (不变)
```
✅ POST /api/v1/auth/register
✅ POST /api/v1/auth/login
✅ GET /api/v1/users/me
```

### 角色管理 (小改)
```
✅ GET /api/v1/roles
✅ POST /api/v1/roles
✅ PUT /api/v1/roles/:id
✅ DELETE /api/v1/roles/:id
   + 同步更新 AnythingLLM Workspace
```

### 对话服务 (大改)
```
✅ POST /api/v1/chat-sessions          # 创建 AnythingLLM Workspace
✅ GET /api/v1/chat-sessions           # 列出用户 Workspace
✅ GET /api/v1/chat-sessions/:id       # 获取 Workspace 详情
✅ POST /api/v1/chat/:id/complete      # → AnythingLLM Chat API
✅ POST /api/v1/chat/:id/stream        # → AnythingLLM Stream API
+  GET /api/v1/chat-sessions/:id/sync  # 新增：同步历史
```

### 知识库 (大改)
```
✅ GET /api/v1/documents               # 列出 Workspace 文档
✅ POST /api/v1/documents              # → AnythingLLM Upload
✅ GET /api/v1/documents/:id           # 获取文档元数据
✅ PUT /api/v1/documents/:id           # 更新元数据
✅ DELETE /api/v1/documents/:id        # → AnythingLLM Delete
+  POST /api/v1/documents/search       # 新增：向量搜索
```

---

## 🔒 安全与隔离

### 数据隔离
- ✅ 每个用户独立 Workspace
- ✅ LanceDB 向量库物理隔离
- ✅ 文档存储隔离
- ✅ 对话历史隔离

### API 安全
- ✅ JWT Token 认证 (RoleCraft)
- ✅ Bearer Token 认证 (AnythingLLM)
- ✅ 用户 ID → Workspace Slug 映射验证

### 权限控制
```go
// 中间件验证
func WorkspaceAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        userId := c.Get("userId")
        workspaceSlug := c.Param("slug")
        
        // 验证 slug 属于当前用户
        if !validateOwnership(userId, workspaceSlug) {
            c.AbortWithStatus(403)
            return
        }
    }
}
```

---

## 📈 性能优化

### 缓存策略
```go
// 用户 Workspace 映射缓存
cache.Set("user_workspace:"+userId, slug, 5*time.Minute)

// 角色配置缓存
cache.Set("role_config:"+roleId, config, 10*time.Minute)
```

### 批量操作
```go
// 批量上传文档
POST /api/v1/documents/batch
→ 打包多个文件
→ 调用 AnythingLLM 批量上传
```

### 异步处理
```go
// 文档处理异步化
POST /api/v1/documents
→ 立即返回 "processing"
→ 后台 goroutine 调用 AnythingLLM
→ 完成后更新状态
```

---

## 🧪 测试计划

### 单元测试
- [ ] AnythingLLM Client 封装测试
- [ ] Workspace 创建/更新/删除测试
- [ ] 文档上传流程测试
- [ ] 对话 API 集成测试

### 集成测试
- [ ] 用户隔离验证 (A 不能访问 B 的数据)
- [ ] 向量检索准确性测试
- [ ] 多轮对话上下文测试
- [ ] 角色切换测试

### 性能测试
- [ ] 并发对话测试 (100 用户)
- [ ] 大批量文档上传测试
- [ ] 向量搜索延迟测试

---

## 📅 时间估算

| 阶段 | 任务 | 时间 |
|------|------|------|
| 阶段 1 | 基础架构 | 1-2 天 |
| 阶段 2 | 对话服务重构 | 1-2 天 |
| 阶段 3 | 知识库重构 | 2-3 天 |
| 阶段 4 | 角色系统增强 | 1-2 天 |
| 测试 | 完整测试 | 1-2 天 |
| **总计** | | **6-11 天** |

---

## ✅ 成功标准

- [ ] 所有现有测试通过
- [ ] 用户数据完全隔离
- [ ] 对话响应时间 <2s
- [ ] 文档上传成功率 >99%
- [ ] 向量检索准确率 >90%
- [ ] 支持 100+ 并发用户

---

**创建时间**: 2026-02-26  
**版本**: v1.0  
**状态**: 待执行
