# RoleCraft AI - Frontend

## 技术栈

- React 18 + TypeScript
- Vite
- Tailwind CSS
- Zustand (状态管理)
- Axios (HTTP 客户端)

## 项目结构

```
frontend/
├── src/
│   ├── api/           # API 客户端
│   │   ├── client.ts  # Axios 实例
│   │   ├── auth.ts    # 认证 API
│   │   ├── user.ts    # 用户 API
│   │   ├── role.ts    # 角色 API
│   │   ├── document.ts # 文档 API
│   │   └── chat.ts    # 对话 API
│   ├── stores/        # 状态管理
│   │   ├── authStore.ts
│   │   ├── roleStore.ts
│   │   └── chatStore.ts
│   ├── hooks/         # 自定义 Hooks
│   ├── pages/         # 页面组件
│   ├── components/    # UI 组件
│   └── main.tsx       # 入口
├── package.json
└── vite.config.ts
```

## 开发

```bash
# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 构建
pnpm build

# 预览
pnpm preview
```

## 环境变量

创建 `.env` 文件：

```
VITE_API_URL=http://localhost:8080
```
