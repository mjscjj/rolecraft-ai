package ai

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// MockAIClient Mock AI 客户端 - 用于开发和测试
type MockAIClient struct {
	responses map[string][]string
}

// NewMockAIClient 创建 Mock AI 客户端
func NewMockAIClient() *MockAIClient {
	return &MockAIClient{
		responses: map[string][]string{
			"greeting": {
				"你好！很高兴为你服务。请问有什么可以帮你的？",
				"嗨！我是你的 AI 助手，随时准备帮助你解决问题。",
				"你好呀！今天想聊点什么？",
			},
			"marketing": {
				"好的！针对这个营销需求，我建议从以下几个角度入手：\n\n1. **目标受众分析** - 明确你的核心用户群体\n2. **价值主张** - 突出产品的独特卖点\n3. **渠道选择** - 根据用户习惯选择合适的推广渠道\n4. **内容策略** - 创作有吸引力的营销内容\n\n需要我详细展开哪个部分？",
				"这个营销想法很不错！我来帮你完善一下：\n\n📊 **市场定位**\n- 目标人群：25-35 岁都市白领\n- 核心需求：高效、便捷、品质\n\n💡 **创意方向**\n- 情感共鸣：讲述真实用户故事\n- 数据支撑：展示产品效果对比\n- 社交传播：设计互动话题\n\n需要我帮你写具体的文案吗？",
			},
			"writing": {
				"没问题！我来帮你写：\n\n---\n\n📝 **标题**：让每一天都充满可能\n\n正文：\n在这个快节奏的时代，我们都在寻找一种平衡——工作与生活的平衡，理想与现实的平衡。\n\n我们的产品，就是为了帮你找到这种平衡而生。\n\n✨ 为什么选择我们？\n- 高效：节省 50% 的时间\n- 简单：3 步即可完成\n- 可靠：99.9% 的用户满意度\n\n现在就开始体验吧！\n\n---\n\n需要调整语气或内容吗？",
				"好的，这是一份文案草稿：\n\n🎯 **核心信息**\n我们的产品能帮你解决 [具体问题]，让你 [获得具体好处]。\n\n📖 **故事线**\n1. 痛点场景描述\n2. 解决方案引入\n3. 使用效果展示\n4. 行动号召\n\n需要我针对某个平台（朋友圈/微博/公众号）优化吗？",
			},
			"analysis": {
				"让我来分析一下：\n\n📈 **数据洞察**\n从你提供的信息来看，有几个关键点值得注意：\n\n1. 趋势向上，但增速放缓\n2. 用户留存率表现良好\n3. 转化率有提升空间\n\n💡 **建议**\n- 优化 onboarding 流程\n- 加强用户教育\n- 测试不同的定价策略\n\n需要我深入分析哪个指标？",
				"这个分析很有意思！我的看法：\n\n✅ **优势**\n- 市场定位清晰\n- 产品差异化明显\n- 团队执行力强\n\n⚠️ **风险**\n- 竞争加剧\n- 获客成本上升\n- 技术迭代快\n\n🎯 **下一步**\n建议优先验证 PMF，再考虑规模化扩张。",
			},
			"code": {
				"好的，这是一个示例代码：\n\n```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```\n\n这段代码实现了基础功能。如果需要添加错误处理或扩展功能，告诉我具体需求。",
				"我来帮你写这段代码：\n\n```python\ndef process_data(data):\n    \"\"\"处理数据的核心函数\"\"\"\n    result = []\n    for item in data:\n        if item.get('valid'):\n            result.append(transform(item))\n    return result\n```\n\n需要添加单元测试吗？",
			},
			"default": {
				"明白了！我来帮你处理这个请求。请给我一点时间思考最佳方案...",
				"收到！这个问题很有意思，让我仔细分析一下...",
				"好的，我理解你的需求了。基于我的经验，我建议...",
				"没问题！让我来帮你解决这个问题。",
			},
		},
	}
}

