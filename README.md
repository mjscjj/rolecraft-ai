# RoleCraft AI - AI 角色管理平台

## 项目概述

RoleCraft AI 是一个企业级 AI 角色管理平台，让用户能够创建、配置和管理具备特定技能和知识的 AI 数字员工。

### 核心价值
- **开箱即用** - 内置多种专业角色模板（助理、法务、营销、HR等）
- **灵活定制** - 支持自定义角色提示词、技能和知识库
- **知识增强** - 上传 PDF、Word 等文档，构建专属知识库
- **多场景服务** - 快速问答、深度对话、API 集成

---

## 项目文档

| 文档 | 说明 |
|------|------|
| [PRD.md](./PRD.md) | 产品需求文档 - 功能定义、用户故事、数据库设计 |
| [UI-DESIGN.md](./UI-DESIGN.md) | UI/UX 设计规范 - 视觉系统、页面设计、交互规范 |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | 技术架构设计 - 系统架构、技术选型、项目结构 |

---

## 核心功能

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
- 智能检索与引用

### 4. 对话服务 (Chat Service)
- **快速问答** - 即时响应
- **任务模式** - 多轮深度对话
- 流式输出、历史记录

### 5. API 平台
- API 密钥管理
- RESTful API 接口
- 用量统计

---

## 技术栈

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

---

## 开发计划

| 阶段 | 周期 | 目标 |
|------|------|------|
| MVP | 4周 | 用户认证、基础角色、简单对话 |
| 核心功能 | 4周 | 角色市场、知识库、多文档管理 |
| 进阶功能 | 4周 | 技能系统、API平台、团队协作 |
| 优化迭代 | 持续 | 性能优化、更多模型、企业功能 |

---

## 快速开始

### 环境要求
- Node.js 18+
- Go 1.21+
- PostgreSQL 15+
- Redis 7+

### 本地启动
```bash
# 克隆项目
git clone <repository-url>
cd rolecraft-ai

# 启动基础设施
docker-compose up -d postgres redis milvus minio

# 启动后端
cd backend
go mod download
go run cmd/server/main.go

# 启动前端
cd frontend
pnpm install
pnpm dev
```

---

## 项目结构

```
rolecraft-ai/
├── frontend/          # React + TypeScript 前端
├── backend/           # Go + Gin 后端
│   ├── internal/
│   │   ├── api/       # API 处理器
│   │   ├── models/    # 数据模型
│   │   ├── service/   # 业务逻辑
│   │   └── repository/# 数据访问
│   └── cmd/server/    # 入口文件
├── PRD.md             # 产品需求文档
├── UI-DESIGN.md       # UI设计规范
└── ARCHITECTURE.md    # 技术架构文档
```

---

## 内置角色模板

| 角色 | 分类 | 描述 |
|------|------|------|
| 智能助理 | 通用 | 全能型办公助手 |
| 营销专家 | 营销 | 营销策划与内容创作 |
| 法务顾问 | 法律 | 合同审查与法律咨询 |
| 财务助手 | 财务 | 财务报表与税务咨询 |
| 技术支持 | 技术 | IT 问题诊断与解决 |
| 前台接待 | 行政 | 客户接待与预约管理 |
| HR 专员 | 人事 | 招聘与员工关系 |
| 产品经理 | 产品 | 需求分析与产品设计 |

---

*Made with ❤️ by RoleCraft AI Team*
