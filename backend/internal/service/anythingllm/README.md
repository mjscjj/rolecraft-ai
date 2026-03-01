# AnythingLLM Client

Go 语言实现的 AnythingLLM API 客户端封装。

## 功能特性

- ✅ 完整的 API 方法封装
- ✅ 自动重试机制（最多 3 次）
- ✅ 错误处理
- ✅ 流式响应支持（SSE）
- ✅ 上下文支持
- ✅ 单元测试覆盖

## 安装

作为 `rolecraft-ai` 项目的一部分，无需单独安装。

## 使用方法

### 初始化客户端

```go
import "rolecraft-ai/backend/internal/service/anythingllm"

// 使用默认配置
client := anythingllm.NewAnythingLLMClient(
    "http://43.134.234.4:3001/api",
    "sk-WaUmgZsMxgeHOpp8SJxK1rmVQxiwfiDJ",
)
```

### 创建工作区

```go
workspace, err := client.CreateWorkspace(
    "user123",           // 用户 ID
    "My Workspace",      // 工作区名称
    "You are a helpful assistant", // 系统提示
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created workspace: %s\n", workspace.Name)
```

### 获取工作区

```go
workspace, err := client.GetWorkspace("user123")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Workspace status: %s\n", workspace.Status)
```

### 对话

```go
// 普通对话
response, err := client.Chat("user123", "Hello!", "chat")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
```

### 流式对话

```go
err := client.StreamChat("user123", "Tell me a story", "chat", func(chunk string) {
    fmt.Print(chunk) // 实时打印响应
})
if err != nil {
    log.Fatal(err)
}
```

### 上传文档

```go
fileData, err := os.ReadFile("document.pdf")
if err != nil {
    log.Fatal(err)
}

response, err := client.UploadDocument("user123", "document.pdf", fileData)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Uploaded: %s\n", response.FileName)
```

### 获取文档列表

```go
docs, err := client.GetDocuments("user123")
if err != nil {
    log.Fatal(err)
}

for _, doc := range docs {
    fmt.Printf("- %s (%s)\n", doc.DocName, doc.DocType)
}
```

### 删除文档

```go
err := client.DeleteDocument("user123", "doc-hash-123")
if err != nil {
    log.Fatal(err)
}
```

### 向量搜索

```go
results, err := client.VectorSearch("user123", "machine learning", 5)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("Score: %.2f - %s\n", result.Score, result.DocName)
}
```

### 获取对话历史

```go
history, err := client.GetChatHistory("user123", 20)
if err != nil {
    log.Fatal(err)
}

for _, item := range history {
    fmt.Printf("Q: %s\nA: %s\n\n", item.Prompt, item.Response)
}
```

## API 方法

| 方法 | 描述 | 参数 |
|------|------|------|
| `NewAnythingLLMClient(baseURL, apiKey)` | 创建客户端 | baseURL: API 地址，apiKey: 认证密钥 |
| `CreateWorkspace(userId, name, systemPrompt)` | 创建工作区 | userId: 用户 ID，name: 名称，systemPrompt: 系统提示 |
| `GetWorkspace(userId)` | 获取工作区 | userId: 用户 ID |
| `Chat(userId, message, mode)` | 对话 | userId: 用户 ID，message: 消息，mode: 模式 (chat/query) |
| `StreamChat(userId, message, mode, callback)` | 流式对话 | callback: 接收响应块的函数 |
| `UploadDocument(userId, fileName, fileData)` | 上传文档 | fileData: 文件字节数据 |
| `GetDocuments(userId)` | 获取文档列表 | - |
| `DeleteDocument(userId, docHash)` | 删除文档 | docHash: 文档哈希 |
| `VectorSearch(userId, query, topN)` | 向量搜索 | query: 搜索词，topN: 返回数量 |
| `GetChatHistory(userId, limit)` | 获取对话历史 | limit: 限制数量（0=无限制） |

## 错误处理

所有方法返回 `error`，调用方应检查错误：

```go
workspace, err := client.GetWorkspace("user123")
if err != nil {
    // 处理错误
    log.Printf("Failed to get workspace: %v", err)
    return
}
```

常见错误类型：
- 网络错误：连接失败、超时
- API 错误：4xx/5xx 响应
- 解析错误：无效的 JSON 响应

## 重试机制

客户端内置自动重试：
- 最多重试 3 次
- 仅对服务器错误（5xx）重试
- 客户端错误（4xx）不重试
- 重试延迟递增（1s, 2s, 3s）

## 测试

运行单元测试：

```bash
cd rolecraft-ai/backend
go test ./internal/service/anythingllm/... -v
```

运行特定测试：

```bash
go test ./internal/service/anythingllm/... -run TestChat -v
```

生成测试覆盖率报告：

```bash
go test ./internal/service/anythingllm/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 配置

使用自定义配置创建客户端：

```go
config := anythingllm.ClientConfig{
    BaseURL:    "http://localhost:3001/api",
    APIKey:     "your-api-key",
    Timeout:    60 * time.Second,
    MaxRetries: 5,
    RetryDelay: 2 * time.Second,
}

client := anythingllm.NewAnythingLLMClientWithConfig(config)
```

## 注意事项

1. **Workspace Slug**: 工作区标识符格式为 `user_{userId}`
2. **API Key**: 需要有效的 AnythingLLM API 密钥
3. **超时**: 默认 HTTP 超时为 30 秒，流式对话为 120 秒
4. **并发**: 客户端实例可安全并发使用

## AnythingLLM API 参考

- 官方文档：https://anythingllm.com/api
- 服务器地址：`http://43.134.234.4:3001/api`

## 许可证

与 `rolecraft-ai` 项目保持一致。
