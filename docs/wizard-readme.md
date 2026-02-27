# RoleCraft AI 引导式创建向导

## 📖 项目概述

引导式创建向导（RoleWizard）是 RoleCraft AI 的核心功能之一，通过**问答对话**代替复杂的提示词编写，让用户在 3 步之内轻松创建专业的 AI 角色。

### 核心理念
> **用问答代替提示词，让人人都能创建专属 AI 助手**

---

## 🎯 功能特性

### 1. 分步问答流程
- ✅ **第 1 步：基础信息** - 名称、用途、风格
- ✅ **第 2 步：能力配置** - 专业领域、应避免事项、特殊要求
- ✅ **第 3 步：测试优化** - 实时测试、满意度评分、一键完成

### 2. 智能推荐系统
- 💡 基于选择推荐相关配置
- ⚠️ 特定领域的风险提醒
- ✨ 行业最佳实践建议

### 3. UI/UX 优化
- 🎨 友好的问题描述和示例
- 📊 清晰的进度指示器
- 💾 进度自动保存
- 🎉 完成庆祝动画

### 4. 后端支持
- 🤖 问答转提示词算法
- 🧪 测试对话接口
- 📦 配置导出优化

---

## 📁 文件结构

```
rolecraft-ai/
├── frontend/
│   └── src/
│       ├── components/
│       │   └── RoleWizard.tsx          # 前端向导组件（~600 行）
│       ├── pages/
│       │   └── RoleMarket.tsx          # 添加向导入口
│       └── App.tsx                      # 添加路由
│
├── backend/
│   └── internal/
│       ├── service/
│       │   └── prompt/
│       │       └── generator.go         # 提示词生成服务（~500 行）
│       └── api/
│           └── handler/
│               └── wizard.go            # API 处理器（~200 行）
│
└── docs/
    └── wizard-guide.md                  # 用户使用指南
```

---

## 🚀 快速开始

### 前端使用

```tsx
import { RoleWizard } from './components/RoleWizard';

// 在路由中添加
<Route path="/roles/wizard" element={<RoleWizard />} />
```

### 后端 API

```bash
# 获取配置选项
GET /api/v1/wizard/options

# 生成提示词
POST /api/v1/wizard/generate
Content-Type: application/json

{
  "name": "营销专家",
  "purpose": "creator",
  "style": "humorous",
  "expertise": ["marketing"],
  "avoidances": ["overpromise"]
}

# 获取智能推荐
POST /api/v1/wizard/recommendations

# 运行测试
POST /api/v1/wizard/test

# 导出配置
POST /api/v1/wizard/export
```

---

## 🔧 技术实现

### 前端技术栈
- **React 18** + TypeScript
- **Tailwind CSS** - 样式
- **Lucide Icons** - 图标
- **localStorage** - 进度保存

### 后端技术栈
- **Go 1.21+**
- **Gin** - Web 框架
- **结构化提示词生成** - 模板引擎

### 核心算法

#### 1. 提示词生成流程
```
用户输入 → 数据验证 → 模板匹配 → 参数调整 → 免责声明 → 最终输出
```

#### 2. 智能推荐逻辑
```go
func GetRecommendations(data WizardData) []Recommendation {
    recommendations = []
    
    // 基于用途推荐
    recommendations += getPurposeRecommendations(data)
    
    // 基于风格推荐
    recommendations += getStyleRecommendations(data)
    
    // 基于专业领域推荐
    recommendations += getExpertiseRecommendations(data)
    
    // 基于避免事项推荐
    recommendations += getAvoidanceRecommendations(data)
    
    return recommendations
}
```

#### 3. 测试评分机制
```go
func evaluateResponse(data, message, response) float64 {
    score = 70.0  // 基础分
    
    // 响应长度评分
    if len(response) in [20, 500]:
        score += 10
    
    // 包含用户消息关键词
    if message in response:
        score += 10
    
    // 风格匹配加分
    score += styleMatchBonus
    
    // 角色名提及加分
    if data.Name in response:
        score += 5
    
    return min(score, 100)
}
```

---

## 📊 数据结构

### WizardData（TypeScript）
```typescript
interface WizardData {
  // 第 1 步：基础信息
  name: string;
  purpose: string;      // assistant|expert|creator|teacher|companion|analyst
  style: string;        // professional|friendly|humorous|concise|detailed|inspirational
  
  // 第 2 步：能力配置
  expertise: string[];  // business|marketing|tech|design|finance|legal|hr|health|education|lifestyle
  avoidances: string[]; // speculation|repetition|jargon|controversy|overpromise|bias
  specialRequirements: string;
  
  // 第 3 步：测试
  testMessage: string;
  testResponse: string;
  satisfaction: number | null;
}
```

