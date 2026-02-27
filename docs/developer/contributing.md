# RoleCraft AI - 贡献指南

> 欢迎参与项目开发！

---

## 目录

1. [行为准则](#1-行为准则)
2. [开发流程](#2-开发流程)
3. [提交代码](#3-提交代码)
4. [代码审查](#4-代码审查)
5. [报告问题](#5-报告问题)
6. [请求功能](#6-请求功能)

---

## 1. 行为准则

### 1.1 我们的承诺

为了营造开放和友好的环境，我们承诺：

- 尊重不同观点和经验
- 接受建设性批评
- 关注对社区最有利的事情
- 对其他社区成员表示同理心

### 1.2 不可接受的行为

- 使用性化的语言或图像
- 人身攻击或侮辱性评论
- 公开或私下骚扰
- 未经许可发布他人信息
- 其他不道德或不专业的行为

---

## 2. 开发流程

### 2.1 Fork 项目

1. 在 GitHub 上 Fork 项目
2. Clone 到本地：
```bash
git clone https://github.com/YOUR_USERNAME/rolecraft-ai.git
cd rolecraft-ai
```

### 2.2 创建分支

```bash
# 保持 main 分支同步
git remote add upstream https://github.com/mjscjj/rolecraft-ai.git
git fetch upstream
git checkout main
git merge upstream/main

# 创建功能分支
git checkout -b feat/your-feature-name
```

**分支命名规范：**
- `feat/` - 新功能
- `fix/` - Bug 修复
- `docs/` - 文档更新
- `style/` - 代码格式
- `refactor/` - 重构
- `test/` - 测试
- `chore/` - 构建/工具

### 2.3 本地开发

**配置环境：**
```bash
# 后端
cd backend
go mod download
go run cmd/server/main.go

# 前端
cd frontend
npm install
npm run dev
```

**运行测试：**
```bash
# 后端测试
cd backend
go test ./...

# 前端测试
cd frontend
npm test
npx playwright test
```

### 2.4 提交更改

```bash
# 添加更改
git add .

# 提交（遵循提交规范）
git commit -m "feat: 添加新功能"

# 推送到远程
git push origin feat/your-feature-name
```

---

## 3. 提交代码

### 3.1 创建 Pull Request

1. 在 GitHub 上创建 PR
2. 填写 PR 描述
3. 关联相关 Issue
4. 等待代码审查

### 3.2 PR 描述模板

```markdown
## 变更说明
简要描述此 PR 的目的

## 相关 Issue
Fixes #123

## 测试计划
- [ ] 单元测试通过
- [ ] E2E 测试通过
- [ ] 手动测试完成

## 截图（如适用）
添加前后对比截图

## 检查清单
- [ ] 代码遵循项目规范
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 无新的警告和错误
```

### 3.3 提交信息规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型：**
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

**示例：**
```
feat(role): 添加角色克隆功能

- 实现角色克隆 API
- 添加前端克隆按钮
- 编写单元测试

Closes #45
```

---

## 4. 代码审查

### 4.1 审查流程

1. **自动化检查**
   - CI/CD 流水线
   - 代码风格检查
   - 测试覆盖率

2. **人工审查**
   - 代码质量
   - 功能正确性
   - 性能影响
   - 安全性

3. **反馈修改**
   - 响应审查意见
   - 进行必要修改
   - 重新提交

### 4.2 审查标准

**代码质量：**
- 清晰易读
- 遵循规范
- 无重复代码
- 适当注释

**功能正确：**
- 实现需求
- 边界处理
- 错误处理
- 测试覆盖

**性能考虑：**
- 时间复杂度
- 空间复杂度
- 资源使用
- 扩展性

---

## 5. 报告问题

### 5.1 创建 Issue

在 GitHub Issues 中创建问题，使用相应模板。

### 5.2 Bug 报告模板

```markdown
## 问题描述
清晰简洁地描述问题

## 复现步骤
1. 第一步
2. 第二步
3. 看到错误

## 期望行为
应该发生什么

## 实际行为
实际发生了什么

## 环境信息
- OS: [e.g. macOS 14.0]
- Browser: [e.g. Chrome 120]
- Version: [e.g. v1.0.0]

## 截图
添加截图帮助说明问题

## 日志
```
错误日志内容
```
```

### 5.3 良好实践

- 搜索是否已有相同问题
- 提供详细复现步骤
- 包含环境信息
- 添加截图或日志
- 标注严重程度

---

## 6. 请求功能

### 6.1 功能请求模板

```markdown
## 功能描述
清晰简洁地描述你想要的功能

## 使用场景
这个功能解决什么问题？
目标用户是谁？

## 实现建议
你建议如何实现？
有无参考实现？

## 替代方案
考虑过哪些替代方案？

## 附加信息
其他相关说明
```

### 6.2 功能评估标准

- 用户需求强烈度
- 实现复杂度
- 对现有功能影响
- 维护成本
- 与产品定位一致性

---

## 7. 开发环境

### 7.1 必需工具

- Git
- Go 1.21+
- Node.js 18+
- Docker（可选）

### 7.2 推荐工具

- VS Code / GoLand
- Postman（API 测试）
- Playwright（E2E 测试）

---

## 8. 测试要求

### 8.1 单元测试

**后端：**
```bash
cd backend
go test ./... -cover
```

**要求：**
- 核心业务逻辑覆盖率 > 80%
- 关键函数必须测试
- 边界条件测试

**前端：**
```bash
cd frontend
npm test
```

### 8.2 E2E 测试

```bash
cd frontend
npx playwright test
```

**要求：**
- 核心流程必须覆盖
- 关键功能必须测试
- 回归测试

---

## 9. 文档要求

### 9.1 代码注释

**Go:**
```go
// GetUser 根据 ID 获取用户信息
// 参数:
//   - id: 用户 ID
// 返回:
//   - *User: 用户对象
//   - error: 错误信息
func GetUser(id string) (*User, error) {
    // 实现
}
```

**TypeScript:**
```typescript
/**
 * 创建新角色
 * @param config - 角色配置
 * @returns 创建的角色
 */
function createRole(config: RoleConfig): Role {
    // 实现
}
```

### 9.2 文档更新

- 更新 README.md
- 更新 API 文档
- 添加变更日志
- 更新用户文档

---

## 10. 发布流程

### 10.1 版本命名

遵循 [Semantic Versioning](https://semver.org/)：

```
MAJOR.MINOR.PATCH
```

- MAJOR: 不兼容的 API 变更
- MINOR: 向后兼容的功能
- PATCH: 向后兼容的 Bug 修复

### 10.2 发布检查清单

- [ ] 所有测试通过
- [ ] 代码审查完成
- [ ] 文档已更新
- [ ] 变更日志已写
- [ ] 版本号已更新
- [ ] 构建成功
- [ ] 部署测试通过

---

## 📞 需要帮助？

- 📖 阅读 [开发者文档](./developer/)
- 💬 加入开发者社群
- 📧 联系维护者

---

*最后更新：2026-02-27*
