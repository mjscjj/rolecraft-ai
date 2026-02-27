package prompt

import (
	"fmt"
	"strings"
	"time"
)

// ============ 数据类型 ============

// WizardData 向导数据模型
type WizardData struct {
	// 第 1 步：基础信息
	Name              string   `json:"name"`
	Purpose           string   `json:"purpose"`
	Style             string   `json:"style"`
	
	// 第 2 步：能力配置
	Expertise         []string `json:"expertise"`
	Avoidances        []string `json:"avoidances"`
	SpecialRequirements string `json:"specialRequirements"`
	
	// 第 3 步：测试
	TestMessage       string   `json:"testMessage"`
	TestResponse      string   `json:"testResponse"`
	Satisfaction      *int     `json:"satisfaction"` // 0=不满意，1=满意，nil=未评分
}

// PurposeOption 用途选项
type PurposeOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// StyleOption 风格选项
type StyleOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// ExpertiseOption 专业领域选项
type ExpertiseOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// AvoidanceOption 应避免事项选项
type AvoidanceOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// Recommendation 智能推荐
type Recommendation struct {
	Type        string `json:"type"`        // suggestion/warning/best_practice
	Priority    string `json:"priority"`    // high/medium/low
	Title       string `json:"title"`
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
}

// TestResult 测试结果
type TestResult struct {
	TestID      string    `json:"testId"`
	Input       string    `json:"input"`
	Output      string    `json:"output"`
	Score       float64   `json:"score"`
	Feedback    string    `json:"feedback"`
	Suggestions []string  `json:"suggestions"`
	Timestamp   time.Time `json:"timestamp"`
}

// GeneratedPrompt 生成的提示词
type GeneratedPrompt struct {
	SystemPrompt   string            `json:"systemPrompt"`
	WelcomeMessage string            `json:"welcomeMessage"`
	ModelConfig    map[string]interface{} `json:"modelConfig"`
	Metadata       PromptMetadata    `json:"metadata"`
}

// PromptMetadata 提示词元数据
type PromptMetadata struct {
	Version       string    `json:"version"`
	GeneratedAt   time.Time `json:"generatedAt"`
	WizardVersion string    `json:"wizardVersion"`
	WordCount     int       `json:"wordCount"`
	EstimatedTokens int     `json:"estimatedTokens"`
}

// ============ 配置数据 ============

var Purposes = []PurposeOption{
	{ID: "assistant", Label: "智能助理", Description: "处理日常事务、安排日程、撰写邮件", Icon: "📋"},
	{ID: "expert", Label: "专业顾问", Description: "提供专业领域的咨询和建议", Icon: "🎯"},
	{ID: "creator", Label: "内容创作", Description: "撰写文案、故事、营销内容", Icon: "✍️"},
	{ID: "teacher", Label: "教学辅导", Description: "知识讲解、学习辅导、技能培训", Icon: "📚"},
	{ID: "companion", Label: "情感陪伴", Description: "聊天解闷、情感支持、心理疏导", Icon: "💙"},
	{ID: "analyst", Label: "数据分析", Description: "数据处理、报告生成、商业分析", Icon: "📊"},
}

var Styles = []StyleOption{
	{ID: "professional", Label: "专业严谨", Description: "正式、准确、条理清晰", Icon: "👔"},
	{ID: "friendly", Label: "友好亲切", Description: "温暖、耐心、易于接近", Icon: "😊"},
	{ID: "humorous", Label: "幽默风趣", Description: "轻松、有趣、富有创意", Icon: "😄"},
	{ID: "concise", Label: "简洁直接", Description: "高效、直接、不啰嗦", Icon: "⚡"},
	{ID: "detailed", Label: "详细周全", Description: "全面、深入、注重细节", Icon: "📝"},
	{ID: "inspirational", Label: "激励鼓舞", Description: "积极、向上、充满能量", Icon: "🌟"},
}

var ExpertiseAreas = []ExpertiseOption{
	{ID: "business", Label: "商务办公", Description: "邮件、文档、会议、项目管理"},
	{ID: "marketing", Label: "市场营销", Description: "策划、文案、推广、品牌"},
	{ID: "tech", Label: "技术编程", Description: "开发、调试、架构、算法"},
	{ID: "design", Label: "创意设计", Description: "UI/UX、平面、创意构思"},
	{ID: "finance", Label: "财务金融", Description: "会计、投资、理财、税务"},
	{ID: "legal", Label: "法律法务", Description: "合同、合规、法律咨询"},
	{ID: "hr", Label: "人力资源", Description: "招聘、培训、绩效、员工关系"},
	{ID: "health", Label: "健康医疗", Description: "健身、营养、心理健康"},
	{ID: "education", Label: "教育培训", Description: "课程、辅导、学习方法"},
	{ID: "lifestyle", Label: "生活休闲", Description: "旅行、美食、购物、娱乐"},
}

