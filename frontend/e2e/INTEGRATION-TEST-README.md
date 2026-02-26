# E2E Integration Tests - RoleCraft AI

端到端集成测试，验证完整用户流程。

## 📋 测试场景

测试覆盖以下完整用户流程：

```
注册 → 登录 → 创建角色 → 上传文档 → 对话 → 验证
```

### 详细测试步骤

#### 1. 用户注册与登录
- ✅ 用户注册成功
- ✅ 使用注册的账号登录
- ✅ 获取当前用户信息

#### 2. 角色创建与管理
- ✅ 创建测试角色
- ✅ 获取角色列表并验证新角色存在
- ✅ 获取角色详情

#### 3. 文档上传与处理
- ✅ 上传测试文档 (TXT)
- ✅ 等待文档处理完成（轮询检查状态）
- ✅ 获取文档列表验证文档存在
- ✅ 文档向量搜索

#### 4. 对话功能测试
- ✅ 创建会话
- ✅ 发送问候消息并获取回复
- ✅ 发送写作请求并获取回复
- ✅ 发送分析问题并获取回复
- ✅ 获取会话历史

#### 5. 完整流程验证
- ✅ 验证完整用户流程数据一致性

#### 6. 错误处理测试
- ✅ 使用错误密码登录失败
- ✅ 未授权访问受保护接口失败
- ✅ 访问不存在的角色失败

## 🚀 快速开始

### 前置条件

- Node.js 20+
- Go 1.21+
- pnpm
- PostgreSQL (测试用)
- Playwright 浏览器

### 本地运行测试

#### 方法 1: 使用测试脚本（推荐）

```bash
cd frontend/e2e
./run-integration-tests.sh
```

脚本会自动：
1. 检查并安装依赖
2. 安装 Playwright 浏览器
3. 构建前后端
4. 启动后端和前端服务
5. 运行 E2E 测试
6. 生成 HTML 报告

#### 方法 2: 手动运行

```bash
# 1. 启动后端
cd backend
export DATABASE_URL="postgres://test:test@localhost:5432/rolecraft_e2e?sslmode=disable"
export JWT_SECRET="test-secret"
export UPLOAD_DIR="/tmp/uploads"
go run cmd/server/main.go

# 2. 启动前端（新终端）
cd frontend
pnpm dev

# 3. 运行测试（新终端）
cd frontend
pnpm exec playwright test e2e/integration.spec.ts
```

### 运行特定测试

```bash
# 运行特定测试文件
pnpm exec playwright test e2e/integration.spec.ts

# 运行特定测试用例（按名称过滤）
pnpm exec playwright test e2e/integration.spec.ts --grep "用户注册"

# 运行特定测试描述块
pnpm exec playwright test e2e/integration.spec.ts --grep "角色创建"

# 带 UI 模式运行
pnpm exec playwright test e2e/integration.spec.ts --ui

# 生成 HTML 报告
pnpm exec playwright test e2e/integration.spec.ts --reporter=html
pnpm exec playwright show-report
```

## 📊 CI/CD 集成

### GitHub Actions

E2E 测试已集成到 GitHub Actions CI 流程中：

- **触发条件**: push 到 main/develop 分支，或 PR 到 main 分支
- **工作流程**: `.github/workflows/e2e-integration.yml`
- **集成点**: 主 CI 流程 (`.github/workflows/ci.yml`)

#### CI 流程

```yaml
backend-test → frontend-test → e2e-integration-test → deploy
```

#### 查看测试结果

1. 在 GitHub Actions 页面找到对应的 workflow run
2. 点击 "E2E Integration Tests" job
3. 下载 `playwright-report` artifact
4. 本地解压后使用 `npx playwright show-report` 查看

## 📁 文件结构

```
frontend/e2e/
├── integration.spec.ts          # E2E 集成测试主文件
├── run-integration-tests.sh     # 本地测试运行脚本
├── login.spec.ts                # 登录测试
├── roles.spec.ts                # 角色管理测试
├── chat.spec.ts                 # 对话功能测试
├── screenshot.spec.ts           # 截图测试
├── ChatStream.spec.ts          # 聊天流测试
└── README.md                    # E2E 测试说明
```

## 🔧 配置说明

### 环境变量

测试使用以下环境变量：

