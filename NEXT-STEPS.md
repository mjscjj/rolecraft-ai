# RoleCraft AI - 下一步行动计划

**创建时间**: 2026-02-27 17:30  
**优先级**: P0 > P1 > P2

---

## 🎯 本周目标 (2026-02-27 ~ 2026-03-05)

**核心目标**: 完成知识库 MVP，整体进度达到 85%

---

## P0 - 必须完成

### 任务 1: 知识库基础功能 ⭐⭐⭐⭐⭐

**优先级**: P0  
**预计工时**: 2 天  
**负责人**: AI Assistant

#### 后端 (1 天)

**API 开发**:
- [ ] `POST /api/v1/documents` - 上传文档
- [ ] `GET /api/v1/documents` - 获取文档列表
- [ ] `GET /api/v1/documents/:id` - 获取文档详情
- [ ] `DELETE /api/v1/documents/:id` - 删除文档
- [ ] `PUT /api/v1/documents/:id` - 更新文档

**服务层**:
- [ ] 文件上传服务
- [ ] 文件存储服务
- [ ] 文档解析服务（TXT/Markdown）

**数据模型**:
```go
type Document struct {
    ID          string    `json:"id"`
    UserID      string    `json:"user_id"`
    Title       string    `json:"title"`
    FileName    string    `json:"file_name"`
    FileSize    int64     `json:"file_size"`
    FileType    string    `json:"file_type"` // pdf, docx, txt, md
    Content     string    `json:"content"`   // 解析后的文本
    Status      string    `json:"status"`    // processing, ready, error
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

**文件位置**:
- `backend/internal/api/handler/document.go`
- `backend/internal/service/document/service.go`
- `backend/internal/models/document.go`

#### 前端 (1 天)

**页面开发**:
- [ ] 知识库列表页面
- [ ] 上传对话框
- [ ] 文档详情页面

**组件开发**:
- [ ] 文档卡片组件
- [ ] 上传进度组件
- [ ] 文件类型图标

**文件位置**:
- `frontend/src/pages/KnowledgeBase.tsx` (已有，需完善)
- `frontend/src/components/DocumentCard.tsx`
- `frontend/src/components/UploadDialog.tsx`

#### 验收标准

- [ ] 可以上传 TXT/Markdown 文件
- [ ] 文档列表正确显示
- [ ] 可以删除文档
- [ ] 文件大小限制（10MB）
- [ ] 错误处理完善

---

### 任务 2: 对话历史管理 ⭐⭐⭐⭐

**优先级**: P0  
**预计工时**: 1 天  
**负责人**: AI Assistant

#### 后端 (0.5 天)

**API 开发**:
- [ ] `GET /api/v1/chat-sessions` - 获取会话列表（已有）
- [ ] `DELETE /api/v1/chat-sessions/:id` - 删除会话
- [ ] `PUT /api/v1/chat-sessions/:id/title` - 重命名会话
- [ ] `POST /api/v1/chat-sessions/:id/archive` - 归档会话

**文件位置**:
- `backend/internal/api/handler/chat.go`

#### 前端 (0.5 天)

**页面开发**:
- [ ] 对话历史列表（侧边栏）
- [ ] 历史搜索功能
- [ ] 删除确认对话框

**文件位置**:
- `frontend/src/components/ChatHistory.tsx`
- `frontend/src/pages/Chat.tsx` (集成)

#### 验收标准

- [ ] 显示所有历史对话
- [ ] 可以搜索对话
- [ ] 可以删除对话
- [ ] 可以重命名对话

---

### 任务 3: 角色市场 ⭐⭐⭐⭐

**优先级**: P0  
**预计工时**: 1 天  
**负责人**: AI Assistant

#### 后端 (0.5 天)

**API 开发**:
- [ ] `GET /api/v1/roles/templates` - 获取模板列表（已有）
- [ ] `GET /api/v1/roles/templates/:id` - 获取模板详情
- [ ] `GET /api/v1/roles/search` - 搜索角色

**文件位置**:
- `backend/internal/api/handler/role.go`

#### 前端 (0.5 天)

**页面开发**:
- [ ] 角色市场页面
- [ ] 搜索筛选组件
- [ ] 角色详情弹窗

**文件位置**:
- `frontend/src/pages/RoleMarket.tsx` (已有，需完善)

#### 验收标准

- [ ] 显示所有角色模板
- [ ] 可以搜索角色
- [ ] 可以按分类筛选
- [ ] 一键使用模板

---

## P1 - 重要功能

### 任务 4: 深度思考集成 ⭐⭐⭐⭐

**优先级**: P1  
**预计工时**: 1-2 天  
**负责人**: AI Assistant

#### 后端 (1 天)

**功能开发**:
- [ ] 思考过程提取（从 AI 回复）
- [ ] 流式推送思考步骤
- [ ] 思考配置 API

**文件位置**:
- `backend/internal/service/thinking/extractor.go` (已有)
- `backend/internal/api/handler/chat.go` (集成)

#### 前端 (1 天)

**功能集成**:
- [ ] 集成到 Chat 组件
- [ ] 思考展示开关
- [ ] 思考步骤动画

**文件位置**:
- `frontend/src/pages/Chat.tsx`
- `frontend/src/components/Thinking/ThinkingDisplay.tsx` (已有)

#### 验收标准

- [ ] 对话时显示思考过程
- [ ] 可以开关思考展示
- [ ] 思考步骤流畅展示

---

### 任务 5: 设置页面 ⭐⭐⭐

**优先级**: P1  
**预计工时**: 1 天  
**负责人**: AI Assistant

#### 前端 (1 天)

**页面开发**:
- [ ] 个人资料设置
- [ ] 修改密码
- [ ] API Key 配置
- [ ] 模型选择
- [ ] 主题切换

**文件位置**:
- `frontend/src/pages/Settings.tsx`

#### 验收标准

- [ ] 可以修改个人资料
- [ ] 可以修改密码
- [ ] 可以切换主题

---

### 任务 6: 知识库高级功能 ⭐⭐⭐⭐

**优先级**: P1  
**预计工时**: 1-2 天  
**负责人**: AI Assistant

#### 后端 (1 天)

**功能开发**:
- [ ] 文档搜索（全文检索）
- [ ] 文档预览（PDF/Word）
- [ ] 文件夹管理
- [ ] 批量操作

**文件位置**:
- `backend/internal/service/document/search.go`

#### 前端 (1 天)

**功能开发**:
- [ ] 搜索框
- [ ] 文档预览组件
- [ ] 文件夹树
- [ ] 批量选择

#### 验收标准

- [ ] 可以搜索文档内容
- [ ] 可以预览文档
- [ ] 可以创建文件夹
- [ ] 可以批量删除

---

## P2 - 完善功能

### 任务 7: 数据分析 ⭐⭐⭐

**优先级**: P2  
**预计工时**: 2 天

**功能**:
- [ ] 对话统计图表
- [ ] 角色使用频率
- [ ] Token 消耗统计
- [ ] 导出报告

---

### 任务 8: 消息操作 ⭐⭐⭐

**优先级**: P2  
**预计工时**: 1 天

**功能**:
- [ ] 编辑消息
- [ ] 重新生成回复
- [ ] 复制消息
- [ ] 分享对话

---

### 任务 9: 对话导出 ⭐⭐

**优先级**: P2  
**预计工时**: 1 天

**功能**:
- [ ] 导出 Markdown
- [ ] 导出 PDF
- [ ] 分享链接

---

## 📅 每日站会

### 每日检查

**时间**: 每天晚上 23:00

**检查项**:
1. 今天完成了什么？
2. 遇到什么问题？
3. 明天计划做什么？
4. 需要帮助吗？

**记录位置**: `memory/YYYY-MM-DD.md`

---

## 🎯 成功标准

### 本周成功标准 (2026-03-05)

- [ ] 知识库可以上传/查看/删除文档
- [ ] 对话历史可以管理
- [ ] 角色市场可以使用
- [ ] 整体进度达到 85%
- [ ] 无严重 Bug

### 本月成功标准 (2026-03-27)

- [ ] 所有 P0 功能完成
- [ ] 所有 P1 功能完成
- [ ] 整体进度达到 95%
- [ ] 可以进行 Beta 测试
- [ ] 性能指标达标

---

## 📊 进度追踪

### 本周进度

| 日期 | 完成的任务 | 进度 |
|------|-----------|------|
| 02-27 | 核心功能完善 | 70% |
| 02-28 | - | - |
| 03-01 | - | - |
| 03-02 | - | - |
| 03-03 | - | - |
| 03-04 | - | - |
| 03-05 | 目标：知识库 MVP | 85% |

### 累计进度

- **总任务**: 20 个
- **已完成**: 7 个 (35%)
- **进行中**: 0 个
- **待开始**: 13 个 (65%)

---

## 🚨 风险与问题

### 已知风险

1. **时间不足**
   - 风险：P0 任务可能无法全部完成
   - 缓解：优先保证知识库功能

2. **技术难度**
   - 风险：文档解析可能遇到技术问题
   - 缓解：先支持 TXT/Markdown，PDF/Word 后续

3. **性能问题**
   - 风险：大文件上传可能超时
   - 缓解：限制文件大小，使用分片上传

### 需要决策

1. **文档存储方式**
   - 选项 A: 本地存储
   - 选项 B: 对象存储（MinIO）
   - 建议：开发环境本地，生产环境 MinIO

2. **知识库检索方式**
   - 选项 A: 简单关键词匹配
   - 选项 B: 向量检索（需要 Milvus）
   - 建议：先用 A，后续升级 B

---

## 📝 会议记录

### 2026-02-27 项目会议

**参会人员**: AI Assistant, User  
**会议时间**: 2026-02-27 17:30

**讨论内容**:
1. 当前项目进度 70%
2. 核心功能已完成
3. 下一步重点是知识库

**决策**:
1. 优先开发知识库功能
2. 本周目标达到 85%
3. 记录项目进度

**待办**:
- [ ] 创建项目进度文档 ✅
- [ ] 创建下一步行动计划 ✅
- [ ] 开始知识库开发

---

**文档维护**: 每次重大更新后更新此文档  
**下次更新**: 2026-02-28 或完成 P0 任务后
