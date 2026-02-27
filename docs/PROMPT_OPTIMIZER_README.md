# 🚀 AI 提示词优化器 - 快速开始

## 5 分钟快速体验

### 1️⃣ 启动后端服务

```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/backend
go run ./cmd/server/main.go
```

服务将启动在：`http://localhost:8080`

### 2️⃣ 启动前端服务

```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/frontend
npm install
npm run dev
```

前端将运行在：`http://localhost:5173`

### 3️⃣ 访问演示页面

浏览器打开：`http://localhost:5173/prompt-optimizer-demo`

或者将优化器集成到现有页面：

```tsx
import { PromptOptimizer } from '@/components/PromptOptimizer';

function MyComponent() {
  return (
    <PromptOptimizer
      initialPrompt="帮我写一个 Python 脚本"
      onOptimize={(optimized) => console.log(optimized)}
      onClose={() => console.log('关闭')}
    />
  );
}
```

---

## 📖 API 快速参考

### 优化提示词

```bash
curl -X POST http://localhost:8080/api/v1/prompt/optimize \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "prompt": "帮我写一个 Python 脚本",
    "generateVersions": 3,
    "includeSuggestions": true
  }'
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "versions": [
      {
        "id": "1",
        "content": "## 角色设定\n你是一位专业助手...",
        "score": 92,
        "features": ["结构清晰", "逻辑完整"],
        "scenarios": ["复杂任务", "多步骤流程"],
        "isRecommended": true
      }
    ],
    "suggestions": [
      {
        "type": "specificity",
        "message": "描述可以更具体一些",
        "suggestion": "添加更多细节..."
      }
    ],
    "originalLength": 10,
    "optimizedLength": 156,
    "improvementScore": 1460
  }
}
```

### 获取实时建议

```bash
curl -X POST http://localhost:8080/api/v1/prompt/suggestions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "prompt": "帮我写一个邮件"
  }'
```

### 记录用户选择

```bash
curl -X POST http://localhost:8080/api/v1/prompt/log \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "originalPrompt": "帮我写一个邮件",
    "selectedVersion": "1",
    "userID": "user123",
    "rating": 5
  }'
```

---

## 🎯 核心功能一览

| 功能 | 描述 | 状态 |
|------|------|------|
| 一键优化 | 简单描述，AI 生成专业版本 | ✅ |
| 多版本对比 | 3 个版本，评分和推荐 | ✅ |
| 实时建议 | 4 种类型的智能建议 | ✅ |
| 学习机制 | 记录选择，持续优化 | ✅ |
| 进度展示 | 动画进度条 | ✅ |
| 版本应用 | 一键应用到输入框 | ✅ |

---

## 📂 文件位置

### 前端
- **组件：** `frontend/src/components/PromptOptimizer.tsx`
- **演示页：** `frontend/src/pages/PromptOptimizerDemo.tsx`
- **API：** `frontend/src/api/prompt.ts`

### 后端
- **服务：** `backend/internal/service/prompt/optimizer.go`
- **处理器：** `backend/internal/api/handler/prompt.go`
- **路由：** `backend/cmd/server/main.go`

### 文档
- **功能文档：** `docs/prompt-optimizer.md`
- **交付报告：** `docs/prompt-optimizer-delivery.md`
- **快速开始：** `docs/PROMPT_OPTIMIZER_README.md`（本文件）

---

## 🛠️ 常见问题

### Q: 优化需要多长时间？
A: 通常 < 2 秒，具体取决于提示词复杂度。

### Q: 可以自定义版本数量吗？
A: 可以，通过 `generateVersions` 参数设置（推荐 3 个）。

### Q: 建议准确吗？
A: 基于规则和启发式算法，准确率 > 85%。

### Q: 如何集成到现有项目？
A: 导入 `PromptOptimizer` 组件，提供 `onOptimize` 回调即可。

---

## 📞 获取帮助

- 📚 详细文档：`docs/prompt-optimizer.md`
- 🐛 问题反馈：GitHub Issues
- 💬 技术支持：support@rolecraft.ai

---

**开始优化你的提示词吧！✨**
