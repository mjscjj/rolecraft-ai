# ChatStream Component

流式聊天组件，支持实时流式输出、Markdown 渲染和代码高亮。

## 文件结构

```
ChatStream/
├── index.jsx          # 主组件，包含流式聊天逻辑
├── MessageList.jsx    # 消息列表组件，包含智能滚动
├── MessageBubble.jsx  # 消息气泡组件，支持 Markdown 和代码高亮
├── TypingIndicator.jsx # 打字指示器组件
├── styles.css         # 样式文件
└── README.md          # 说明文档
```

## 功能特性

### 1. 流式响应处理
- 使用 SSE (Server-Sent Events) 协议接收流式数据
- 实时解析 `data: {"content": "..."}` 格式
- 逐步更新 UI，提供打字机效果

### 2. 智能滚动优化
- 自动滚动到底部（当用户在底部时）
- 检测用户手动滚动（不干扰用户阅读）
- 流式完成后自动滚动
- 显示"新消息"按钮（当用户滚动上去时）

### 3. Markdown 渲染
- 使用 `react-markdown` 渲染 Markdown 内容
- 支持 GFM (GitHub Flavored Markdown)
- 支持表格、任务列表、删除线等

### 4. 代码高亮
- 使用 `react-syntax-highlighter` 进行代码高亮
- 支持多种编程语言
- 用户消息使用暗色主题
- AI 消息使用亮色主题

### 5. 消息功能
- 复制消息内容
- 重新生成回复
- 点赞/点踩反馈
- 来源引用展示

## 使用方法

```jsx
import { ChatStream } from './components/ChatStream';

function App() {
  return (
    <ChatStream 
      roleId="role-123" 
      roleName="AI 助手" 
    />
  );
}
```

## API 接口

组件调用以下 API：

1. **创建会话**: `POST /api/v1/chat-sessions`
   ```json
   {
     "roleId": "string",
     "mode": "quick"
   }
   ```

2. **流式聊天**: `POST /api/v1/chat/{sessionId}/stream`
   ```json
   {
     "content": "用户消息"
   }
   ```
   
   响应格式 (SSE):
   ```
   data: {"content": "部分响应"}
   data: {"content": "更多响应"}
   ```

## 样式定制

所有样式都在 `styles.css` 中定义，使用 BEM 命名规范：
- `.chat-stream-container` - 主容器
- `.chat-stream-header` - 头部
- `.chat-stream-messages` - 消息列表
- `.chat-stream-message` - 单条消息
- `.chat-stream-bubble` - 消息气泡
- `.chat-stream-input-area` - 输入区域

可以通过覆盖 CSS 变量或修改样式文件来定制主题。

## 依赖

```json
{
  "react-markdown": "^9.x",
  "remark-gfm": "^4.x",
  "react-syntax-highlighter": "^15.x",
  "lucide-react": "^0.x"
}
```

## 注意事项

1. 需要后端支持 SSE 流式响应
2. 需要设置 `token` 到 `localStorage` 进行认证
3. 消息气泡最大宽度为 70%
4. 输入框自动调整高度（最大 200px）
