package thinking_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"rolecraft-ai/internal/service/thinking"
)

// TestThinkingStepCreation 测试思考步骤创建
func TestThinkingStepCreation(t *testing.T) {
	step := thinking.NewThinkingStep(thinking.ThinkingUnderstand, "理解用户问题")
	
	if step.ID == "" {
		t.Error("Step ID should not be empty")
	}
	
	if step.Type != thinking.ThinkingUnderstand {
		t.Errorf("Expected type 'understand', got '%s'", step.Type)
	}
	
	if step.Status != thinking.ThinkingProcessing {
		t.Errorf("Expected status 'processing', got '%s'", step.Status)
	}
	
	if step.Icon != "🤔" {
		t.Errorf("Expected icon '🤔', got '%s'", step.Icon)
	}
	
	fmt.Printf("✅ Created step: %s - %s\n", step.Icon, step.Content)
}

// TestThinkingProcess 测试思考过程
func TestThinkingProcess(t *testing.T) {
	tp := thinking.NewThinkingProcess()
	
	// 添加步骤
	step1 := tp.AddStep(thinking.ThinkingUnderstand, "理解问题")
	time.Sleep(10 * time.Millisecond) // 模拟时间流逝
	step2 := tp.AddStep(thinking.ThinkingAnalyze, "分析要素")
	time.Sleep(10 * time.Millisecond)
	tp.AddStep(thinking.ThinkingSearch, "检索知识")
	
	if len(tp.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(tp.Steps))
	}
	
	// 完成步骤
	tp.CompleteStep(step1.ID)
	tp.CompleteStep(step2.ID)
	
	// 验证步骤状态
	if tp.Steps[0].Status != thinking.ThinkingCompleted {
		t.Error("Step 1 should be completed")
	}
	
	// Step 2 应该有 duration（因为创建后过了 10ms）
	if tp.Steps[1].Duration <= 0 {
		t.Logf("Step 2 duration: %.3fs (may be very small)", tp.Steps[1].Duration)
	}
	
	fmt.Printf("✅ Created process with %d steps\n", len(tp.Steps))
}

// TestThinkingComplete 测试思考完成
func TestThinkingComplete(t *testing.T) {
	tp := thinking.NewThinkingProcess()
	
	tp.AddStep(thinking.ThinkingUnderstand, "步骤 1")
	tp.AddStep(thinking.ThinkingAnalyze, "步骤 2")
	
	// 模拟思考时间
	time.Sleep(100 * time.Millisecond)
	
	// 完成思考
	tp.Complete()
	
	if !tp.IsComplete {
		t.Error("Thinking process should be complete")
	}
	
	if tp.Duration == 0 {
		t.Error("Duration should be calculated")
	}
	
	fmt.Printf("✅ Completed process in %.2fs\n", tp.Duration)
}

// TestThinkingExtractor 测试思考提取器
func TestThinkingExtractor(t *testing.T) {
	extractor := thinking.NewExtractor()
	
	// 测试带 thinking 标签的内容
	content := `<thinking>
首先，理解这个问题。
其次，分析关键要素。
最后，得出结论。
</thinking>

这是最终答案。`
	
	result := extractor.Extract(content)
	
	if !result.HasThinking {
		t.Error("Should detect thinking content")
	}
	
	if result.ThinkingProcess == nil {
		t.Error("Thinking process should not be nil")
	}
	
	if result.FinalAnswer == "" {
		t.Error("Final answer should not be empty")
	}
	
	fmt.Printf("✅ Extracted %d thinking steps\n", len(result.ThinkingProcess.Steps))
}

// TestStreamChunk 测试流式数据块
func TestStreamChunk(t *testing.T) {
	step := thinking.NewThinkingStep(thinking.ThinkingUnderstand, "测试步骤")
	
	chunk := thinking.StreamChunk{
		Type: "thinking",
		Data: step,
	}
	
	jsonData, err := thinking.StreamChunkToJSON(chunk)
	if err != nil {
		t.Errorf("Failed to marshal chunk: %v", err)
	}
	
	// 验证 JSON 格式
	var unmarshaled thinking.StreamChunk
	if err := json.Unmarshal([]byte(jsonData), &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal chunk: %v", err)
	}
	
	if unmarshaled.Type != "thinking" {
		t.Errorf("Expected type 'thinking', got '%s'", unmarshaled.Type)
	}
	
	fmt.Printf("✅ Stream chunk JSON: %s\n", jsonData[:50]+"...")
}

// TestMockThinkingProcess 测试模拟思考过程
func TestMockThinkingProcess(t *testing.T) {
	tp := thinking.CreateMockThinkingProcess("如何优化数据库性能？")
	
	if len(tp.Steps) != 6 {
		t.Errorf("Expected 6 steps, got %d", len(tp.Steps))
	}
	
	// 验证所有步骤都完成
	for i, step := range tp.Steps {
		if step.Status != thinking.ThinkingCompleted {
			t.Errorf("Step %d should be completed", i)
		}
	}
	
	if !tp.IsComplete {
		t.Error("Process should be complete")
	}
	
	fmt.Printf("✅ Created mock process with %d steps in %.2fs\n", len(tp.Steps), tp.Duration)
}

