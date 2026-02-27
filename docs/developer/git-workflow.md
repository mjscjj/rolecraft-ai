# RoleCraft AI - Git 工作流

> 版本控制和协作流程

---

## 1. 分支模型

### 1.1 分支类型

```
main (生产)
  │
  ├── develop (开发)
  │     │
  │     ├── feature/user-auth
  │     ├── feature/role-management
  │     └── fix/login-bug
  │
  └── release/v1.0.0 (发布)
```

**分支说明：**

| 分支 | 用途 | 命名 |
|------|------|------|
| `main` | 生产环境，稳定版本 | - |
| `develop` | 开发主分支 | - |
| `feature/*` | 新功能开发 | `feat/功能名` |
| `fix/*` | Bug 修复 | `fix/问题描述` |
| `release/*` | 版本发布 | `release/v1.0.0` |
| `hotfix/*` | 紧急修复 | `hotfix/问题描述` |

### 1.2 分支策略

**长期分支：**
- `main` - 永远稳定，可直接部署
- `develop` - 集成分支，包含最新开发代码

**短期分支：**
- 从 `develop` 创建
- 完成后合并回 `develop`
- 删除已合并分支

---

## 2. 开发流程

### 2.1 开始新功能

```bash
# 1. 同步 develop 分支
git checkout develop
git pull upstream develop

# 2. 创建功能分支
git checkout -b feat/user-authentication

# 3. 开发功能
# ... 编写代码 ...

# 4. 提交更改
git add .
git commit -m "feat: 实现用户认证功能"

# 5. 推送到远程
git push origin feat/user-authentication
```

### 2.2 提交规范

遵循 Conventional Commits：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型：**
- `feat` - 新功能
- `fix` - Bug 修复
- `docs` - 文档更新
- `style` - 代码格式（不影响功能）
- `refactor` - 重构
- `test` - 测试
- `chore` - 构建/工具

**示例：**
```bash
feat(auth): 添加 JWT Token 认证

- 实现登录 API
- 添加 Token 刷新机制
- 编写单元测试

Closes #123
```

### 2.3 保持分支更新

```bash
# 定期同步 develop
git fetch upstream
git rebase upstream/develop

# 解决冲突后
git push origin feat/user-authentication --force-with-lease
```

---

## 3. Pull Request 流程

### 3.1 创建 PR

1. **推送到远程**
```bash
git push origin feat/user-authentication
```

2. **在 GitHub 创建 PR**
   - 标题清晰
   - 描述详细
   - 关联 Issue
   - 选择 Reviewer

### 3.2 PR 描述模板

```markdown
## 变更说明
实现用户认证功能，包括登录、Token 刷新

## 相关 Issue
Fixes #123

## 测试计划
- [x] 单元测试通过
- [x] E2E 测试通过
- [x] 手动测试完成

## 截图
[添加截图]

## 检查清单
- [x] 代码遵循规范
- [x] 添加了测试
- [x] 更新了文档
```

### 3.3 代码审查

**审查要点：**
- 代码质量
- 功能正确性
- 测试覆盖
- 性能影响
- 安全性

**审查响应：**
```bash
# 根据审查意见修改
git add .
git commit -m "address review comments"

# 推送到同一分支
git push origin feat/user-authentication
```

### 3.4 合并 PR

**合并方式：**
- **Squash and Merge** - 推荐，压缩为一个提交
- **Rebase and Merge** - 保持提交历史线性
- **Create Merge Commit** - 保留完整历史

**合并后：**
1. 删除功能分支
2. 更新本地仓库
```bash
git checkout develop
git pull upstream develop
git branch -d feat/user-authentication
```

---

## 4. 发布流程

### 4.1 准备发布

```bash
# 1. 从 develop 创建 release 分支
git checkout -b release/v1.0.0 develop

# 2. 更新版本号
# package.json, version.go 等

# 3. 更新 CHANGELOG.md
# 添加版本更新说明

# 4. 提交
git commit -m "chore: bump version to 1.0.0"
```

### 4.2 测试验证

- 功能测试
- 回归测试
- 性能测试
- 文档检查

### 4.3 合并发布

```bash
# 1. 合并到 main
git checkout main
git merge --no-ff release/v1.0.0
git tag -a v1.0.0 -m "Release version 1.0.0"

# 2. 合并回 develop
git checkout develop
git merge --no-ff release/v1.0.0

# 3. 删除 release 分支
git branch -d release/v1.0.0

# 4. 推送
git push origin main develop --tags
```

---

## 5. 紧急修复

### 5.1 Hotfix 流程

```bash
# 1. 从 main 创建 hotfix 分支
git checkout -b hotfix/login-bug main

# 2. 修复 Bug
# ... 修复代码 ...

# 3. 提交
git commit -m "fix: 修复登录 Token 验证问题"

# 4. 合并到 main 和 develop
git checkout main
git merge --no-ff hotfix/login-bug
git tag -a v1.0.1 -m "Hotfix 1.0.1"

git checkout develop
git merge --no-ff hotfix/login-bug

# 5. 删除分支
git branch -d hotfix/login-bug
git push origin main develop --tags
```

---

## 6. 最佳实践

### 6.1 提交频率

- **频繁提交** - 小步快跑
- **原子提交** - 每个提交完成一个功能
- **及时推送** - 避免本地堆积

### 6.2 分支管理

- **分支命名** - 清晰明确
- **及时删除** - 合并后删除
- **定期同步** - 避免大冲突

### 6.3 冲突解决

```bash
# 遇到冲突时
git fetch upstream
git rebase upstream/develop

# 解决冲突文件
# ... 编辑文件 ...

# 继续 rebase
git add .
git rebase --continue
```

### 6.4 回滚操作

```bash
# 撤销最后一次提交（保留更改）
git reset --soft HEAD~1

# 撤销提交和更改
git reset --hard HEAD~1

# 撤销已推送的提交
git revert <commit-hash>
git push origin develop
```

---

## 7. 工具配置

### 7.1 Git 配置

```bash
# 全局配置
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# 别名
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit
git config --global alias.st status
```

### 7.2 Git Hooks

**使用 Husky（前端）：**
```json
{
  "husky": {
    "hooks": {
      "pre-commit": "lint-staged",
      "commit-msg": "commitlint -E HUSKY_GIT_PARAMS"
    }
  }
}
```

**Pre-commit 检查：**
- 代码格式化
- Lint 检查
- 测试运行

---

## 8. 常见问题

### Q1: 如何撤销已推送的提交？

```bash
# 使用 revert（安全）
git revert <commit-hash>
git push origin develop
```

### Q2: 如何处理大冲突？

```bash
# 使用 merge tool
git mergetool

# 或手动解决
# 1. 找到冲突文件
# 2. 编辑解决冲突标记
# 3. git add 解决的文件
# 4. git rebase --continue
```

### Q3: 如何重写提交历史？

```bash
# 交互式 rebase
git rebase -i HEAD~5

# 可以：
# - 修改提交顺序
# - 合并提交
# - 修改提交信息
```

---

## 📚 相关资源

- [Git 官方文档](https://git-scm.com/doc)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)

---

*最后更新：2026-02-27*