// ChatCompletion Mock 聊天完成
func (m *MockAIClient) ChatCompletion(ctx context.Context, messages []ChatMessage, temperature float64) (*ChatResponse, error) {
	// 模拟处理延迟
	time.Sleep(500 * time.Millisecond)

	// 获取最后一条用户消息
	var lastUserMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserMessage = messages[i].Content
			break
		}
	}

	// 根据消息内容选择合适的回复类别
	category := m.categorizeMessage(lastUserMessage)
	responses := m.responses[category]
	if len(responses) == 0 {
		responses = m.responses["default"]
	}

	// 随机选择一个回复
	response := responses[rand.Intn(len(responses))]

	return &ChatResponse{
		ID:      fmt.Sprintf("mock-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "mock-v1",
		Choices: []struct {
			Index        int          `json:"index"`
			Message      ChatMessage  `json:"message"`
			Delta        *ChatMessage `json:"delta,omitempty"`
			FinishReason string       `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: response,
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     10,
			CompletionTokens: len([]rune(response)) / 4,
			TotalTokens:      len([]rune(response)) / 4 + 10,
		},
	}, nil
}

// ChatCompletionStream Mock 流式聊天
func (m *MockAIClient) ChatCompletionStream(ctx context.Context, messages []ChatMessage, temperature float64) (<-chan *StreamChunk, <-chan error) {
	chunkChan := make(chan *StreamChunk, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)

		// 获取回复
		var lastUserMessage string
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				lastUserMessage = messages[i].Content
				break
			}
		}

		category := m.categorizeMessage(lastUserMessage)
		responses := m.responses[category]
		if len(responses) == 0 {
			responses = m.responses["default"]
		}
		response := responses[rand.Intn(len(responses))]

		// 模拟流式输出
		chunkSize := 3
		for i := 0; i < len(response); i += chunkSize {
			end := i + chunkSize
			if end > len(response) {
				end = len(response)
			}
			chunk := response[i:end]

			chunkChan <- &StreamChunk{
				ID:      fmt.Sprintf("mock-%d", time.Now().UnixNano()),
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   "mock-v1",
				Choices: []struct {
					Index        int         `json:"index"`
					Delta        ChatMessage `json:"delta"`
					FinishReason string      `json:"finish_reason"`
				}{
					{
						Index: 0,
						Delta: ChatMessage{
							Role:    "assistant",
							Content: chunk,
						},
						FinishReason: "",
					},
				},
			}

			time.Sleep(50 * time.Millisecond)
		}

		// 发送结束标记
		chunkChan <- &StreamChunk{
			Choices: []struct {
				Index        int         `json:"index"`
				Delta        ChatMessage `json:"delta"`
				FinishReason string      `json:"finish_reason"`
			}{
				{
					FinishReason: "stop",
				},
			},
		}
	}()

	return chunkChan, errChan
}

// categorizeMessage 根据消息内容分类
func (m *MockAIClient) categorizeMessage(message string) string {
	message = strings.ToLower(message)

	// 问候
	if strings.Contains(message, "你好") || strings.Contains(message, "嗨") || strings.Contains(message, "hello") || strings.Contains(message, "hi") {
		return "greeting"
	}

	// 营销相关
	if strings.Contains(message, "营销") || strings.Contains(message, "推广") || strings.Contains(message, "市场") || strings.Contains(message, "品牌") {
		return "marketing"
	}

	// 写作相关
	if strings.Contains(message, "写") || strings.Contains(message, "文案") || strings.Contains(message, "文章") || strings.Contains(message, "邮件") {
		return "writing"
	}

	// 分析相关
	if strings.Contains(message, "分析") || strings.Contains(message, "数据") || strings.Contains(message, "报告") || strings.Contains(message, "趋势") {
		return "analysis"
	}

	// 代码相关
	if strings.Contains(message, "代码") || strings.Contains(message, "编程") || strings.Contains(message, "function") || strings.Contains(message, "code") {
		return "code"
	}

	return "default"
}
