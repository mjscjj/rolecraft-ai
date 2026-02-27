# RoleCraft AI - 代码风格指南

> 统一的代码规范和风格

---

## 目录

1. [Go 代码规范](#1-go 代码规范)
2. [TypeScript 代码规范](#2-typescript 代码规范)
3. [通用规范](#3-通用规范)
4. [命名约定](#4-命名约定)
5. [注释规范](#5-注释规范)

---

## 1. Go 代码规范

### 1.1 格式化

使用 `gofmt` 自动格式化：

```bash
gofmt -w .
```

**VS Code 设置：**
```json
{
  "go.formatTool": "gofmt",
  "go.lintTool": "golint",
  "go.vetOnSave": "package"
}
```

### 1.2 命名规范

**包名：**
- 小写，无下划线
- 简短明确
```go
package user      // ✅
package User      // ❌
package user_mgr  // ❌
```

**变量名：**
- 驼峰式
- 简短有意义
```go
var userName string    // ✅
var user_name string   // ❌
var u string           // ❌ 除非上下文清晰
```

**常量名：**
- 全大写，下划线分隔
```go
const MaxRetryCount = 3
const APIVersion = "v1"
```

**接口名：**
- 单个方法：-er 后缀
- 多个方法：描述性名称
```go
type Reader interface { Read() }
type DataSource interface { Read(); Write() }
```

### 1.3 错误处理

**必须检查错误：**
```go
result, err := DoSomething()
if err != nil {
    return err
}
```

**错误信息规范：**
```go
// ✅ 小写开头，不加标点
return fmt.Errorf("failed to connect database: %w", err)

// ❌ 大写开头，加标点
return fmt.Errorf("Failed to connect database.", err)
```

**自定义错误类型：**
```go
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}
```

### 1.4 函数设计

**函数长度：**
- 建议不超过 50 行
- 单一职责

**参数数量：**
- 不超过 5 个
- 过多时使用结构体

```go
// ✅ 使用配置结构体
type Config struct {
    Host     string
    Port     int
    Timeout  time.Duration
}

func NewClient(cfg Config) *Client {}

// ❌ 参数过多
func NewClient(host string, port int, timeout time.Duration, ...) *Client {}
```

### 1.5 测试规范

**测试文件命名：**
```
<package>_test.go
```

**测试函数命名：**
```go
func TestUserService_GetUser(t *testing.T) {}
func TestUserService_GetUser_NotFound(t *testing.T) {}
```

**表格驱动测试：**
```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"add positive", 1, 2, 3},
        {"add negative", -1, -2, -3},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.want {
                t.Errorf("Add() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## 2. TypeScript 代码规范

### 2.1 格式化

使用 Prettier 自动格式化：

```bash
npm run lint:fix
```

**.prettierrc:**
```json
{
  "semi": false,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5"
}
```

### 2.2 命名规范

**变量和函数：**
- 驼峰式（camelCase）
```typescript
const userName = 'John'
function getUserInfo() {}
```

**组件：**
- 大驼峰式（PascalCase）
```typescript
function UserProfile() {}
const UserProfile = () => {}
```

**常量和枚举：**
- 全大写，下划线分隔
```typescript
const MAX_RETRY_COUNT = 3
enum UserRole { ADMIN = 'admin', USER = 'user' }
```

**类型和接口：**
- 大驼峰式
```typescript
interface UserInfo {}
type UserRole = 'admin' | 'user'
```

### 2.3 类型规范

**显式类型声明：**
```typescript
// ✅ 函数返回值
function getUser(id: string): User | null {}

// ✅ 变量类型
const users: User[] = []
```

**避免 any：**
```typescript
// ❌ 避免使用 any
function process(data: any) {}

// ✅ 使用 unknown 或具体类型
function process(data: unknown) {
    if (typeof data === 'string') {
        // 处理字符串
    }
}
```

**类型别名 vs 接口：**
```typescript
// 对象类型优先使用 interface
interface User {
    id: string
    name: string
}

// 联合类型使用 type
type Status = 'pending' | 'success' | 'error'
```

### 2.4 React 规范

**组件结构：**
```typescript
import React, { useState, useEffect } from 'react'
import styles from './UserProfile.module.css'

interface Props {
    userId: string
    showAvatar?: boolean
}

export function UserProfile({ userId, showAvatar = true }: Props) {
    const [user, setUser] = useState<User | null>(null)
    
    useEffect(() => {
        // 加载用户
    }, [userId])
    
    return (
        <div className={styles.container}>
            {/* JSX */}
        </div>
    )
}
```

**Hooks 规则：**
- 只在顶层调用
- 只在 React 函数中调用
- 自定义 Hooks 以 `use` 开头

### 2.5 错误处理

**Try-Catch：**
```typescript
try {
    await api.getUser(id)
} catch (error) {
    if (error instanceof ApiError) {
        handleError(error)
    }
}
```

**错误边界：**
```typescript
class ErrorBoundary extends React.Component {
    state = { hasError: false }
    
    static getDerivedStateFromError() {
        return { hasError: true }
    }
    
    render() {
        if (this.state.hasError) {
            return <FallbackUI />
        }
        return this.props.children
    }
}
```

---

## 3. 通用规范

### 3.1 代码组织

**文件结构：**
```
src/
├── components/     # 可复用组件
├── pages/         # 页面组件
├── hooks/         # 自定义 Hooks
├── utils/         # 工具函数
├── types/         # 类型定义
└── api/           # API 客户端
```

**导入顺序：**
```typescript
// 1. 标准库
import React from 'react'

// 2. 第三方库
import axios from 'axios'

// 3. 内部模块
import { User } from '@/types'
import { api } from '@/api'

// 4. 相对路径
import styles from './UserProfile.module.css'
```

### 3.2 代码质量

**DRY 原则：**
- 避免重复代码
- 提取公共逻辑

**KISS 原则：**
- 保持简单
- 避免过度设计

**YAGNI 原则：**
- 不实现不需要的功能
- 按需扩展

---

## 4. 命名约定

### 4.1 布尔变量

使用肯定的布尔值：
```typescript
const isVisible = true      // ✅
const hidden = false        // ❌

const hasPermission = true  // ✅
const noPermission = false  // ❌
```

**前缀：**
- `is` - 状态
- `has` - 拥有
- `can` - 能力
- `should` - 应该

### 4.2 集合命名

使用复数形式：
```typescript
const users = []        // ✅
const userList = []     // ⚠️ 可接受
const userArray = []    // ❌
```

### 4.3 函数命名

**动词 + 名词：**
```typescript
function getUser() {}
function createUser() {}
function updateUser() {}
function deleteUser() {}
```

**布尔返回：**
```typescript
function isValid() {}
function hasPermission() {}
function canEdit() {}
```

---

## 5. 注释规范

### 5.1 文档注释

**Go:**
```go
// GetUser 根据 ID 获取用户
// 
// 参数:
//   - id: 用户 ID
// 
// 返回:
//   - user: 用户对象
//   - err: 错误信息
func GetUser(id string) (*User, error)
```

**TypeScript:**
```typescript
/**
 * 获取用户信息
 * @param id - 用户 ID
 * @returns 用户对象
 * @throws {NotFoundError} 用户不存在
 */
function getUser(id: string): Promise<User>
```

### 5.2 行内注释

**解释为什么，而不是做什么：**
```typescript
// ❌ 冗余注释
i++ // i 加 1

// ✅ 解释原因
// 从 0 开始索引，所以需要加 1
const actualIndex = i + 1
```

**TODO 注释：**
```typescript
// TODO: 优化性能，当前复杂度 O(n²)
function processData() {}

// FIXME: 处理时区问题
function convertTime() {}
```

---

## 6. 审查检查清单

### 6.1 代码审查

- [ ] 遵循命名规范
- [ ] 代码格式化
- [ ] 错误处理完整
- [ ] 边界条件考虑
- [ ] 测试覆盖
- [ ] 注释清晰

### 6.2 性能检查

- [ ] 无不必要的循环
- [ ] 合理使用缓存
- [ ] 避免内存泄漏
- [ ] 异步处理适当

### 6.3 安全检查

- [ ] 输入验证
- [ ] SQL 注入防护
- [ ] XSS 防护
- [ ] 敏感信息不泄露

---

## 📚 相关资源

- [Effective Go](https://golang.org/doc/effective_go)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Clean Code](https://www.amazon.com/Clean-Code-Handbook-Software-Craftsmanship/dp/0132350882)

---

*最后更新：2026-02-27*