var Avoidances = []AvoidanceOption{
	{ID: "speculation", Label: "猜测臆断", Description: "不确定的信息要明确说明"},
	{ID: "repetition", Label: "重复啰嗦", Description: "避免重复相同内容"},
	{ID: "jargon", Label: "专业术语", Description: "少用晦涩难懂的专业词汇"},
	{ID: "controversy", Label: "敏感话题", Description: "避开政治、宗教等敏感议题"},
	{ID: "overpromise", Label: "过度承诺", Description: "不夸大能力，诚实告知局限"},
	{ID: "bias", Label: "主观偏见", Description: "保持客观中立，不带个人偏见"},
}

// ============ 核心服务 ============

// PromptGenerator 提示词生成器
type PromptGenerator struct {
	version string
}

// NewPromptGenerator 创建提示词生成器
func NewPromptGenerator() *PromptGenerator {
	return &PromptGenerator{
		version: "1.0.0",
	}
}

// GeneratePrompt 生成完整的系统提示词
func (g *PromptGenerator) GeneratePrompt(data WizardData) GeneratedPrompt {
	systemPrompt := g.buildSystemPrompt(data)
	welcomeMessage := g.buildWelcomeMessage(data)
	modelConfig := g.buildModelConfig(data)
	
	metadata := PromptMetadata{
		Version:         g.version,
		GeneratedAt:     time.Now(),
		WizardVersion:   "1.0.0",
		WordCount:       len(strings.Fields(systemPrompt)),
		EstimatedTokens: len([]rune(systemPrompt)) / 4,
	}
	
	return GeneratedPrompt{
		SystemPrompt:   systemPrompt,
		WelcomeMessage: welcomeMessage,
		ModelConfig:    modelConfig,
		Metadata:       metadata,
	}
}

// buildSystemPrompt 构建系统提示词
func (g *PromptGenerator) buildSystemPrompt(data WizardData) string {
	var sb strings.Builder
	
	// 标题
	sb.WriteString(fmt.Sprintf("# 角色设定：%s\n\n", data.Name))
	
	// 核心定位
	purpose := g.getPurposeLabel(data.Purpose)
	style := g.getStyleLabel(data.Style)
	
	sb.WriteString("## 核心定位\n")
	sb.WriteString(fmt.Sprintf("你是一位%s的 AI 助手。你的主要职责是帮助用户%s。\n\n", 
		purpose, g.getPurposeDescription(data.Purpose)))
	
	// 说话风格
	sb.WriteString("## 说话风格\n")
	sb.WriteString(fmt.Sprintf("%s。在交流中，你应该%s。\n\n", 
		style, g.getStyleDescription(data.Style)))
	
	// 专业领域
	if len(data.Expertise) > 0 {
		sb.WriteString("## 专业领域\n")
		expertiseLabels := g.getExpertiseLabels(data.Expertise)
		sb.WriteString(fmt.Sprintf("你擅长以下领域：%s。在这些领域内，你应该提供专业、准确的建议和信息。\n\n",
			strings.Join(expertiseLabels, "、")))
	}
	
	// 应避免事项
	if len(data.Avoidances) > 0 {
		sb.WriteString("## 应避免事项\n")
		sb.WriteString("请注意避免以下情况：\n")
		for _, avoidanceID := range data.Avoidances {
			avoidance := g.getAvoidance(avoidanceID)
			if avoidance != nil {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", avoidance.Label, avoidance.Description))
			}
		}
		sb.WriteString("\n")
	}
	
	// 特殊要求
	if strings.TrimSpace(data.SpecialRequirements) != "" {
		sb.WriteString("## 特殊要求\n")
		sb.WriteString(data.SpecialRequirements + "\n\n")
	}
	
	// 行为准则
	sb.WriteString("## 行为准则\n")
	sb.WriteString("1. 始终以帮助用户为首要目标\n")
	sb.WriteString("2. 如遇不确定的信息，诚实告知而非猜测\n")
	sb.WriteString("3. 保持专业且友好的态度\n")
	sb.WriteString("4. 回答应清晰、有条理、实用\n\n")
	
	// 开始指令
	sb.WriteString("## 开始\n")
	sb.WriteString(fmt.Sprintf("现在，请以%s的身份，用%s的方式，开始为用户提供帮助。", 
		data.Name, style))
	
	return sb.String()
}

