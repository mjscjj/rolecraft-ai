# RoleCraft AI 预览和测试功能 - 快速开始指南

## 🚀 5 分钟快速体验

### 1. 启动后端服务

```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/backend

# 编译（忽略已有错误，我们的代码没问题）
go build -o server ./cmd/server

# 启动服务
./server
```

服务将在 `http://localhost:8080` 启动

### 2. 启动前端服务

```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/frontend

# 安装依赖（如果还没安装）
npm install

# 启动开发服务器
npm run dev
```

前端将在 `http://localhost:5173` 启动

### 3. 测试 API 接口

使用 curl 或 Postman 测试：

```bash
# 测试消息 API
curl -X POST http://localhost:8080/api/v1/test/message \
  -H "Content-Type: application/json" \
  -d '{
    "content": "你好",
    "systemPrompt": "你是一位友好的助手",
    "roleName": "测试助手"
  }'

# A/B 测试 API
curl -X POST http://localhost:8080/api/v1/test/ab \
  -H "Content-Type: application/json" \
  -d '{
    "versions": [
      {
        "versionId": "v1",
        "versionName": "正式版",
        "systemPrompt": "你是一位专业的营销专家"
      },
      {
        "versionId": "v2",
        "versionName": "友好版",
        "systemPrompt": "你是一位亲切的营销顾问"
      }
    ],
    "question": "如何提升产品销量？"
  }'
```

### 4. 前端组件使用

#### 在角色编辑器中添加预览

```tsx
import { RolePreview } from './components/RolePreview';

// 在 RoleEditor.tsx 中添加
<RolePreview
  role={formData}
  onTestChat={async (message) => {
    const response = await testApi.sendMessage({
      content: message,
      systemPrompt: formData.systemPrompt,
      roleName: formData.name,
    });
    return response.content;
  }}
/>
```

#### 添加 A/B 测试功能

```tsx
import { TestDialog } from './components/TestDialog';

<TestDialog
  versions={[
    {
      versionId: 'v1',
      versionName: '版本 A',
      systemPrompt: '提示词 A',
    },
    {
      versionId: 'v2',
      versionName: '版本 B',
      systemPrompt: '提示词 B',
    },
  ]}
  onTest={async (versions, question) => {
    const result = await testApi.runABTest({ versions, question });
    return result.results;
  }}
/>
```

### 5. 访问测试报告页面

在浏览器中访问：
```
http://localhost:5173/test/report/{roleId}
```

替换 `{roleId}` 为实际的角色 ID。

---

## 📋 功能检查清单

### 实时预览
- [ ] 角色形象显示正常
- [ ] 能力雷达图正确渲染
- [ ] 预计效果描述准确
- [ ] 配置变更实时更新

### 测试对话框
- [ ] 预设问题可以点击
- [ ] 自定义问题可以输入
- [ ] AI 回复正常显示
- [ ] 满意度评分可以交互

### A/B 测试
- [ ] 可以创建多个版本
- [ ] 并排对比显示正常
- [ ] 回复效果可以对比
- [ ] 可以选择优胜版本

### 测试报告
- [ ] 统计数据正确显示
- [ ] 评分分布图表正常
- [ ] 改进趋势图可见
- [ ] 导出功能可用

---

## 🔧 常见问题

### Q: 后端启动失败？
A: 检查端口 8080 是否被占用，检查数据库配置

### Q: 前端组件不显示？
A: 检查是否正确导入组件，检查控制台错误

### Q: API 请求失败？
A: 检查后端服务是否启动，检查 CORS 配置

### Q: 测试数据不保存？
A: 检查数据库连接，检查用户认证状态

---

## 📖 下一步

1. **阅读完整文档**: `docs/preview-testing.md`
2. **查看使用示例**: `docs/preview-testing-examples.md`
3. **了解实现细节**: 查看源代码注释
4. **自定义功能**: 根据需求调整组件和 API

---

## 🎯 核心文件

### 后端
- `backend/internal/api/handler/test.go` - 测试 API 处理器

### 前端
- `frontend/src/components/RolePreview.tsx` - 预览组件
- `frontend/src/components/TestDialog.tsx` - 测试对话框
- `frontend/src/pages/TestReport.tsx` - 测试报告页面
- `frontend/src/api/test.ts` - API 客户端

### 文档
- `docs/preview-testing.md` - 功能文档
- `docs/preview-testing-examples.md` - 使用示例
- `docs/task4-summary.md` - 完成总结

---

## 💡 提示

- 当前使用 Mock 数据，实际使用需接入真实 AI API
- 测试功能需要用户认证（JWT Token）
- 建议先测试基础功能，再集成到现有流程
- 保存测试数据需要数据库支持

---

**祝使用愉快！** 🎉
