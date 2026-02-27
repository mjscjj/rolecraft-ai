# RoleCraft AI 深度思考模块 - 使用文档

**创建日期**: 2026-02-27  
**版本**: 1.0.0  
**状态**: ✅ 已完成

---

## 📋 目录

1. [功能概述](#功能概述)
2. [架构设计](#架构设计)
3. [后端使用](#后端使用)
4. [前端使用](#前端使用)
5. [思考步骤类型](#思考步骤类型)
6. [API 接口](#api 接口)
7. [示例代码](#示例代码)
8. [常见问题](#常见问题)

---

## 功能概述

深度思考模块让 AI 在回答问题前展示其思考过程，提升用户体验和产品"智能感"。

### 核心特性

✅ **渐进式展示** - 流式展示思考的每一步  
✅ **6 种思考类型** - 理解、分析、检索、组织、结论、灵感  
✅ **实时反馈** - 显示思考时长和进度  
✅ **可折叠设计** - 用户可控制查看细节  
✅ **美观动画** - 流畅的过渡效果  
✅ **响应式支持** - 适配移动端和桌面端  

---

## 架构设计

```
┌─────────────────────────────────────────┐
│          前端 (React + TypeScript)      │
├─────────────────────────────────────────┤
│  ThinkingDisplay 组件                    │
│  ├─ ThinkingStepItem (步骤项)           │
│  └─ ThinkingDisplay.css (样式)          │
└─────────────────────────────────────────┘
                    ↕ SSE (Server-Sent Events)
┌─────────────────────────────────────────┐
│          后端 (Go + Gin)                │
├─────────────────────────────────────────┤
│  ChatHandler.ChatStreamWithThinking     │
│  ├─ thinking.Service                    │
│  └─ thinking.Extractor                  │
└─────────────────────────────────────────┘
```

---

## 后端使用

### 1. 导入思考服务

```go
import "rolecraft-ai/internal/service/thinking"
```

### 2. 创建思考服务

```go
thinkingSvc := thinking.NewService()
```

### 3. 使用流式发送器

```go
// 创建流式发送器
sender := thinkingSvc.NewStreamThinkingSender(func(chunk thinking.StreamChunk) {
    // 通过 SSE 发送到客户端
    jsonData, _ := json.Marshal(chunk)
    fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
    flusher.Flush()
})

// 逐步添加思考步骤
sender.AddThinkingStep(thinking.ThinkingUnderstand, "理解用户问题")
sender.AddThinkingStep(thinking.ThinkingAnalyze, "分析关键要素")
sender.AddThinkingStep(thinking.ThinkingSearch, "检索相关知识")
sender.AddThinkingStep(thinking.ThinkingOrganize, "组织答案结构")
sender.AddThinkingStep(thinking.ThinkingConclude, "得出结论")

// 完成思考过程
sender.Complete()

// 发送最终答案
sender.SendAnswer(responseContent)
```

### 4. 思考步骤类型

```go
thinking.ThinkingUnderstand  // 🤔 理解问题
thinking.ThinkingAnalyze     // 🔍 分析要素
thinking.ThinkingSearch      // 📚 检索知识
thinking.ThinkingOrganize    // 📝 组织答案
thinking.ThinkingConclude    // ✅ 得出结论
thinking.ThinkingInsight     // 💡 灵感闪现
```

### 5. 提取已有思考过程

```go
extractor := thinking.NewExtractor()
result := extractor.Extract(responseContent)

if result.HasThinking {
    // 处理思考过程
    process := result.ThinkingProcess
    answer := result.FinalAnswer
}
```

---

## 前端使用

### 1. 导入组件

```typescript
import ThinkingDisplay, { ThinkingProcess } from './Thinking/ThinkingDisplay';
```

### 2. 基本使用

```tsx
<ThinkingDisplay
  thinkingProcess={thinkingProcess}
  isStreaming={isStreaming}
  defaultExpanded={true}
  onToggle={(expanded) => console.log('Toggled:', expanded)}
/>
```

### 3. 思考过程数据结构

```typescript
interface ThinkingProcess {
  steps: ThinkingStep[];
  startTime: number;      // Unix timestamp (ms)
  endTime?: number;
  duration: number;       // 总耗时（秒）
  isComplete: boolean;
}

interface ThinkingStep {
  id: string;
  type: ThinkingStepType;  // 'understand' | 'analyze' | 'search' | 'organize' | 'conclude' | 'insight'
  content: string;
  timestamp: number;
  status: 'pending' | 'processing' | 'completed';
  icon: string;
  duration?: number;       // 步骤耗时（秒）
}
```

### 4. 流式接收示例

```typescript
const response = await fetch('/api/chat/stream-with-thinking', {
  method: 'POST',
  body: JSON.stringify({ content: message }),
});

const reader = response.body.getReader();
const decoder = new TextDecoder();
let thinkingSteps: ThinkingStep[] = [];

while (true) {
  const { done, value } = await reader.read();
  if (done) break;

  const chunk = decoder.decode(value);
  const lines = chunk.split('\n');

  for (const line of lines) {
    if (line.startsWith('data: ')) {
      const data = JSON.parse(line.slice(6));
      
      if (data.type === 'thinking') {
        thinkingSteps.push(data.data);
        // 更新 UI
      }
      
      if (data.type === 'answer') {
        // 显示最终答案
      }
      
      if (data.type === 'done') {
        // 完成
      }
    }
  }
}
```

---

## 思考步骤类型

| 类型 | 图标 | 颜色 | 用途 |
|------|------|------|------|
| understand | 🤔 | #667eea | 理解用户问题 |
| analyze | 🔍 | #764ba2 | 分析关键要素 |
| search | 📚 | #f093fb | 检索知识 |
| organize | 📝 | #f5576c | 组织答案 |
| conclude | ✅ | #4facfe | 得出结论 |
| insight | 💡 | #43e97b | 灵感闪现 |

---

## API 接口

### POST /api/v1/chat/:id/stream-with-thinking

**描述**: 发送消息并接收带思考过程的流式响应

**请求**:
```json
{
  "content": "用户问题"
}
```

**响应** (SSE):
```
data: {"type":"thinking","data":{"id":"1","type":"understand","content":"理解问题","status":"processing"}}

data: {"type":"thinking","data":{"id":"1","type":"understand","content":"理解问题","status":"completed","duration":0.3}}

data: {"type":"answer","data":{"content":"最终答案"}}

data: {"type":"done","done":true}
```

---

## 示例代码

### 完整后端示例

```go
func (h *ChatHandler) ChatStreamWithThinking(c *gin.Context) {
    sessionId := c.Param("id")
    var req SendMessageRequest
    c.ShouldBindJSON(&req)

    // 设置 SSE
    c.Header("Content-Type", "text/event-stream")
    flusher, _ := c.Writer.(http.Flusher)

    // 创建思考发送器
    sender := h.thinkingSvc.NewStreamThinkingSender(func(chunk thinking.StreamChunk) {
        jsonData, _ := json.Marshal(chunk)
        fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
        flusher.Flush()
    })

    // 流式思考步骤
    sender.AddThinkingStep(thinking.ThinkingUnderstand, "理解："+req.Content[:30])
    sender.AddThinkingStep(thinking.ThinkingAnalyze, "分析要素")
    sender.AddThinkingStep(thinking.ThinkingSearch, "检索知识")
    sender.AddThinkingStep(thinking.ThinkingOrganize, "组织答案")
    sender.AddThinkingStep(thinking.ThinkingConclude, "得出结论")
    
    // 完成思考
    sender.Complete()
    
    // 发送答案
    answer := generateAnswer(req.Content)
    sender.SendAnswer(answer)
    
    // 完成
    jsonData, _ := json.Marshal(thinking.StreamChunk{Type: "done", Done: true})
    fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
}
```

### 完整前端示例

```tsx
import React, { useState } from 'react';
import ThinkingDisplay from './Thinking/ThinkingDisplay';

const ChatComponent = () => {
  const [thinkingProcess, setThinkingProcess] = useState<ThinkingProcess | null>(null);
  const [isStreaming, setIsStreaming] = useState(false);
  const [answer, setAnswer] = useState('');

  const handleSend = async (message: string) => {
    setIsStreaming(true);
    setThinkingProcess(null);
    setAnswer('');

    const response = await fetch('/api/chat/stream-with-thinking', {
      method: 'POST',
      body: JSON.stringify({ content: message }),
    });

    const reader = response.body.getReader();
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value);
      const lines = chunk.split('\n');

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = JSON.parse(line.slice(6));
          
          if (data.type === 'thinking') {
            setThinkingProcess(prev => ({
              steps: [...(prev?.steps || []), data.data],
              startTime: prev?.startTime || Date.now(),
              duration: (Date.now() - (prev?.startTime || Date.now())) / 1000,
              isComplete: false,
            }));
          }
          
          if (data.type === 'answer') {
            setAnswer(data.data.content);
          }
          
          if (data.type === 'done') {
            setIsStreaming(false);
            setThinkingProcess(prev => prev ? { ...prev, isComplete: true } : null);
          }
        }
      }
    }
  };

  return (
    <div>
      {thinkingProcess && (
        <ThinkingDisplay
          thinkingProcess={thinkingProcess}
          isStreaming={isStreaming}
        />
      )}
      {answer && <div className="answer">{answer}</div>}
    </div>
  );
};
```

---

## 常见问题

### Q1: 思考过程太长怎么办？

**A**: 使用折叠功能，默认只显示标题，用户点击展开查看详情。

```tsx
<ThinkingDisplay defaultExpanded={false} />
```

### Q2: 如何自定义思考步骤？

**A**: 修改 `STEP_CONFIG` 配置：

```typescript
const STEP_CONFIG = {
  custom: { label: '自定义步骤', icon: '🎯', color: '#ff6b6b' },
};
```

### Q3: 思考时长如何计算？

**A**: 自动计算，从 `startTime` 到当前时间（流式中）或 `endTime`（完成后）。

### Q4: 支持暗色模式吗？

**A**: 支持，CSS 中已包含 `@media (prefers-color-scheme: dark)` 样式。

### Q5: 如何禁用某个思考步骤？

**A**: 在后端调用时跳过该步骤即可：

```go
// 只使用 3 个步骤
sender.AddThinkingStep(thinking.ThinkingUnderstand, "...")
sender.AddThinkingStep(thinking.ThinkingAnalyze, "...")
sender.AddThinkingStep(thinking.ThinkingConclude, "...")
```

---

## 文件清单

### 后端
- ✅ `backend/internal/service/thinking/extractor.go` - 思考提取器
- ✅ `backend/internal/service/thinking/service.go` - 思考服务
- ✅ `backend/internal/api/handler/chat.go` - 集成到 ChatHandler

### 前端
- ✅ `frontend/src/components/Thinking/ThinkingDisplay.tsx` - 主组件
- ✅ `frontend/src/components/Thinking/ThinkingDisplay.css` - 样式
- ✅ `frontend/src/components/Thinking/ChatWithThinkingExample.tsx` - 使用示例

### 文档
- ✅ `THINKING-MODULE-README.md` - 本文档

---

## 下一步优化

- [ ] 支持思考过程编辑
- [ ] 添加思考模板库
- [ ] 支持多分支思考
- [ ] 思考质量评估
- [ ] 导出思考过程

---

**开发团队**: RoleCraft AI  
**联系方式**: support@rolecraft.ai  
**文档版本**: v1.0.0
