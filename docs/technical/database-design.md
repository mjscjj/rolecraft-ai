# RoleCraft AI - 数据库设计文档

> 数据模型与表结构详解

---

## 目录

1. [数据库概览](#1-数据库概览)
2. [核心实体](#2-核心实体)
3. [表结构详解](#3-表结构详解)
4. [索引设计](#4-索引设计)
5. [数据字典](#5-数据字典)

---

## 1. 数据库概览

### 1.1 数据库选型

**开发环境：** SQLite 3  
**生产环境：** PostgreSQL 15+

### 1.2 ER 图

```
┌─────────────┐       ┌─────────────┐
│    User     │       │  Workspace  │
├─────────────┤       ├─────────────┤
│ id          │◄──────│ owner_id    │
│ email       │       │ id          │
│ name        │       │ name        │
│ avatar      │       │ type        │
└──────┬──────┘       └──────┬──────┘
       │                     │
       │ 1:N                 │ 1:N
       ▼                     ▼
┌─────────────┐       ┌─────────────┐
│ Workspace   │       │    Role     │
│   Member    │       ├─────────────┤
├─────────────┤       │ id          │
│ workspace_id│       │ name        │
│ user_id     │       │ system_prompt│
│ role        │       │ model_config│
└─────────────┘       └──────┬──────┘
                             │
                             │ 1:N
                             ▼
                      ┌─────────────┐
                      │ChatSession  │
                      ├─────────────┤
                      │ id          │
                      │ role_id     │
                      │ user_id     │
                      │ messages    │
                      └─────────────┘
```

---

## 2. 核心实体

### 2.1 用户（User）

系统使用者，可以创建和管理多个工作空间。

### 2.2 工作空间（Workspace）

资源隔离单元，包含角色、文档、对话等资源。

### 2.3 角色（Role）

AI 数字员工，包含提示词、技能、知识库配置。

### 2.4 文档（Document）

知识库文档，支持多种格式，向量化存储。

### 2.5 对话会话（ChatSession）

用户与角色的对话记录，支持多种模式。

---

## 3. 表结构详解

### 3.1 users 表

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    avatar VARCHAR(500),
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**字段说明：**
- `id`: 用户唯一标识
- `email`: 登录邮箱，全局唯一
- `password_hash`: bcrypt 加密密码
- `name`: 用户昵称
- `avatar`: 头像 URL

### 3.2 workspaces 表

```sql
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) DEFAULT 'personal',
    description TEXT,
    logo VARCHAR(500),
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**字段说明：**
- `type`: personal（个人）, team（团队）, enterprise（企业）
- `settings`: JSON 配置，包含配额等

### 3.3 workspace_members 表

```sql
CREATE TABLE workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workspace_id, user_id)
);
```

**角色类型：**
- `owner`: 所有者
- `admin`: 管理员
- `member`: 普通成员

### 3.4 roles 表

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    avatar VARCHAR(500),
    category VARCHAR(50),
    system_prompt TEXT NOT NULL,
    welcome_message TEXT,
    model_config JSONB DEFAULT '{}',
    skills JSONB DEFAULT '[]',
    is_template BOOLEAN DEFAULT FALSE,
    is_public BOOLEAN DEFAULT FALSE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**model_config 结构：**
```json
{
  "model": "gpt-4",
  "temperature": 0.7,
  "max_tokens": 1000,
  "top_p": 0.9
}
```

### 3.5 documents 表

```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    file_type VARCHAR(20) NOT NULL,
    file_size BIGINT,
    file_path VARCHAR(500),
    status VARCHAR(20) DEFAULT 'pending',
    chunk_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**status 枚举：**
- `pending`: 待处理
- `processing`: 处理中
- `completed`: 已完成
- `failed`: 失败

### 3.6 chat_sessions 表

```sql
CREATE TABLE chat_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    mode VARCHAR(20) DEFAULT 'quick',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**mode 枚举：**
- `quick`: 快速问答
- `task`: 任务模式

### 3.7 messages 表

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    sources JSONB DEFAULT '[]',
    tokens INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**role 枚举：**
- `user`: 用户消息
- `assistant`: AI 回复
- `system`: 系统消息

### 3.8 api_keys 表

```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    permissions JSONB DEFAULT '{}',
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 4. 索引设计

### 4.1 用户相关

```sql
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);
```

### 4.2 角色相关

```sql
CREATE INDEX idx_roles_workspace_id ON roles(workspace_id);
CREATE INDEX idx_roles_category ON roles(category);
CREATE INDEX idx_roles_is_template ON roles(is_template);
CREATE INDEX idx_roles_created_at ON roles(created_at);
```

### 4.3 对话相关

```sql
CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE INDEX idx_chat_sessions_updated_at ON chat_sessions(updated_at);
```

### 4.4 文档相关

```sql
CREATE INDEX idx_documents_workspace_id ON documents(workspace_id);
CREATE INDEX idx_documents_status ON documents(status);
```

---

## 5. 数据字典

### 5.1 枚举类型

**workspace.type:**
| 值 | 说明 |
|---|------|
| personal | 个人空间 |
| team | 团队空间 |
| enterprise | 企业空间 |

**roles.category:**
| 值 | 说明 |
|---|------|
| general | 通用办公 |
| marketing | 营销销售 |
| legal | 法律咨询 |
| finance | 财务会计 |
| technology | 技术支持 |
| hr | 人力资源 |
| product | 产品设计 |

**documents.status:**
| 值 | 说明 |
|---|------|
| pending | 待处理 |
| processing | 处理中 |
| completed | 已完成 |
| failed | 失败 |

**chat_sessions.mode:**
| 值 | 说明 |
|---|------|
| quick | 快速问答 |
| task | 任务模式 |

---

## 📚 相关文档

- [API 参考文档](./api-reference.md)
- [系统架构图](./architecture.md)

---

*最后更新：2026-02-27*