### GeneratedPrompt（Go）
```go
type GeneratedPrompt struct {
    SystemPrompt   string                 `json:"systemPrompt"`
    WelcomeMessage string                 `json:"welcomeMessage"`
    ModelConfig    map[string]interface{} `json:"modelConfig"`
    Metadata       PromptMetadata         `json:"metadata"`
}

type PromptMetadata struct {
    Version         string    `json:"version"`
    GeneratedAt     time.Time `json:"generatedAt"`
    WordCount       int       `json:"wordCount"`
    EstimatedTokens int       `json:"estimatedTokens"`
}
```

---

## 🎨 UI 设计

### 配色方案
```css
/* 主色调 */
--primary: #3b82f6;       /* blue-500 */
--primary-dark: #2563eb;  /* blue-600 */

/* 中性色 */
--slate-900: #0f172a;
--slate-700: #334155;
--slate-500: #64748b;
--slate-200: #e2e8f0;
```

### 组件层次
```
1. Header - 标题和说明
2. Step Navigation - 步骤导航（圆环 + 连线）
3. Step Content - 当前步骤内容
   - 问题描述
   - 选项卡片（可点击）
   - 输入框
   - 智能推荐面板
4. Navigation Buttons - 上一步/下一步/完成
5. Progress Hint - 保存提示
```

### 响应式设计
- **Desktop**: 3 列布局，完整功能
- **Tablet**: 2 列布局，优化触控
- **Mobile**: 1 列布局，简化交互

---

## 🧪 测试指南

### 单元测试（Go）
```go
func TestGeneratePrompt(t *testing.T) {
    generator := NewPromptGenerator()
    
    data := WizardData{
        Name:      "测试助手",
        Purpose:   "assistant",
        Style:     "professional",
        Expertise: []string{"business"},
    }
    
    result := generator.GeneratePrompt(data)
    
    if result.SystemPrompt == "" {
        t.Error("SystemPrompt should not be empty")
    }
    
    if !strings.Contains(result.SystemPrompt, "测试助手") {
        t.Error("SystemPrompt should contain role name")
    }
}
```

### E2E 测试（Playwright）
```typescript
test('create role with wizard', async ({ page }) => {
  await page.goto('/roles/wizard');
  
  // Step 1
  await page.fill('input[placeholder*="名称"]', '测试助手');
  await page.click('button:has-text("智能助理")');
  await page.click('button:has-text("专业严谨")');
  await page.click('button:has-text("下一步")');
  
  // Step 2
  await page.click('button:has-text("商务办公")');
  await page.click('button:has-text("下一步")');
  
  // Step 3
  await page.fill('input[placeholder*="测试"]', '你好');
  await page.click('button:has-text("发送")');
  await page.click('button:has-text("完成创建")');
  
  // Verify
  await expect(page.locator('text=创建成功')).toBeVisible();
});
```

---

## 📈 性能优化

### 前端优化
1. **懒加载**: 组件按需加载
2. **本地缓存**: localStorage 保存进度
3. **防抖处理**: 输入框防抖（300ms）
4. **虚拟滚动**: 大量选项时启用

### 后端优化
1. **模板缓存**: 提示词模板预加载
2. **并发处理**: 推荐计算并行化
3. **响应压缩**: gzip 压缩传输

---

## 🔒 安全考虑

### 输入验证
- ✅ 名称长度限制（1-20 字）
- ✅ 特殊字符过滤
- ✅ XSS 防护
- ✅ SQL 注入防护

### 数据保护
- ✅ 本地数据加密存储
- ✅ 敏感信息脱敏
- ✅ API 访问鉴权

---

## 🐛 已知问题

### v1.0.0
- ⚠️ 移动端键盘弹出时布局可能错位
- ⚠️ 超长特殊要求（>500 字）未截断
- ⚠️ 测试对话暂未调用真实 AI API

### 待优化
- 🔄 支持导入导出角色配置
- 🔄 支持多人协作创建
- 🔄 增加更多预设模板

---

## 📝 开发日志

### 2024-02-27
- ✅ 完成前端 RoleWizard 组件
- ✅ 完成后端 generator.go 服务
- ✅ 完成 API handler 和路由
- ✅ 编写用户使用指南
- ✅ 集成到现有系统

### 2024-03-01 (计划)
- 🔄 接入真实 AI API 进行测试
- 🔄 添加更多智能推荐规则
- 🔄 优化移动端体验

### 2024-03-05 (计划)
- 🔄 添加 A/B 测试功能
- 🔄 支持配置模板市场
- 🔄 增加数据分析面板

---

## 🤝 贡献指南

### 代码规范
- **Frontend**: ESLint + Prettier
- **Backend**: gofmt + go vet
- **Commits**: Conventional Commits

### 提交流程
1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📄 许可证

MIT License - 详见 [LICENSE](../LICENSE)

---

## 📞 联系方式

- **项目负责人**: RoleCraft AI Team
- **邮箱**: support@rolecraft.ai
- **文档**: https://docs.rolecraft.ai/wizard

---

**最后更新**: 2024-02-27  
**版本**: 1.0.0