// buildWelcomeMessage 构建欢迎语
func (g *PromptGenerator) buildWelcomeMessage(data WizardData) string {
	purpose := g.getPurposeLabel(data.Purpose)
	style := data.Style
	
	var greeting string
	switch style {
	case "friendly":
		greeting = "很高兴见到你！"
	case "humorous":
		greeting = "哈喽！准备好一起做些有趣的事了吗？"
	case "inspirational":
		greeting = "你好！让我们一起创造精彩！"
	case "concise":
		greeting = "你好，请问有什么需要？"
	case "detailed":
		greeting = "你好！我很乐意为你提供详细的帮助和指导。"
	default:
		greeting = "你好！有什么可以帮助你的吗？"
	}
	
	if data.Name != "" {
		return fmt.Sprintf("你好！我是%s，%s%s", data.Name, purpose, greeting)
	}
	return fmt.Sprintf("你好！我是你的 AI 助手，%s%s", purpose, greeting)
}

// buildModelConfig 构建模型配置
func (g *PromptGenerator) buildModelConfig(data WizardData) map[string]interface{} {
	config := map[string]interface{}{
		"temperature": 0.7,
		"top_p":       0.9,
		"frequency_penalty": 0.5,
		"presence_penalty":  0.5,
	}
	
	// 根据风格调整参数
	switch data.Style {
	case "humorous", "inspirational":
		config["temperature"] = 0.8
		config["presence_penalty"] = 0.7
	case "professional", "concise":
		config["temperature"] = 0.6
		config["frequency_penalty"] = 0.7
	case "detailed":
		config["temperature"] = 0.7
		config["max_tokens"] = 2000
	}
	
	return config
}

// ============ 智能推荐 ============

// GetRecommendations 获取智能推荐
func (g *PromptGenerator) GetRecommendations(data WizardData) []Recommendation {
	var recommendations []Recommendation
	
	// 基于用途的推荐
	recommendations = append(recommendations, g.getPurposeRecommendations(data)...)
	
	// 基于风格的推荐
	recommendations = append(recommendations, g.getStyleRecommendations(data)...)
	
	// 基于专业领域的推荐
	recommendations = append(recommendations, g.getExpertiseRecommendations(data)...)
	
	// 基于避免事项的推荐
	recommendations = append(recommendations, g.getAvoidanceRecommendations(data)...)
	
	return recommendations
}

func (g *PromptGenerator) getPurposeRecommendations(data WizardData) []Recommendation {
	var recs []Recommendation
	
	switch data.Purpose {
	case "assistant":
		recs = append(recs, Recommendation{
			Type:        "suggestion",
			Priority:    "medium",
			Title:       "技能建议",
			Description: "建议开启「日程管理」和「邮件撰写」技能，提升办公效率",
		})
		
	case "expert":
		recs = append(recs, Recommendation{
			Type:        "best_practice",
			Priority:    "high",
			Title:       "专业资质",
			Description: "在提示词中明确说明专业背景和资质，增强可信度",
			Example:     "你是一位拥有 10 年经验的资深专家，持有 PMP、CBAP 等认证...",
		})
		
	case "creator":
		recs = append(recs, Recommendation{
			Type:        "suggestion",
			Priority:    "medium",
			Title:       "创意思维",
			Description: "建议设置较高的 temperature 参数（0.8-0.9），激发更多创意",
		})
		
	case "teacher":
		recs = append(recs, Recommendation{
			Type:        "best_practice",
			Priority:    "high",
			Title:       "教学方法",
			Description: "采用循序渐进的教学方式，先了解学生基础再调整难度",
			Example:     "在回答前先询问：「你目前对这个主题了解多少？」",
		})
		
	case "companion":
		recs = append(recs, Recommendation{
			Type:        "warning",
			Priority:    "high",
			Title:       "重要提醒",
			Description: "情感陪伴类角色需要明确边界，不能替代专业心理咨询",
			Example:     "添加免责声明：「我是一个 AI 伙伴，如需专业帮助请咨询心理医生」",
		})
	}
	
	return recs
}

