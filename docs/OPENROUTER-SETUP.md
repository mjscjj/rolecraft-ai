# RoleCraft AI - OpenRouter 配置指南

**更新日期**: 2026-02-27  
**状态**: ✅ 生产就绪

---

## 🎯 概述

RoleCraft AI 现已集成 **OpenRouter**，支持访问 100+ 种 AI 模型，包括：
- Google Gemini 系列
- OpenAI GPT 系列
- Anthropic Claude 系列
- DeepSeek 系列
- Qwen 通义千问系列
- MiniMax 等

---

## 📋 OpenRouter 配置信息

### 从 OpenClaw 配置迁移

**配置文件位置**: `/Users/claw/.openclaw/openclaw.json`

**OpenRouter 配置**:
```json
{
  "baseUrl": "https://openrouter.ai/api/v1",
  "apiKey": "sk-or-v1-3592fb02bc6293692a756d866ba34ba92543f2823469c8783e71542931c950"
}
```

---

## 🚀 快速开始

### 方式 1：使用 OpenClaw 配置（推荐）

**步骤 1: 创建 `.env` 文件**
```bash
cd /Users/claw/.openclaw/workspace-work/rolecraft-ai/backend
cp .env.example .env
```

**步骤 2: 编辑 `.env` 文件**
```bash
vim .env
```

**填入配置**:
```env
# OpenRouter 配置（使用 OpenClaw 的密钥）
OPENROUTER_URL=https://openrouter.ai/api/v1
OPENROUTER_KEY=sk-or-v1-3592fb02bc6293692a756d866ba34ba92543f2823469c8783e71542931c950
OPENROUTER_MODEL=google/gemini-3-flash-preview
```

**步骤 3: 重启后端**
```bash
# 停止当前服务
pkill -f "go run cmd/server"

# 重新启动
go run cmd/server/main.go
```

---

### 方式 2：手动配置

**步骤 1: 获取 OpenRouter API Key**
1. 访问 https://openrouter.ai/keys
2. 登录/注册账号
3. 创建新的 API Key
4. 复制密钥

**步骤 2: 配置环境变量**
```bash
# 临时设置（当前终端有效）
export OPENROUTER_KEY=sk-or-v1-YOUR_KEY
export OPENROUTER_MODEL=google/gemini-3-flash-preview

# 永久设置（添加到 ~/.zshrc 或 ~/.bashrc）
echo 'export OPENROUTER_KEY=sk-or-v1-YOUR_KEY' >> ~/.zshrc
echo 'export OPENROUTER_MODEL=google/gemini-3-flash-preview' >> ~/.zshrc
source ~/.zshrc
```

---

## 🤖 可用模型列表

### 推荐模型

| 模型 ID | 名称 | 上下文 | 推理 | 适用场景 |
|---------|------|--------|------|----------|
| `google/gemini-3-flash-preview` | Gemini 3 Flash | 200K | ❌ | 快速对话 ⭐ |
| `google/gemini-3-pro-preview` | Gemini 3 Pro | 200K | ❌ | 复杂任务 |
| `google/gemini-3.1-pro-preview` | Gemini 3.1 Pro | 1M | ❌ | 超长文档 |
| `anthropic/claude-opus-4.6` | Claude Opus 4.6 | 200K | ✅ | 深度思考 |
| `anthropic/claude-sonnet-4.6` | Claude Sonnet 4.6 | 200K | ❌ | 日常对话 |
| `deepseek/deepseek-v3.2-speciale` | DeepSeek V3.2 | 200K | ❌ | 中文优化 |
| `z-ai/glm-5` | GLM-5 | 200K | ✅ | 推理任务 |

### 免费模型

| 模型 ID | 名称 | 限制 |
|---------|------|------|
| `deepseek/deepseek-r1-0528:free` | DeepSeek R1 Free | 免费额度 |
| `google/gemini-3-flash-preview` | Gemini 3 Flash | 免费额度 |
| `minimax/minimax-m2.5` | MiniMax M2.5 | 免费额度 |

---

## 🔧 切换模型

### 方式 1：环境变量

```bash
# 使用 Gemini 3 Pro
export OPENROUTER_MODEL=google/gemini-3-pro-preview

# 使用 Claude Opus
export OPENROUTER_MODEL=anthropic/claude-opus-4.6

# 使用 DeepSeek
export OPENROUTER_MODEL=deepseek/deepseek-v3.2-speciale
```

### 方式 2：代码中配置

```go
import "rolecraft-ai/internal/service/ai"

// 创建客户端时指定模型
client := ai.NewOpenRouterClient(ai.OpenRouterConfig{
    APIKey:  "sk-or-v1-xxx",
    BaseURL: "https://openrouter.ai/api/v1",
    Model:   "google/gemini-3-pro-preview",
})

// 或运行时切换
client.SetModel("anthropic/claude-opus-4.6")
```

---

## 📊 性能对比

### 响应速度

| 模型 | 平均响应时间 | 适用场景 |
|------|-------------|----------|
| Gemini 3 Flash | < 1s | 快速问答 ⭐ |
| Gemini 3 Pro | 1-2s | 复杂任务 |
| Claude Sonnet | 1-2s | 日常对话 |
| Claude Opus | 2-3s | 深度思考 |
| DeepSeek V3.2 | 1-2s | 中文优化 |

