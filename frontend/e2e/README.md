# RoleCraft AI E2E 测试指南

## 📦 安装 Playwright

```bash
cd frontend
npm install -D @playwright/test
npx playwright install chromium
```

## 🚀 运行测试

```bash
# 运行所有测试
npx playwright test

# 运行特定测试
npx playwright test e2e/login.spec.ts

# 有头模式（查看浏览器）
npx playwright test --headed

# 生成测试报告
npx playwright test --reporter=html
```

## 📋 测试用例

### 1. 认证流程 (login.spec.ts)
- ✅ 用户登录
- ✅ 用户注册
- ✅ 错误密码处理
- ✅ Token 存储

### 2. 角色管理 (roles.spec.ts)
- ✅ 获取角色列表
- ✅ 创建新角色
- ✅ 编辑角色
- ✅ 删除角色

### 3. 对话功能 (chat.spec.ts)
- ✅ 创建会话
- ✅ 发送消息
- ✅ 接收 AI 回复
- ✅ 消息历史加载

### 4. 知识库 (documents.spec.ts)
- ✅ 文档列表
- ✅ 上传文档
- ✅ 删除文档

## 🔧 配置

编辑 `playwright.config.ts`:
- `baseURL`: http://localhost:5173
- `API_BASE`: http://localhost:8080/api/v1
- `timeout`: 30000ms

## 📊 测试报告

测试完成后运行：
```bash
npx playwright show-report
```