func (g *PromptGenerator) getStyleRecommendations(data WizardData) []Recommendation {
	var recs []Recommendation
	
	switch data.Style {
	case "professional":
		recs = append(recs, Recommendation{
			Type:        "best_practice",
			Priority:    "medium",
			Title:       "格式化输出",
			Description: "使用结构化的回答格式，如分点、标题、总结等",
		})
		
	case "friendly":
		recs = append(recs, Recommendation{
			Type:        "suggestion",
			Priority:    "low",
			Title:       "增加亲和力",
			Description: "可以适当使用表情符号和温暖的语气词",
		})
		
	case "humorous":
		recs = append(recs, Recommendation{
			Type:        "warning",
			Priority:    "medium",
			Title:       "幽默边界",
			Description: "注意幽默的场合和对象，避免敏感话题",
		})
		
	case "concise":
		recs = append(recs, Recommendation{
			Type:        "best_practice",
			Priority:    "medium",
			Title:       "简洁原则",
			Description: "采用金字塔原理：结论先行，再展开说明",
		})
	}
	
	return recs
}

func (g *PromptGenerator) getExpertiseRecommendations(data WizardData) []Recommendation {
	var recs []Recommendation
	
	for _, expertiseID := range data.Expertise {
		switch expertiseID {
		case "legal":
			recs = append(recs, Recommendation{
				Type:        "warning",
				Priority:    "high",
				Title:       "法律免责声明",
				Description: "必须添加免责声明，说明不构成正式法律意见",
				Example:     "重要提示：我的回答仅供参考，不构成正式法律意见。具体案件请咨询执业律师。",
			})
			
		case "health":
			recs = append(recs, Recommendation{
				Type:        "warning",
				Priority:    "high",
				Title:       "医疗免责声明",
				Description: "必须添加医疗免责声明，建议用户咨询专业医师",
				Example:     "重要提示：我的建议不能替代专业医疗诊断。如有健康问题请咨询医生。",
			})
			
		case "finance":
			recs = append(recs, Recommendation{
				Type:        "warning",
				Priority:    "high",
				Title:       "投资风险提示",
				Description: "需要说明投资有风险，建议仅供参考",
				Example:     "投资有风险，入市需谨慎。我的建议仅供参考，不构成投资建议。",
			})
			
		case "tech":
			recs = append(recs, Recommendation{
				Type:        "suggestion",
				Priority:    "medium",
				Title:       "代码示例",
				Description: "提供代码示例时，建议包含注释和使用说明",
			})
		}
	}
	
	return recs
}

func (g *PromptGenerator) getAvoidanceRecommendations(data WizardData) []Recommendation {
	var recs []Recommendation
	
	for _, avoidanceID := range data.Avoidances {
		switch avoidanceID {
		case "speculation":
			recs = append(recs, Recommendation{
				Type:        "best_practice",
				Priority:    "medium",
				Title:       "不确定性表达",
				Description: "使用「可能」「一般来说」「据我所知」等限定词",
			})
			
		case "overpromise":
			recs = append(recs, Recommendation{
				Type:        "best_practice",
				Priority:    "high",
				Title:       "能力边界",
				Description: "明确说明 AI 的能力限制，管理用户预期",
				Example:     "「作为 AI，我无法...但我可以帮你...」",
			})
		}
	}
	
	return recs
}

// ============ 测试服务 ============

// RunTest 运行测试对话
func (g *PromptGenerator) RunTest(data WizardData, testMessage string) TestResult {
	// 生成测试响应（简化版，实际应调用 AI API）
	response := g.generateTestResponse(data, testMessage)
	
	// 评估响应质量
	score := g.evaluateResponse(data, testMessage, response)
	
	// 生成反馈
	feedback := g.generateFeedback(data, testMessage, response, score)
	
	// 生成优化建议
	suggestions := g.generateTestSuggestions(data, score)
	
	return TestResult{
		TestID:      fmt.Sprintf("test_%d", time.Now().Unix()),
		Input:       testMessage,
		Output:      response,
		Score:       score,
		Feedback:    feedback,
		Suggestions: suggestions,
		Timestamp:   time.Now(),
	}
}

func (g *PromptGenerator) generateTestResponse(data WizardData, message string) string {
	style := data.Style
	name := data.Name
	
	var response string
	
	// 根据风格生成不同的响应
	switch style {
	case "friendly":
		response = fmt.Sprintf("你好呀！我是%s，很高兴能帮助你！关于你说的「%s」，让我想想...", name, message)
	case "professional":
		response = fmt.Sprintf("您好，我是%s。针对您提出的问题「%s」，我将从专业角度为您分析。", name, message)
	case "humorous":
		response = fmt.Sprintf("哈喽！我是%s~ 哇，这个问题有意思！「%s」，让我来秀一波操作！", name, message)
	case "concise":
		response = fmt.Sprintf("我是%s。问题：「%s」。解答如下：", name, message)
	default:
		response = fmt.Sprintf("你好！我是%s。关于「%s」这个问题，我来帮你解答。", name, message)
	}
	
	return response
}

