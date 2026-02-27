package thinking

import (
	"regexp"
	"strings"
	"time"
)

// ThinkingStepType 思考步骤类型
type ThinkingStepType string

const (
	// ThinkingUnderstand 理解问题
	ThinkingUnderstand ThinkingStepType = "understand"
	// ThinkingAnalyze 分析要素
	ThinkingAnalyze ThinkingStepType = "analyze"
	// ThinkingSearch 检索知识
	ThinkingSearch ThinkingStepType = "search"
	// ThinkingOrganize 组织答案
	ThinkingOrganize ThinkingStepType = "organize"
	// ThinkingConclude 得出结论
	ThinkingConclude ThinkingStepType = "conclude"
	// ThinkingInsight 灵感闪现
	ThinkingInsight ThinkingStepType = "insight"
)

// ThinkingStepStatus 思考步骤状态
type ThinkingStepStatus string

const (
	// ThinkingPending 等待中
	ThinkingPending ThinkingStepStatus = "pending"
	// ThinkingProcessing 处理中
	ThinkingProcessing ThinkingStepStatus = "processing"
	// ThinkingCompleted 已完成
	ThinkingCompleted ThinkingStepStatus = "completed"
)

// ThinkingStep 思考步骤
type ThinkingStep struct {
	ID        string           `json:"id"`
	Type      ThinkingStepType `json:"type"`
	Content   string           `json:"content"`
	Timestamp int64            `json:"timestamp"` // Unix timestamp in milliseconds
	Status    ThinkingStepStatus `json:"status"`
	Icon      string           `json:"icon"`
	Duration  float64          `json:"duration,omitempty"` // 步骤耗时（秒）
}

// ThinkingProcess 思考过程
type ThinkingProcess struct {
	Steps     []ThinkingStep `json:"steps"`
	StartTime int64          `json:"startTime"` // Unix timestamp in milliseconds
	EndTime   int64          `json:"endTime,omitempty"`
	Duration  float64        `json:"duration"` // 总耗时（秒）
	IsComplete bool          `json:"isComplete"`
}

// ThinkingStepIcon 获取思考步骤对应的图标
func ThinkingStepIcon(stepType ThinkingStepType) string {
	icons := map[ThinkingStepType]string{
		ThinkingUnderstand: "🤔",
		ThinkingAnalyze:    "🔍",
		ThinkingSearch:     "📚",
		ThinkingOrganize:   "📝",
		ThinkingConclude:   "✅",
		ThinkingInsight:    "💡",
	}
	
	if icon, ok := icons[stepType]; ok {
		return icon
	}
	return "💭"
}

// ThinkingStepTypeLabel 获取思考步骤的中文标签
func ThinkingStepTypeLabel(stepType ThinkingStepType) string {
	labels := map[ThinkingStepType]string{
		ThinkingUnderstand: "理解问题",
		ThinkingAnalyze:    "分析要素",
		ThinkingSearch:     "检索知识",
		ThinkingOrganize:   "组织答案",
		ThinkingConclude:   "得出结论",
		ThinkingInsight:    "灵感闪现",
	}
	
	if label, ok := labels[stepType]; ok {
		return label
	}
	return "思考中"
}

// NewThinkingStep 创建新的思考步骤
func NewThinkingStep(stepType ThinkingStepType, content string) ThinkingStep {
	return ThinkingStep{
		ID:        NewUUID(),
		Type:      stepType,
		Content:   content,
		Timestamp: time.Now().UnixMilli(),
		Status:    ThinkingProcessing,
		Icon:      ThinkingStepIcon(stepType),
	}
}

// NewThinkingProcess 创建新的思考过程
func NewThinkingProcess() *ThinkingProcess {
	return &ThinkingProcess{
		Steps:     make([]ThinkingStep, 0),
		StartTime: time.Now().UnixMilli(),
		IsComplete: false,
	}
}

// AddStep 添加思考步骤
func (tp *ThinkingProcess) AddStep(stepType ThinkingStepType, content string) ThinkingStep {
	step := NewThinkingStep(stepType, content)
	tp.Steps = append(tp.Steps, step)
	return step
}

// CompleteStep 完成思考步骤
func (tp *ThinkingProcess) CompleteStep(stepID string) {
	for i := range tp.Steps {
		if tp.Steps[i].ID == stepID {
			tp.Steps[i].Status = ThinkingCompleted
			// 计算步骤耗时
			stepEnd := time.Now().UnixMilli()
			tp.Steps[i].Duration = float64(stepEnd-tp.Steps[i].Timestamp) / 1000.0
			break
		}
	}
}