```bash
# 数据库
DATABASE_URL=postgres://test:test@localhost:5432/rolecraft_e2e?sslmode=disable

# JWT 认证
JWT_SECRET=e2e-test-jwt-secret-key-for-testing-only

# 文件上传
UPLOAD_DIR=/tmp/uploads

# AnythingLLM (可选)
ANYTHINGLLM_BASE_URL=http://localhost:3001/api/v1
ANYTHINGLLM_API_KEY=test-api-key
ANYTHINGLLM_WORKSPACE=e2e_test_workspace
```

### Playwright 配置

配置位于 `frontend/playwright.config.ts`：

```typescript
{
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
}
```

## 📈 测试报告

### HTML 报告

运行测试后生成 HTML 报告：

```bash
pnpm exec playwright show-report
```

报告包含：
- ✅/❌ 测试用例状态
- ⏱️ 执行时间
- 📸 失败截图
- 🔍 详细日志

### 控制台输出

测试运行时会输出详细日志：

```
Running 15 tests using 1 worker

  ✓  1 e2e/integration.spec.ts:30:5 › 端到端集成测试 › 1. 用户注册与登录 › 用户注册成功 (1.2s)
  ✓  2 e2e/integration.spec.ts:45:5 › 端到端集成测试 › 1. 用户注册与登录 › 使用注册的账号登录 (856ms)
  ✓  3 e2e/integration.spec.ts:59:5 › 端到端集成测试 › 1. 用户注册与登录 › 获取当前用户信息 (423ms)
  ...
```

## 🐛 故障排查

### 常见问题

#### 1. 后端启动失败

```bash
# 检查数据库连接
psql postgres://test:test@localhost:5432/rolecraft_e2e

# 查看后端日志
cat /tmp/backend.log
```

#### 2. 前端启动失败

```bash
# 检查端口是否被占用
lsof -i :5173

# 查看前端日志
cat /tmp/frontend.log
```

#### 3. 测试超时

增加超时时间：

```bash
pnpm exec playwright test e2e/integration.spec.ts --timeout=60000
```

#### 4. 文档处理失败

检查 AnythingLLM 连接：

```bash
curl http://localhost:3001/api/v1/health
```

### 调试模式

```bash
# 有头模式（显示浏览器）
pnpm exec playwright test e2e/integration.spec.ts --headed

# 调试模式（逐步执行）
pnpm exec playwright test e2e/integration.spec.ts --debug

# 输出详细日志
DEBUG=pw:api pnpm exec playwright test e2e/integration.spec.ts
```

## 📝 最佳实践

### 测试数据隔离

- 每个测试使用唯一的邮箱/数据
- 使用 `Date.now()` 生成唯一标识
- 测试完成后清理数据（如需要）

### 异步操作处理

```typescript
// 轮询检查异步操作完成
for (let attempt = 0; attempt < maxAttempts; attempt++) {
  const status = await checkStatus();
  if (status === 'completed') break;
  await page.waitForTimeout(pollInterval);
}
```

### 错误处理

```typescript
// 优雅的错误处理
test('测试用例', async ({ page }) => {
  try {
    // 测试逻辑
  } catch (error) {
    test.fail();
    throw error;
  }
});
```

## 🎯 扩展测试

### 添加新测试场景

1. 在 `integration.spec.ts` 中添加新的 `test.describe` 块
2. 使用现有的 `authToken`, `roleId` 等上下文变量
3. 遵循 AAA 模式 (Arrange-Act-Assert)

### 性能测试

```typescript
test('API 响应时间测试', async ({ page }) => {
  const startTime = Date.now();
  await page.request.post(`${API_BASE}/chat/...`);
  const responseTime = Date.now() - startTime;
  expect(responseTime).toBeLessThan(5000); // 5 秒内
});
```

## 📚 相关文档

- [E2E 测试总体说明](./README.md)
- [Playwright 官方文档](https://playwright.dev)
- [项目架构文档](../../ARCHITECTURE.md)
- [API 文档](../../backend/docs/API.md)

## 🤝 贡献指南

1. 确保所有现有测试通过
2. 为新功能添加相应的 E2E 测试
3. 更新本文档
4. 提交 PR 时包含测试结果

---

**维护者**: RoleCraft AI Team  
**最后更新**: 2026-02-26