### 成本对比

| 模型 | 输入价格 | 输出价格 | 免费额度 |
|------|---------|---------|----------|
| Gemini 3 Flash | $0 | $0 | ✅ |
| Gemini 3 Pro | $0 | $0 | ✅ |
| Claude Opus | $0 | $0 | ✅ |
| DeepSeek V3.2 | $0 | $0 | ✅ |

**注**: OpenRouter 提供免费额度，个人使用基本够用

---

## 🎯 使用示例

### 1. 普通对话

```bash
curl -X POST http://localhost:8080/api/v1/chat/SESSION_ID/complete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "content": "你好，请介绍一下你自己"
  }'
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "userMessage": {...},
    "assistantMessage": {
      "content": "你好！我是 RoleCraft AI 助手...",
      "model": "google/gemini-3-flash-preview"
    }
  }
}
```

---

### 2. 流式对话

```bash
curl -X POST http://localhost:8080/api/v1/chat/SESSION_ID/stream \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "content": "如何学习 Go 语言？"
  }'
```

**流式响应** (SSE):
```
data: {"type":"chunk","data":{"content":"学"}}
data: {"type":"chunk","data":{"content":"习"}}
data: {"type":"chunk","data":{"content":"G"}}
data: {"type":"chunk","data":{"content":"o"}}
...
data: {"type":"done"}
```

---

### 3. 深度思考（带思考过程）

```bash
curl -X POST http://localhost:8080/api/v1/chat/SESSION_ID/stream-with-thinking \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "content": "如何设计一个高并发系统？"
  }'
```

**响应**:
```
data: {"type":"thinking","data":{"icon":"🤔","content":"理解问题..."} }
data: {"type":"thinking","data":{"icon":"🔍","content":"分析要素..."} }
data: {"type":"thinking","data":{"icon":"📚","content":"检索知识..."} }
data: {"type":"answer","data":{"content":"设计高并发系统需要考虑..."} }
```

---

## 🔍 故障排查

### 问题 1: API Key 无效

**错误**:
```
API error: 401 - {"error": "Invalid API key"}
```

**解决**:
```bash
# 检查 API Key 是否正确
echo $OPENROUTER_KEY

# 重新设置
export OPENROUTER_KEY=sk-or-v1-正确的密钥
```

---

### 问题 2: 模型不可用

**错误**:
```
API error: 400 - {"error": "Model not found"}
```

**解决**:
```bash
# 检查模型 ID 是否正确
echo $OPENROUTER_MODEL

# 使用可用模型
export OPENROUTER_MODEL=google/gemini-3-flash-preview
```

---

### 问题 3: 请求超时

**错误**:
```
context deadline exceeded
```

**解决**:
```bash
# 检查网络连接
curl https://openrouter.ai/api/v1

# 使用更快的模型
export OPENROUTER_MODEL=google/gemini-3-flash-preview
```

---

### 问题 4: 余额不足

**错误**:
```
API error: 402 - {"error": "Insufficient credits"}
```

**解决**:
1. 访问 https://openrouter.ai/credits
2. 充值账户
3. 或切换到免费模型

---

## 📈 监控用量

### 查看 API 用量

访问：https://openrouter.ai/activity

### 设置用量告警

1. 访问 https://openrouter.ai/settings
2. 设置每月预算上限
3. 启用邮件通知

---

## 🎨 最佳实践

### 1. 选择合适的模型

**快速问答**: `google/gemini-3-flash-preview`  
**复杂任务**: `google/gemini-3-pro-preview`  
**深度思考**: `anthropic/claude-opus-4.6`  
**中文优化**: `deepseek/deepseek-v3.2-speciale`

### 2. 控制成本

- 使用免费模型测试
- 设置预算上限
- 监控用量
- 缓存常用回复

### 3. 提升性能

- 使用流式响应
- 合理设置 temperature
- 控制 max_tokens
- 使用短上下文

---

## 📝 配置文件位置

| 文件 | 位置 | 说明 |
|------|------|------|
| `.env` | `backend/.env` | 环境变量配置 |
| `.env.example` | `backend/.env.example` | 配置示例 |
| `config.go` | `backend/internal/config/config.go` | 配置加载代码 |
| `openrouter.go` | `backend/internal/service/ai/openrouter.go` | OpenRouter 客户端 |

---

## 🔗 相关文档

- [OpenRouter 官方文档](https://openrouter.ai/docs)
- [可用模型列表](https://openrouter.ai/models)
- [定价说明](https://openrouter.ai/pricing)
- [API 参考](https://openrouter.ai/api-docs)

---

## ✅ 验证配置

**测试命令**:
```bash
# 检查环境变量
env | grep OPENROUTER

# 测试 API
curl -X POST https://openrouter.ai/api/v1/chat/completions \
  -H "Authorization: Bearer $OPENROUTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "google/gemini-3-flash-preview",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**预期响应**:
```json
{
  "id": "gen-xxx",
  "choices": [{
    "message": {"content": "Hello! How can I help you?"}
  }]
}
```

---

**配置完成，开始使用真正的 AI 对话！** 🎉
