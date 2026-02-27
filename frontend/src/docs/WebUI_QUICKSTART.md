# RoleCraft WebUI 快速开始

## 🚀 5 分钟快速上手

### 第一步：安装依赖（已完成）

```bash
cd rolecraft-ai/frontend
npm install
```

已安装的新依赖：
- `highlight.js` - 代码高亮
- `katex` - LaTeX 公式渲染
- `rehype-highlight` - Markdown 代码高亮插件
- `rehype-katex` - Markdown LaTeX 插件
- `remark-math` - Markdown 数学语法支持

### 第二步：在 App.tsx 中添加路由

```tsx
// src/App.tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import ChatWebUI from './pages/ChatWebUI';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* 现有路由 */}
        <Route path="/" element={<Dashboard />} />
        
        {/* 新增 WebUI 路由 */}
        <Route path="/chat" element={<ChatWebUI />} />
        <Route path="/chat/:roleId" element={<ChatWebUI />} />
        
        {/* 其他路由 */}
      </Routes>
    </BrowserRouter>
  );
}
```

### 第三步：引入样式（可选，如果未在 App 中全局引入）

```tsx
// src/main.tsx 或 src/App.tsx
import './styles/webui.css';
```

### 第四步：启动开发服务器

```bash
npm run dev
```

访问 `http://localhost:5173/chat` 即可看到新的对话界面！

---

## 💡 使用示例

### 基础用法

```tsx
import ChatWebUI from './pages/ChatWebUI';

// 最简单用法
<ChatWebUI />
```

### 指定角色 ID

```tsx
// 启动时自动创建该角色的对话
<ChatWebUI initialRoleId="role-123" />
```

### 从角色列表跳转

```tsx
// 在角色卡片组件中
const handleStartChat = (roleId: string) => {
  navigate(`/chat/${roleId}`);
};
```

---

## 🎨 自定义配置

### 修改主题色

编辑 `src/styles/webui.css`：

```css
:root {
  /* 主色调 - 改成你的品牌色 */
  --accent-color: #6366f1;      /* 主色 */
  --accent-hover: #4f46e5;      /* 悬停色 */
  --accent-light: #e0e7ff;      /* 浅色 */
  
  /* 背景色 */
  --bg-primary: #ffffff;
  --bg-secondary: #f7f7f8;
  
  /* 文字颜色 */
  --text-primary: #1a1a1a;
  --text-secondary: #6b6b6b;
}
```

### 修改默认模型

编辑 `src/pages/ChatWebUI.tsx`：

```tsx
const [selectedModel, setSelectedModel] = useState('qwen-max'); // 改成默认模型
```

### 自定义欢迎页面

编辑 `src/pages/ChatWebUI.tsx` 中的 `suggestions` 数组：

```tsx
const suggestions = [
  { icon: '🎯', text: '你的第一个建议' },
  { icon: '✨', text: '你的第二个建议' },
  // ...
];
```

---

## 🔌 API 集成

### 后端 API 配置

确保 `.env` 文件中配置了正确的 API 地址：

```env
VITE_API_URL=http://localhost:8080/api/v1
```

### 认证配置

系统会自动使用 localStorage 中的 token：

```javascript
// 登录后保存 token
localStorage.setItem('token', 'your-jwt-token');
localStorage.setItem('user_name', 'User Name');
```

---

## 📱 移动端适配

WebUI 已自动适配移动端，无需额外配置。

### 测试移动端

1. 打开浏览器开发者工具
2. 切换到设备模拟模式
3. 选择任意移动设备
4. 访问 `/chat` 页面

---

## 🐛 常见问题

### Q1: 样式不生效？

**解决：** 确保已引入 `webui.css` 文件

```tsx
import './styles/webui.css';
```

### Q2: Markdown 不渲染？

**解决：** 检查依赖是否安装完整

```bash
npm install react-markdown remark-gfm remark-math rehype-katex rehype-highlight
```

### Q3: 代码高亮不工作？

**解决：** 确保安装了 highlight.js 和 rehype-highlight

```bash
npm install highlight.js rehype-highlight
```

### Q4: LaTeX 公式显示异常？

**解决：** 确保安装了 katex 和 rehype-katex，并引入了样式

```bash
npm install katex rehype-katex
```

样式已自动在 `webui.css` 中引入：
```css
@import 'katex/dist/katex.min.css';
```

### Q5: 无法连接后端？

**解决：** 检查 `.env` 配置和后端服务是否启动

```bash
# 检查后端服务
curl http://localhost:8080/api/v1/health
```

---

## 📚 更多文档

- **完整使用指南** - `WebUI_GUIDE.md`
- **实现报告** - `WebUI_IMPLEMENTATION.md`
- **API 文档** - `../api/chat.ts`

---

## 🎉 开始使用

现在你已经准备好了！访问 `/chat` 开始体验全新的 RoleCraft WebUI！

**Happy Chatting!** 🎭
