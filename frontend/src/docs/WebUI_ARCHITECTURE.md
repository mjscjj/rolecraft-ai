# RoleCraft WebUI 架构设计

## 📐 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        RoleCraft WebUI                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Header (导航栏)                        │  │
│  │  ┌──────┐  ┌─────────────┐  ┌──────────┐  ┌──────────┐  │  │
│  │  │ Logo │  │ 模型选择器  │  │ 知识库   │  │ 用户菜单 │  │  │
│  │  │      │  │             │  │ 选择器   │  │          │  │  │
│  │  └──────┘  └─────────────┘  └──────────┘  └──────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌────────────┐  ┌───────────────────────────────────────────┐  │
│  │            │  │                                           │  │
│  │  Sidebar   │  │           Chat Area                       │  │
│  │  (侧边栏)  │  │         (主对话区域)                       │  │
│  │            │  │                                           │  │
│  │ ┌────────┐ │  │  ┌─────────────────────────────────────┐  │  │
│  │ │ + 新建  │ │  │  │  Welcome Screen                     │  │  │
│  │ └────────┘ │  │  │  (欢迎界面/建议卡片)                 │  │  │
│  │ ┌────────┐ │  │  └─────────────────────────────────────┘  │  │
│  │ │ 搜索   │ │  │                                           │  │
│  │ └────────┘ │  │  ┌─────────────────────────────────────┐  │  │
│  │ ┌────────┐ │  │  │  Messages Container                 │  │  │
│  │ │ 今天   │ │  │  │  ┌───────────────────────────────┐  │  │  │
│  │ │ ○ 对话 1│ │  │  │  │ Message (AI)                  │  │  │  │
│  │ │ ○ 对话 2│ │  │  │  ├───────────────────────────────┤  │  │  │
│  │ └────────┘ │  │  │  │ Message (User)                │  │  │  │
│  │ ┌────────┐ │  │  │  ├───────────────────────────────┤  │  │  │
│  │ │ 昨天   │ │  │  │  │ Message (AI)                  │  │  │  │
│  │ │ ○ 对话 3│ │  │  │  └───────────────────────────────┘  │  │  │
│  │ └────────┘ │  │  └─────────────────────────────────────┘  │  │
│  │ ┌────────┐ │  │                                           │  │
│  │ │ 更早   │ │  │  ┌─────────────────────────────────────┐  │  │
│  │ │ ○ 对话 4│ │  │  │         ChatInput                 │  │  │
│  │ └────────┘ │  │  │  (输入框/附件/语音/发送)             │  │  │
│  │            │  │  └─────────────────────────────────────┘  │  │
│  └────────────┘  └───────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🧩 组件层次结构

```
ChatWebUI (主页面)
│
├── Header (导航栏)
│   ├── Logo
│   ├── ModelSelector (模型选择)
│   ├── KnowledgeSelector (知识库选择)
│   ├── ExportButton (导出按钮)
│   ├── ThemeToggle (主题切换)
│   ├── SettingsButton (设置)
│   └── UserMenu (用户菜单)
│
├── Sidebar (侧边栏)
│   ├── NewChatButton (新建对话)
│   ├── SearchInput (搜索框)
│   ├── ChatGroups (对话分组)
│   │   ├── Today (今天)
│   │   ├── Yesterday (昨天)
│   │   └── Earlier (更早)
│   └── CollapseButton (折叠按钮)
│
└── ChatArea (对话区域)
    │
    ├── WelcomeScreen (欢迎界面) [无对话时显示]
    │   ├── WelcomeIcon
    │   ├── WelcomeTitle
    │   ├── WelcomeSubtitle
    │   └── SuggestionCards (建议卡片)
    │
    └── MessagesView (消息视图) [有对话时显示]
        ├── MessagesContainer (消息容器)
        │   └── Message[] (消息列表)
        │       ├── Avatar (头像)
        │       ├── Content (内容)
        │       │   └── ReactMarkdown (Markdown 渲染)
        │       │       ├── Code (代码高亮)
        │       │       ├── Math (LaTeX 公式)
        │       │       └── Table (表格)
        │       ├── Timestamp (时间戳)
        │       └── Actions (操作栏)
        │           ├── Edit (编辑)
        │           ├── Regenerate (重新生成)
        │           ├── Copy (复制)
        │           └── Rate (评分)
        │
        └── ChatInput (输入框)
            ├── Textarea (文本输入)
            ├── Preview (Markdown 预览)
            ├── AttachButton (附件)
            ├── VoiceButton (语音)
            ├── MarkdownToggle (Markdown 切换)
            ├── SendButton (发送)
            └── CharCounter (字符计数)
```