func (g *PromptGenerator) evaluateResponse(data WizardData, message, response string) float64 {
	score := 70.0 // 基础分
	
	// 响应长度评分
	if len(response) > 20 && len(response) < 500 {
		score += 10
	}
	
	// 包含用户消息关键词
	if strings.Contains(response, message) {
		score += 10
	}
	
	// 风格匹配
	style := data.Style
	if style == "friendly" && strings.Contains(response, "高兴") {
		score += 5
	}
	if style == "professional" && strings.Contains(response, "您") {
		score += 5
	}
	
	// 包含角色名
	if strings.Contains(response, data.Name) {
		score += 5
	}
	
	// 上限 100
	if score > 100 {
		score = 100
	}
	
	return score
}

func (g *PromptGenerator) generateFeedback(data WizardData, message, response string, score float64) string {
	if score >= 90 {
		return "回答质量优秀！很好地体现了角色特点和风格。"
	} else if score >= 75 {
		return "回答不错，符合角色定位。可以考虑进一步优化细节。"
	} else if score >= 60 {
		return "回答基本合格，但还有改进空间。建议调整提示词或配置。"
	}
	return "回答质量有待提升。建议重新审视角色定位和风格设置。"
}

func (g *PromptGenerator) generateTestSuggestions(data WizardData, score float64) []string {
	var suggestions []string
	
	if score < 75 {
		suggestions = append(suggestions, "尝试在提示词中添加更多具体的行为指导")
		suggestions = append(suggestions, "调整 temperature 参数可能会改善回答质量")
	}
	
	if data.Style == "professional" && score < 80 {
		suggestions = append(suggestions, "可以添加「使用专业术语但要解释清楚」的指导")
	}
	
	if len(data.Expertise) > 3 && score < 80 {
		suggestions = append(suggestions, "专业领域过多可能导致焦点分散，建议聚焦核心领域")
	}
	
	return suggestions
}

// ============ 辅助方法 ============

func (g *PromptGenerator) getPurposeLabel(id string) string {
	for _, p := range Purposes {
		if p.ID == id {
			return p.Label
		}
	}
	return "智能助理"
}

func (g *PromptGenerator) getPurposeDescription(id string) string {
	for _, p := range Purposes {
		if p.ID == id {
			return p.Description
		}
	}
	return "处理各种日常任务"
}

func (g *PromptGenerator) getStyleLabel(id string) string {
	for _, s := range Styles {
		if s.ID == id {
			return s.Label
		}
	}
	return "专业友好"
}

func (g *PromptGenerator) getStyleDescription(id string) string {
	for _, s := range Styles {
		if s.ID == id {
			return s.Description
		}
	}
	return "保持专业且友好的态度"
}

func (g *PromptGenerator) getExpertiseLabels(ids []string) []string {
	var labels []string
	for _, id := range ids {
		for _, e := range ExpertiseAreas {
			if e.ID == id {
				labels = append(labels, e.Label)
				break
			}
		}
	}
	return labels
}

func (g *PromptGenerator) getAvoidance(id string) *AvoidanceOption {
	for _, a := range Avoidances {
		if a.ID == id {
			return &a
		}
	}
	return nil
}

// ============ 导出配置 ============

// ExportRoleConfig 导出角色配置
func (g *PromptGenerator) ExportRoleConfig(data WizardData) map[string]interface{} {
	generated := g.GeneratePrompt(data)
	
	return map[string]interface{}{
		"name":           data.Name,
		"description":    fmt.Sprintf("%s - %s", g.getPurposeLabel(data.Purpose), g.getStyleLabel(data.Style)),
		"category":       g.getPrimaryExpertise(data.Expertise),
		"systemPrompt":   generated.SystemPrompt,
		"welcomeMessage": generated.WelcomeMessage,
		"modelConfig":    generated.ModelConfig,
		"metadata":       generated.Metadata,
	}
}

func (g *PromptGenerator) getPrimaryExpertise(ids []string) string {
	if len(ids) == 0 {
		return "通用"
	}
	for _, e := range ExpertiseAreas {
		if e.ID == ids[0] {
			return e.Label
		}
	}
	return "通用"
}