// Complete 完成思考过程
func (tp *ThinkingProcess) Complete() {
	tp.EndTime = time.Now().UnixMilli()
	tp.Duration = float64(tp.EndTime-tp.StartTime) / 1000.0
	tp.IsComplete = true
	
	// 确保所有步骤都标记为完成
	for i := range tp.Steps {
		if tp.Steps[i].Status == ThinkingProcessing {
			tp.Steps[i].Status = ThinkingCompleted
		}
	}
}

// Extractor 思考过程提取器
type Extractor struct {
	// 正则表达式用于提取思考标签内容
	thinkingTagRegex *regexp.Regexp
}

// NewExtractor 创建思考过程提取器
func NewExtractor() *Extractor {
	return &Extractor{
		thinkingTagRegex: regexp.MustCompile(`<thinking>(.*?)</thinking>`),
	}
}

// ExtractResult 提取结果
type ExtractResult struct {
	ThinkingProcess *ThinkingProcess
	FinalAnswer     string
	HasThinking     bool
}

// Extract 从内容中提取思考过程
func (e *Extractor) Extract(content string) *ExtractResult {
	result := &ExtractResult{
		FinalAnswer: content,
		HasThinking: false,
	}
	
	// 查找 thinking 标签
	matches := e.thinkingTagRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		// 没有找到 thinking 标签，尝试从内容中智能提取
		result.ThinkingProcess = e.extractSmart(content)
		if result.ThinkingProcess != nil && len(result.ThinkingProcess.Steps) > 0 {
			result.HasThinking = true
		}
		return result
	}
	
	// 提取到 thinking 内容
	thinkingContent := matches[1]
	result.HasThinking = true
	
	// 创建思考过程
	tp := NewThinkingProcess()
	
	// 解析思考步骤（按行分割）
	lines := strings.Split(thinkingContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 识别步骤类型
		stepType, stepContent := e.parseStepLine(line)
		if stepContent != "" {
			tp.AddStep(stepType, stepContent)
		}
	}
	
	// 如果没有解析出步骤，将整个 thinking 内容作为单个步骤
	if len(tp.Steps) == 0 && strings.TrimSpace(thinkingContent) != "" {
		tp.AddStep(ThinkingUnderstand, strings.TrimSpace(thinkingContent))
	}
	
	tp.Complete()
	result.ThinkingProcess = tp
	
	// 移除 thinking 标签，保留最终答案
	result.FinalAnswer = e.thinkingTagRegex.ReplaceAllString(content, "")
	result.FinalAnswer = strings.TrimSpace(result.FinalAnswer)
	
	return result
}

// parseStepLine 解析思考步骤行
func (e *Extractor) parseStepLine(line string) (ThinkingStepType, string) {
	// 尝试匹配各种格式
	// 格式 1: [类型] 内容
	// 格式 2: 类型：内容
	// 格式 3: emoji 内容
	
	// 检查是否包含类型标记
	if strings.Contains(line, "理解") || strings.Contains(line, "分析") {
		return ThinkingUnderstand, line
	}
	
	if strings.Contains(line, "分析") || strings.Contains(line, "要素") {
		return ThinkingAnalyze, line
	}
	
	if strings.Contains(line, "检索") || strings.Contains(line, "搜索") || strings.Contains(line, "知识") {
		return ThinkingSearch, line
	}
	
	if strings.Contains(line, "组织") || strings.Contains(line, "整理") {
		return ThinkingOrganize, line
	}
	
	if strings.Contains(line, "结论") || strings.Contains(line, "总结") {
		return ThinkingConclude, line
	}
	
	if strings.Contains(line, "灵感") || strings.Contains(line, "想法") {
		return ThinkingInsight, line
	}
	
	// 默认作为理解步骤
	return ThinkingUnderstand, line
}

// extractSmart 智能提取思考过程（当没有 thinking 标签时）
func (e *Extractor) extractSmart(content string) *ThinkingProcess {
	// 这是一个简化的实现，实际可以根据 AI 模型的输出格式进行优化
	// 例如，检测逻辑连接词、分段等
	
	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		return nil
	}
	
	tp := NewThinkingProcess()
	
	// 简单地将前几行作为思考步骤
	maxSteps := 3
	if len(lines) < maxSteps {
		maxSteps = len(lines)
	}
	
	for i := 0; i < maxSteps; i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			tp.AddStep(ThinkingUnderstand, line)
		}
	}
	
	if len(tp.Steps) > 0 {
		return tp
	}
	
	return nil
}

// StreamChunk 流式数据块
type StreamChunk struct {
	Type   string      `json:"type"` // "thinking" | "answer" | "done"
	Data   interface{} `json:"data"`
	Done   bool        `json:"done,omitempty"`
}

// NewUUID 生成 UUID（简化版本）
func NewUUID() string {
	// 使用 timestamp + random 作为简易 UUID
	return time.Now().Format("20060102150405.000")
}
