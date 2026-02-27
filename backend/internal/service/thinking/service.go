package thinking

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Service 思考服务
type Service struct {
	extractor *Extractor
}

// NewService 创建思考服务
func NewService() *Service {
	return &Service{
		extractor: NewExtractor(),
	}
}

// StreamThinkingSender 流式思考发送器
type StreamThinkingSender struct {
	thinkProcess *ThinkingProcess
	callback     func(chunk StreamChunk)
}

// NewStreamThinkingSender 创建流式思考发送器
func (s *Service) NewStreamThinkingSender(callback func(chunk StreamChunk)) *StreamThinkingSender {
	return &StreamThinkingSender{
		thinkProcess: NewThinkingProcess(),
		callback:     callback,
	}
}

// AddThinkingStep 添加思考步骤（流式）
func (sts *StreamThinkingSender) AddThinkingStep(stepType ThinkingStepType, content string) {
	step := sts.thinkProcess.AddStep(stepType, content)
	
	// 立即发送思考步骤
	chunk := StreamChunk{
		Type: "thinking",
		Data: step,
	}
	
	if sts.callback != nil {
		sts.callback(chunk)
	}
	
	// 模拟思考延迟（可选）
	time.Sleep(300 * time.Millisecond)
	
	// 标记步骤完成
	sts.thinkProcess.CompleteStep(step.ID)
	
	// 发送步骤完成更新
	chunk.Data = step
	if sts.callback != nil {
		sts.callback(chunk)
	}
}

// Complete 完成思考过程
func (sts *StreamThinkingSender) Complete() {
	sts.thinkProcess.Complete()
	
	chunk := StreamChunk{
		Type: "thinking_done",
		Data: sts.thinkProcess,
		Done: true,
	}
	
	if sts.callback != nil {
		sts.callback(chunk)
	}
}

// SendAnswer 发送最终答案
func (sts *StreamThinkingSender) SendAnswer(content string) {
	chunk := StreamChunk{
		Type: "answer",
		Data: map[string]string{"content": content},
	}
	
	if sts.callback != nil {
		sts.callback(chunk)
	}
}

// ProcessWithThinking 处理带思考过程的响应
func (s *Service) ProcessWithThinking(
	userMessage string,
	generateResponse func() (string, error),
) (*ThinkingProcess, string, error) {
	// 创建思考过程
	tp := NewThinkingProcess()
	
	// 步骤 1: 理解问题
	step1 := tp.AddStep(ThinkingUnderstand, "理解用户问题："+truncateString(userMessage, 50))
	time.Sleep(200 * time.Millisecond)
	tp.CompleteStep(step1.ID)
	
	// 步骤 2: 分析要素
	step2 := tp.AddStep(ThinkingAnalyze, "分析问题关键要素")
	time.Sleep(300 * time.Millisecond)
	tp.CompleteStep(step2.ID)
	
	// 步骤 3: 检索知识
	step3 := tp.AddStep(ThinkingSearch, "检索相关知识库")
	time.Sleep(400 * time.Millisecond)
	tp.CompleteStep(step3.ID)
	
	// 步骤 4: 组织答案
	step4 := tp.AddStep(ThinkingOrganize, "组织回答结构")
	time.Sleep(300 * time.Millisecond)
	tp.CompleteStep(step4.ID)
	
	// 生成最终响应
	response, err := generateResponse()
	if err != nil {
		return nil, "", err
	}
	
	// 步骤 5: 得出结论
	step5 := tp.AddStep(ThinkingConclude, "得出最终结论")
	time.Sleep(200 * time.Millisecond)
	tp.CompleteStep(step5.ID)
	
	// 完成思考过程
	tp.Complete()
	
	return tp, response, nil
}

// ExtractThinkingFromResponse 从响应中提取思考过程
func (s *Service) ExtractThinkingFromResponse(content string) *ExtractResult {
	return s.extractor.Extract(content)
}