---

## 🔄 数据流

```
┌─────────────────┐
│   User Action   │
│  (用户操作)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   ChatInput     │
│  (输入组件)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  handleSend     │
│  (发送处理)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  useChatStore   │
│  (状态管理)      │
│  - sendMessage  │
│  - appendMessage│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    chatApi      │
│  (API 调用)      │
│  - streamMessage│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Backend API   │
│  (后端接口)      │
│  /chat/:id/stream│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Stream Response│
│  (流式响应)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  updateMessage  │
│  (更新消息)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Message      │
│  (消息组件)      │
│  (重新渲染)      │
└─────────────────┘
```

---

## 📦 状态管理

### Zustand Store 结构

```typescript
interface ChatState {
  // 数据
  sessions: ChatSession[];        // 会话列表
  currentSession: ChatSession | null;  // 当前会话
  messages: Message[];            // 消息列表
  
  // 状态
  isLoading: boolean;             // 加载状态
  isStreaming: boolean;           // 流式响应中
  error: string | null;           // 错误信息
  
  // 操作
  fetchSessions: () => Promise<void>;
  createSession: (roleId, title, mode) => Promise<ChatSession>;
  fetchSession: (id) => Promise<void>;
  sendMessage: (content) => Promise<void>;
  sendStreamMessage: (content) => Promise<void>;
  // ...
}
```

### 组件本地状态

```typescript
// ChatWebUI 组件
const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
const [isDarkTheme, setIsDarkTheme] = useState(false);
const [selectedModel, setSelectedModel] = useState('qwen-plus');
const [selectedKnowledge, setSelectedKnowledge] = useState('none');

// ChatInput 组件
const [value, setValue] = useState('');
const [showPreview, setShowPreview] = useState(false);
const [height, setHeight] = useState(56);

// Message 组件
const [isEditing, setIsEditing] = useState(false);
const [editContent, setEditContent] = useState(content);

// Sidebar 组件
const [searchQuery, setSearchQuery] = useState('');
const [editingId, setEditingId] = useState<string | null>(null);
const [editTitle, setEditTitle] = useState('');
```

---

## 🎨 样式系统

### CSS 变量（主题）

```css
:root {
  /* 颜色系统 */
  --bg-primary: #ffffff;
  --bg-secondary: #f7f7f8;
  --bg-tertiary: #ececf1;
  --text-primary: #1a1a1a;
  --text-secondary: #6b6b6b;
  --accent-color: #6366f1;
  
  /* 尺寸系统 */
  --sidebar-width: 280px;
  --header-height: 60px;
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 16px;
  
  /* 阴影系统 */
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.05);
  --shadow-md: 0 4px 6px rgba(0,0,0,0.07);
  --shadow-lg: 0 10px 15px rgba(0,0,0,0.1);
}

[data-theme='dark'] {
  --bg-primary: #1a1a1a;
  --bg-secondary: #2d2d2d;
  --text-primary: #ffffff;
  /* ... */
}
```

### 响应式断点

```css
/* 移动端 */
@media (max-width: 768px) {
  .webui-sidebar {
    position: fixed;
    transform: translateX(-100%);
  }
  
  .message-content {
    max-width: 85%;
  }
}

/* 桌面端 */
@media (min-width: 769px) {
  .webui-sidebar {
    position: relative;
    transform: translateX(0);
  }
}
```

---

## 🔌 API 集成

### 主要 API 端点

```typescript
// 会话管理
GET    /api/v1/chat-sessions          // 获取会话列表
POST   /api/v1/chat-sessions          // 创建会话
GET    /api/v1/chat-sessions/:id      // 获取会话详情
DELETE /api/v1/chat-sessions/:id      // 删除会话
PUT    /api/v1/chat-sessions/:id/title // 更新标题
POST   /api/v1/chat-sessions/:id/archive // 归档

// 消息
POST   /api/v1/chat/:id/complete      // 发送消息（普通）
POST   /api/v1/chat/:id/stream        // 发送消息（流式）
PUT    /api/v1/chat/:id/messages/:mid // 编辑消息
POST   /api/v1/chat/:id/messages/:mid/regenerate // 重新生成
POST   /api/v1/chat/messages/:id/rate // 评分

// 导出
POST   /api/v1/chat-sessions/:id/export // 导出对话
```

