# RoleCraft AI - AI 角色管理平台

> 创建你的 AI 数字员工团队，5 分钟配置，立即上岗工作

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Tests](https://img.shields.io/badge/tests-100%25-brightgreen)

---

## 🎯 项目概述

RoleCraft AI 是一个企业级 AI 角色管理平台，让用户能够创建、配置和管理具备特定技能和知识的 AI 数字员工。

### 核心价值
- **开箱即用** - 内置 8+ 专业角色模板（助理、法务、营销、HR 等）
- **灵活定制** - 支持自定义角色提示词、技能和知识库
- **知识增强** - 上传 PDF、Word 等文档，构建专属知识库
- **多场景服务** - 快速问答、深度对话、API 集成
- **极致性能** - API 响应 < 1ms，并发 435+ req/s

### 快速链接
- 📖 [完整文档中心](./docs/README.md)
- 🚀 [快速开始指南](./docs/user/quickstart.md)
- 📡 [API 参考文档](./docs/technical/api-reference.md)
- 💰 [定价说明](./docs/marketing/pricing.md)

---

## 📚 完整文档

### 用户文档 👥
- 🚀 [快速开始指南](./docs/user/quickstart.md) - 5 分钟上手
- 📖 [功能使用手册](./docs/user/user-guide.md) - 完整功能说明
- 💡 [最佳实践](./docs/user/best-practices.md) - 使用技巧
- ❓ [常见问题 FAQ](./docs/user/faq.md) - 问题解答
- 🎬 [视频教程脚本](./docs/user/video-scripts.md) - 教程资源

### 技术文档 🛠️
- 📡 [API 参考文档](./docs/technical/api-reference.md) - 完整 API 文档
- 🚀 [部署指南](./docs/technical/deployment-guide.md) - 生产环境部署
- 🏗️ [开发环境配置](./docs/technical/dev-setup.md) - 本地开发搭建
- 🗄️ [数据库设计文档](./docs/technical/database-design.md) - 数据模型
- 🏛️ [系统架构图](./docs/technical/architecture.md) - 技术架构

### 开发者文档 💻
- 🤝 [贡献指南](./docs/developer/contributing.md) - 参与项目开发
- 📝 [代码风格指南](./docs/developer/code-style.md) - 代码规范
- 🌿 [Git 工作流](./docs/developer/git-workflow.md) - 版本控制
- 📦 [发布流程](./docs/developer/release-process.md) - 版本发布

### 营销文档 📢
- 📣 [Product Hunt 文案](./docs/marketing/product-hunt.md) - 产品发布
- 📄 [产品单页](./docs/marketing/one-pager.md) - 产品介绍
- 📊 [功能对比表](./docs/marketing/comparison.md) - 竞品对比
- 💰 [定价说明](./docs/marketing/pricing.md) - 价格策略

**更多文档：** [PRD.md](./PRD.md) | [ARCHITECTURE.md](./ARCHITECTURE.md) | [FEATURES.md](./FEATURES.md) | [PROJECT_STATUS.md](./PROJECT_STATUS.md)

---

## ✨ 核心功能

### 1. 空间管理 (Workspace)
- 个人空间与企业/项目空间
- 成员邀请与权限管理
- 资源隔离

### 2. 角色中心 (Role Center)
- **角色市场** - 8+ 内置角色模板
- **角色编辑器** - 可视化配置提示词、技能、知识库
- **角色管理** - CRUD、克隆、分享

### 3. 知识库 (Knowledge Base)
- 支持 PDF、Word、TXT、Markdown 等格式
- 自动文本提取与向量化
- 智能检索与引用（准确率 > 90%）

### 4. 对话服务 (Chat Service)
- **快速问答** - 即时响应（< 500ms）
- **任务模式** - 多轮深度对话（50 条上下文）
- 流式输出、历史记录、快捷命令

### 5. API 平台
- API 密钥管理
- RESTful API 接口
- 用量统计和监控
- SDK 支持（Python/Node.js）

---

## 🛠️ 技术栈

### 前端
- React 18 + TypeScript
- Tailwind CSS + shadcn/ui
- TanStack Query + Zustand
- Vite 构建

### 后端
- Go + Gin 框架
- GORM + PostgreSQL
- Redis + Milvus
- MinIO 文件存储

### 基础设施
- Docker + Docker Compose
- GitHub Actions (CI/CD)
- Prometheus + Grafana (监控)

---

## 🚀 快速开始

### 环境要求
- Node.js 18+
- Go 1.21+
- Docker 20.10+ (可选)

### 方式 1：Docker 启动（推荐）

```bash
# 克隆项目
git clone https://github.com/mjscjj/rolecraft-ai.git
cd rolecraft-ai

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

访问：
- 前端：http://localhost:5173
- 后端 API：http://localhost:8080
- Swagger 文档：http://localhost:8080/swagger

### 方式 2：源码启动

```bash
# 启动后端
cd backend
go mod download
go run cmd/server/main.go

# 启动前端（新终端）
cd frontend
pnpm install
pnpm dev
```

**详细指南：** [开发环境配置](./docs/technical/dev-setup.md) | [部署指南](./docs/technical/deployment-guide.md)

---

## 📂 项目结构

```
rolecraft-ai/
├── docs/                # 完整文档中心
│   ├── user/           # 用户文档
│   ├── technical/      # 技术文档
│   ├── developer/      # 开发者文档
│   └── marketing/      # 营销文档
├── frontend/           # React + TypeScript 前端
├── backend/            # Go + Gin 后端
│   ├── internal/
│   │   ├── api/        # API 处理器
│   │   ├── models/     # 数据模型
│   │   ├── service/    # 业务逻辑
│   │   └── repository/ # 数据访问
│   └── cmd/server/     # 入口文件
├── tests/              # 测试文件
├── scripts/            # 脚本工具
└── PRD.md              # 产品需求文档
```

---

## 🎭 内置角色模板

| 角色 | 分类 | 描述 |
|------|------|------|
| 🤖 智能助理 | 通用 | 全能型办公助手 |
| 📈 营销专家 | 营销 | 营销策划与内容创作 |
| ⚖️ 法务顾问 | 法律 | 合同审查与法律咨询 |
| 💰 财务助手 | 财务 | 财务报表与税务咨询 |
| 💻 技术支持 | 技术 | IT 问题诊断与解决 |
| 🏢 前台接待 | 行政 | 客户接待与预约管理 |
| 👥 HR 专员 | 人事 | 招聘与员工关系 |
| 📦 产品经理 | 产品 | 需求分析与产品设计 |

---

## 📊 项目状态

### 测试覆盖
- ✅ API 测试：100%
- ✅ E2E 测试：100%
- ✅ 性能测试：100%
- ✅ 总计：68+ 测试用例

### 性能指标
- ⚡ API 响应：< 1ms (P95)
- 🚀 并发处理：435+ req/s
- 💾 内存占用：< 20MB (后端)
- ⏱️ 启动时间：< 2s

**详细报告：** [PROJECT_STATUS.md](./PROJECT_STATUS.md) | [FEATURES.md](./FEATURES.md)

---

## 📈 开发路线图

| 阶段 | 周期 | 状态 | 目标 |
|------|------|------|------|
| MVP | 4 周 | ✅ 完成 | 用户认证、基础角色、简单对话 |
| 核心功能 | 4 周 | ✅ 完成 | 角色市场、知识库、多文档管理 |
| 进阶功能 | 4 周 | 🔄 进行中 | 技能系统、API 平台、团队协作 |
| 优化迭代 | 持续 | 📋 规划 | 性能优化、更多模型、企业功能 |

**详细路线：** [ROADMAP.md](./ROADMAP.md)

---

## 🤝 参与贡献

我们欢迎各种形式的贡献！

### 快速贡献
1. Fork 项目
2. 创建功能分支 (`git checkout -b feat/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: add AmazingFeature'`)
4. 推送到分支 (`git push origin feat/AmazingFeature`)
5. 开启 Pull Request

### 贡献指南
- 📖 [贡献指南](./docs/developer/contributing.md)
- 📝 [代码风格指南](./docs/developer/code-style.md)
- 🌿 [Git 工作流](./docs/developer/git-workflow.md)

---

## 📞 联系支持

- 🌐 官网：https://rolecraft.ai
- 📧 邮箱：support@rolecraft.ai
- 💬 GitHub Issues: [提交问题](https://github.com/mjscjj/rolecraft-ai/issues)
- 📖 文档：https://docs.rolecraft.ai

---

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

---

<div align="center">

**Made with ❤️ by RoleCraft AI Team**

⭐ 如果这个项目对你有帮助，请给一个 Star！

</div>