// ThinkingToJSON 将思考过程转换为 JSON
func ThinkingToJSON(tp *ThinkingProcess) (string, error) {
	data, err := json.Marshal(tp)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ThinkingFromJSON 从 JSON 解析思考过程
func ThinkingFromJSON(jsonStr string) (*ThinkingProcess, error) {
	var tp ThinkingProcess
	if err := json.Unmarshal([]byte(jsonStr), &tp); err != nil {
		return nil, err
	}
	return &tp, nil
}

// FormatThinkingDuration 格式化思考时长
func FormatThinkingDuration(seconds float64) string {
	if seconds < 1.0 {
		return fmt.Sprintf("%.1fs", seconds)
	}
	return fmt.Sprintf("%.1fs", seconds)
}

// GetThinkingStepLabel 获取思考步骤的完整标签（图标 + 文字）
func GetThinkingStepLabel(stepType ThinkingStepType) string {
	icon := ThinkingStepIcon(stepType)
	label := ThinkingStepTypeLabel(stepType)
	return fmt.Sprintf("%s %s", icon, label)
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// CreateMockThinkingProcess 创建模拟思考过程（用于测试）
func CreateMockThinkingProcess(query string) *ThinkingProcess {
	tp := NewThinkingProcess()
	
	// 模拟 6 种思考步骤
	steps := []struct {
		stepType ThinkingStepType
		content  string
	}{
		{ThinkingUnderstand, "理解问题：" + truncateString(query, 40)},
		{ThinkingAnalyze, "分析关键要素：识别主要需求和约束条件"},
		{ThinkingSearch, "检索知识：从知识库中查找相关信息"},
		{ThinkingOrganize, "组织答案：按照逻辑结构整理内容"},
		{ThinkingInsight, "灵感闪现：想到一个更好的解决方案"},
		{ThinkingConclude, "得出结论：综合以上分析形成最终答案"},
	}
	
	for _, step := range steps {
		s := tp.AddStep(step.stepType, step.content)
		time.Sleep(100 * time.Millisecond)
		tp.CompleteStep(s.ID)
	}
	
	tp.Complete()
	return tp
}

// StreamChunkToJSON 将流式数据块转换为 JSON
func StreamChunkToJSON(chunk StreamChunk) (string, error) {
	data, err := json.Marshal(chunk)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CreateSSEData 创建 SSE 格式的数据
func CreateSSEData(chunk StreamChunk) (string, error) {
	jsonData, err := StreamChunkToJSON(chunk)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("data: %s\n\n", jsonData), nil
}

// ParseThinkingSteps 从文本中解析思考步骤（支持多种格式）
func ParseThinkingSteps(text string) []ThinkingStep {
	var steps []ThinkingStep
	
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 检测行首的标记
		var stepType ThinkingStepType
		var content string
		
		// 格式 1: 1. 内容
		// 格式 2: - 内容
		// 格式 3: * 内容
		// 格式 4: 首先/其次/最后
		
		if strings.HasPrefix(line, "首先") || strings.HasPrefix(line, "第一") {
			stepType = ThinkingUnderstand
			content = line
		} else if strings.HasPrefix(line, "其次") || strings.HasPrefix(line, "第二") {
			stepType = ThinkingAnalyze
			content = line
		} else if strings.HasPrefix(line, "然后") || strings.HasPrefix(line, "第三") {
			stepType = ThinkingSearch
			content = line
		} else if strings.HasPrefix(line, "接着") || strings.HasPrefix(line, "第四") {
			stepType = ThinkingOrganize
			content = line
		} else if strings.HasPrefix(line, "最后") || strings.HasPrefix(line, "总之") {
			stepType = ThinkingConclude
			content = line
		} else {
			// 默认根据行号判断
			switch i % 6 {
			case 0:
				stepType = ThinkingUnderstand
			case 1:
				stepType = ThinkingAnalyze
			case 2:
				stepType = ThinkingSearch
			case 3:
				stepType = ThinkingOrganize
			case 4:
				stepType = ThinkingInsight
			case 5:
				stepType = ThinkingConclude
			}
			content = line
		}
		
		if content != "" {
			step := NewThinkingStep(stepType, content)
			step.Status = ThinkingCompleted
			steps = append(steps, step)
		}
	}
	
	return steps
}