### API 调用流程

```
Component Action
    ↓
Event Handler (handleSendMessage)
    ↓
Store Action (sendStreamMessage)
    ↓
API Call (chatApi.streamMessage)
    ↓
HTTP Request (fetch)
    ↓
Backend API
    ↓
Stream Response (SSE)
    ↓
Chunk Handler (onChunk)
    ↓
Store Update (updateLastMessage)
    ↓
Component Re-render (Message)
```

---

## 🚀 性能优化

### 1. 渲染优化

```typescript
// 使用 useMemo 避免重复计算
const groupedSessions = useMemo(() => {
  // 分组逻辑
}, [sessions]);

const filteredGroups = useMemo(() => {
  // 过滤逻辑
}, [groupedSessions, searchQuery]);
```

### 2. 滚动优化

```typescript
// 自动滚动到底部
const messagesEndRef = useRef<HTMLDivElement>(null);

useEffect(() => {
  messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
}, [messages]);
```

### 3. 输入优化

```typescript
// 防抖自动调整高度
useEffect(() => {
  if (textareaRef.current) {
    textareaRef.current.style.height = 'auto';
    const scrollHeight = textareaRef.current.scrollHeight;
    const newHeight = Math.min(Math.max(scrollHeight, MIN_HEIGHT), MAX_HEIGHT);
    textareaRef.current.style.height = `${newHeight}px`;
  }
}, [value]);
```

---

## 🛡️ 错误处理

### 错误边界

```typescript
// Toast 错误提示
{error && (
  <div className="error-toast">
    <span>⚠️</span>
    <span>{error}</span>
    <button onClick={clearError}>×</button>
  </div>
)}
```

### API 错误处理

```typescript
try {
  await chatApi.sendMessage(content);
} catch (error: any) {
  console.error('Failed to send message:', error);
  alert('发送失败：' + error.message);
}
```

---

## 📊 文件依赖关系

```
ChatWebUI.tsx
├── Sidebar.tsx
├── Message.tsx
│   ├── react-markdown
│   ├── remark-gfm
│   ├── remark-math
│   ├── rehype-katex
│   └── rehype-highlight
├── ChatInput.tsx
│   ├── react-markdown
│   ├── remark-gfm
│   ├── remark-math
│   └── rehype-katex
├── webui.css
│   └── katex.min.css
├── chatApi.ts
└── useChatStore.ts
```

---

## 🎯 设计模式

### 1. 组件组合模式

```typescript
// 小组件组合成大组件
<ChatWebUI>
  <Header />
  <Sidebar />
  <ChatArea>
    <WelcomeScreen /> 或 <MessagesView>
  </ChatArea>
</ChatWebUI>
```

### 2. 状态提升模式

```typescript
// 状态保存在父组件，通过 props 传递
<Sidebar
  sessions={sessions}
  onSelectSession={handleSelectSession}
  onDeleteSession={handleDeleteSession}
/>
```

### 3. 自定义 Hook 模式

```typescript
// 使用 Zustand 管理全局状态
const {
  sessions,
  currentSession,
  messages,
  sendStreamMessage,
} = useChatStore();
```

---

## 📈 可扩展性

### 添加新功能

1. **新组件** → 在 `components/WebUI/` 创建
2. **新样式** → 在 `webui.css` 添加
3. **新 API** → 在 `api/chat.ts` 扩展
4. **新状态** → 在 `chatStore.ts` 添加

### 主题定制

```css
/* 在 webui.css 中覆盖变量 */
:root {
  --accent-color: #your-color;
}
```

### 插件系统（未来）

```typescript
// 消息插件
interface MessagePlugin {
  render: (message: Message) => React.ReactNode;
  actions: MessageAction[];
}

// 注册插件
registerPlugin(myPlugin);
```

---

## 🔮 未来架构演进

### v1.0 (当前)
- ✅ 基础对话功能
- ✅ Markdown/LaTeX支持
- ✅ 主题系统

### v2.0 (计划)
- 🔄 虚拟滚动（长对话优化）
- 🔄 离线支持（PWA）
- 🔄 插件系统

### v3.0 (愿景)
- 🔄 实时协作
- 🔄 AI 推荐
- 🔄 多模态支持

---

**架构版本：** v1.0.0  
**最后更新：** 2026-02-27