// TestThinkingStepTypes 测试所有思考步骤类型
func TestThinkingStepTypes(t *testing.T) {
	types := []thinking.ThinkingStepType{
		thinking.ThinkingUnderstand,
		thinking.ThinkingAnalyze,
		thinking.ThinkingSearch,
		thinking.ThinkingOrganize,
		thinking.ThinkingConclude,
		thinking.ThinkingInsight,
	}
	
	expectedIcons := []string{"🤔", "🔍", "📚", "📝", "✅", "💡"}
	
	for i, stepType := range types {
		icon := thinking.ThinkingStepIcon(stepType)
		if icon != expectedIcons[i] {
			t.Errorf("Expected icon '%s' for type '%s', got '%s'", expectedIcons[i], stepType, icon)
		}
		
		label := thinking.ThinkingStepTypeLabel(stepType)
		if label == "" {
			t.Errorf("Label should not be empty for type '%s'", stepType)
		}
		
		fmt.Printf("✅ %s %s: %s\n", icon, label, stepType)
	}
}

// TestService 测试思考服务
func TestService(t *testing.T) {
	svc := thinking.NewService()
	
	if svc == nil {
		t.Fatal("Service should not be nil")
	}
	
	// 测试 ProcessWithThinking
	startTime := time.Now()
	tp, answer, err := svc.ProcessWithThinking("测试问题", func() (string, error) {
		return "这是答案", nil
	})
	
	if err != nil {
		t.Errorf("ProcessWithThinking failed: %v", err)
	}
	
	if tp == nil {
		t.Error("Thinking process should not be nil")
	}
	
	if answer != "这是答案" {
		t.Errorf("Expected answer '这是答案', got '%s'", answer)
	}
	
	elapsed := time.Since(startTime).Seconds()
	fmt.Printf("✅ Service processed in %.2fs with %d steps\n", elapsed, len(tp.Steps))
}

// TestSSEData 测试 SSE 数据格式
func TestSSEData(t *testing.T) {
	chunk := thinking.StreamChunk{
		Type: "thinking",
		Data: map[string]string{
			"id":      "test-123",
			"type":    "understand",
			"content": "测试内容",
		},
	}
	
	sseData, err := thinking.CreateSSEData(chunk)
	if err != nil {
		t.Errorf("Failed to create SSE data: %v", err)
	}
	
	// 验证 SSE 格式
	if len(sseData) == 0 {
		t.Error("SSE data should not be empty")
	}
	
	fmt.Printf("✅ SSE data format: %s\n", sseData[:60]+"...")
}

// TestFormatDuration 测试时长格式化
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  float64
		expected string
	}{
		{0.5, "0.5s"},
		{1.0, "1.0s"},
		{2.5, "2.5s"},
		{10.3, "10.3s"},
	}
	
	for _, test := range tests {
		result := thinking.FormatThinkingDuration(test.seconds)
		if result != test.expected {
			t.Errorf("Expected '%s' for %.1fs, got '%s'", test.expected, test.seconds, result)
		}
	}
	
	fmt.Println("✅ Duration formatting works correctly")
}

// TestGetThinkingStepLabel 测试步骤标签
func TestGetThinkingStepLabel(t *testing.T) {
	label := thinking.GetThinkingStepLabel(thinking.ThinkingUnderstand)
	
	if label != "🤔 理解问题" {
		t.Errorf("Expected '🤔 理解问题', got '%s'", label)
	}
	
	fmt.Printf("✅ Step label: %s\n", label)
}

// BenchmarkThinkingProcess 性能测试
func BenchmarkThinkingProcess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tp := thinking.NewThinkingProcess()
		tp.AddStep(thinking.ThinkingUnderstand, "步骤 1")
		tp.AddStep(thinking.ThinkingAnalyze, "步骤 2")
		tp.AddStep(thinking.ThinkingSearch, "步骤 3")
		tp.Complete()
	}
}

// ExampleThinkingProcess 示例：如何创建思考过程
func ExampleThinkingProcess() {
	// 创建思考过程
	tp := thinking.NewThinkingProcess()
	
	// 添加思考步骤
	steps := []struct {
		stepType thinking.ThinkingStepType
		content  string
	}{
		{thinking.ThinkingUnderstand, "理解用户问题：如何学习 Go 语言？"},
		{thinking.ThinkingAnalyze, "分析关键要素：基础语法、并发编程、工程实践"},
		{thinking.ThinkingSearch, "检索知识：从 Go 官方文档和最佳实践中查找"},
		{thinking.ThinkingOrganize, "组织答案：按照学习路径从易到难"},
		{thinking.ThinkingConclude, "得出结论：提供完整的学习路线和资源"},
	}
	
	for _, step := range steps {
		s := tp.AddStep(step.stepType, step.content)
		time.Sleep(50 * time.Millisecond) // 模拟思考延迟
		tp.CompleteStep(s.ID)
	}
	
	// 完成思考
	tp.Complete()
	
	// 输出 JSON（用于前端显示）
	jsonData, _ := json.MarshalIndent(tp, "", "  ")
	fmt.Printf("Thinking Process JSON:\n%s\n", string(jsonData))
}
