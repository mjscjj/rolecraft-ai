# 深度思考模块 - 快速集成指南

**5 分钟快速上手** ⚡

---

## 🚀 第一步：后端集成 (2 分钟)

### 1. 注册 API 路由

在 `cmd/server/main.go` 中添加新路由：

```go
// 找到 chat 相关路由注册部分
authorized.POST("/chat/:id/stream", chatHandler.ChatStream)

// 添加这一行：
authorized.POST("/chat/:id/stream-with-thinking", chatHandler.ChatStreamWithThinking)
```

### 2. 重启后端服务

```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/backend
go run cmd/server/main.go
```

### 3. 测试 API

```bash
curl -X POST http://localhost:8080/api/v1/chat/SESSION_ID/stream-with-thinking \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"content": "你好"}'
```

---

## 🎨 第二步：前端集成 (3 分钟)

### 1. 复制组件文件

```bash
# 组件已创建在：
frontend/src/components/Thinking/ThinkingDisplay.tsx
frontend/src/components/Thinking/ThinkingDisplay.css
```

### 2. 在 Chat 页面中导入

```typescript
// 在 Chat.tsx 或 ChatWebUI.tsx 中
import ThinkingDisplay from './Thinking/ThinkingDisplay';
import { ThinkingProcess } from './Thinking/ThinkingDisplay';
```

### 3. 修改消息数据结构

```typescript
interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  thinkingProcess?: ThinkingProcess;  // 新增
  isStreaming?: boolean;
}
```

### 4. 在消息渲染中使用

```tsx
{messages.map((message) => (
  <div key={message.id} className="message">
    {/* 思考过程展示 */}
    {message.thinkingProcess && (
      <ThinkingDisplay
        thinkingProcess={message.thinkingProcess}
        isStreaming={message.isStreaming}
      />
    )}
    
    {/* 消息内容 */}
    <div className="message-content">
      {message.content}
    </div>
  </div>
))}
```

### 5. 修改流式接收逻辑

```typescript
// 在 handleStreamChat 函数中
while (true) {
  const { done, value } = await reader.read();
  if (done) break;

  const chunk = decoder.decode(value);
  const lines = chunk.split('\n');

  for (const line of lines) {
    if (line.startsWith('data: ')) {
      const data = JSON.parse(line.slice(6));
      
      // 处理思考步骤
      if (data.type === 'thinking') {
        setMessages(prev => prev.map(msg => {
          if (msg.id === currentAiMessageId) {
            const steps = [...(msg.thinkingProcess?.steps || []), data.data];
            return {
              ...msg,
              thinkingProcess: {
                steps,
                startTime: msg.thinkingProcess?.startTime || Date.now(),
                duration: (Date.now() - (msg.thinkingProcess?.startTime || Date.now())) / 1000,
                isComplete: false,
              }
            };
          }
          return msg;
        }));
      }
      
      // 处理最终答案
      if (data.type === 'answer') {
        // 更新 message.content
      }
      
      // 处理完成
      if (data.type === 'done') {
        setMessages(prev => prev.map(msg => {
          if (msg.id === currentAiMessageId) {
            return {
              ...msg,
              isStreaming: false,
              thinkingProcess: msg.thinkingProcess ? {
                ...msg.thinkingProcess,
                isComplete: true,
              } : undefined
            };
          }
          return msg;
        }));
      }
    }
  }
}
```

### 6. 修改 API 调用

```typescript
// 将 API 端点改为带 thinking 的版本
const response = await fetch(`/api/v1/chat/${sessionId}/stream-with-thinking`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  },
  body: JSON.stringify({ content: message }),
});
```

---

## ✅ 完成！

现在你的 AI 对话就会显示思考过程了！

---

## 🎯 效果预览

```
┌─────────────────────────────────────────┐
│ 🧠 深度思考中... (3 步，2.3s)      [收起] │
├─────────────────────────────────────────┤
│ ████████████░░░░░░░░░░ 60%              │
│                                         │
│ 🤔 理解问题                    ✓ 0.3s   │
│    理解用户问题：你好                   │
│                                         │
│ 🔍 分析要素                    ✓ 0.5s   │
│    分析关键要素和约束条件               │
│                                         │
│ 📚 检索知识                    ⚙️ 处理中 │
│    从知识库检索相关信息                 │
│                                         │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ 你好！我是 AI 助手，很高兴为你服务！      │
└─────────────────────────────────────────┘
```

---

## 🔧 可选配置

### 1. 默认折叠思考

```tsx
<ThinkingDisplay
  thinkingProcess={thinkingProcess}
  defaultExpanded={false}  // 默认折叠
/>
```

### 2. 禁用思考显示

```tsx
{SHOW_THINKING && message.thinkingProcess && (
  <ThinkingDisplay ... />
)}
```

### 3. 自定义思考步骤

在后端自定义步骤数量和类型：

```go
sender.AddThinkingStep(thinking.ThinkingUnderstand, "自定义步骤 1")
sender.AddThinkingStep(thinking.ThinkingInsight, "自定义步骤 2")
sender.AddThinkingStep(thinking.ThinkingConclude, "自定义步骤 3")
```

---

## 📱 移动端适配

组件已内置响应式设计，无需额外配置：

```css
/* 自动适配 <768px 屏幕 */
@media (max-width: 768px) {
  .thinking-header { padding: 10px 12px; }
  .thinking-step { padding: 10px; }
}
```

---

## 🌙 暗色模式

组件已支持系统暗色模式，自动适配：

```css
@media (prefers-color-scheme: dark) {
  .thinking-display { 
    background: linear-gradient(135deg, #2d3748 0%, #1a202c 100%);
  }
}
```

---

## 🐛 常见问题

### Q: 思考过程不显示？

**A**: 检查：
1. 后端是否返回 `type: "thinking"` 数据
2. 前端是否正确解析 SSE 数据
3. `thinkingProcess` 状态是否更新

### Q: 样式错乱？

**A**: 确保 CSS 文件已正确导入：

```typescript
import './Thinking/ThinkingDisplay.css';
```

### Q: 流式更新卡顿？

**A**: 优化建议：
1. 使用 `React.memo` 包装组件
2. 添加防抖处理
3. 减少不必要的状态更新

---

## 📚 完整文档

- [THINKING-MODULE-README.md](./THINKING-MODULE-README.md) - 详细使用文档
- [THINKING-IMPLEMENTATION-SUMMARY.md](./THINKING-IMPLEMENTATION-SUMMARY.md) - 实施总结

---

## 🎉 开始使用吧！

现在你的 RoleCraft AI 已经具备业界领先的深度思考展示能力了！

**下一步**: 
1. 测试功能
2. 收集用户反馈
3. 根据反馈优化

**有问题？** 查看完整文档或联系开发团队。

---

**最后更新**: 2026-02-27  
**版本**: v1.0.0  
**状态**: ✅ 生产就绪
