package prompt

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// PromptVersion 提示词版本
type PromptVersion struct {
	ID            string   `json:"id"`
	Content       string   `json:"content"`
	Score         int      `json:"score"`
	Features      []string `json:"features"`
	Scenarios     []string `json:"scenarios"`
	IsRecommended bool     `json:"isRecommended"`
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	Type       string `json:"type"` // specificity, example, tone, completeness
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

// OptimizationResult 优化结果
type OptimizationResult struct {
	Versions          []PromptVersion        `json:"versions"`
	Suggestions       []OptimizationSuggestion `json:"suggestions"`
	OriginalLength    int                    `json:"originalLength"`
	OptimizedLength   int                    `json:"optimizedLength"`
	ImprovementScore  int                    `json:"improvementScore"`
}

// OptimizeRequest 优化请求
type OptimizeRequest struct {
	Prompt             string `json:"prompt"`
	GenerateVersions   int    `json:"generateVersions"`
	IncludeSuggestions bool   `json:"includeSuggestions"`
}

// Optimizer 提示词优化器
type Optimizer struct {
}

// NewOptimizer 创建优化器实例
func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

// Optimize 执行提示词优化
func (o *Optimizer) Optimize(ctx context.Context, req OptimizeRequest) (*OptimizationResult, error) {
	if req.GenerateVersions <= 0 {
		req.GenerateVersions = 3
	}

	// 生成多个版本
	versions, err := o.generateVersions(ctx, req.Prompt, req.GenerateVersions)
	if err != nil {
		return nil, fmt.Errorf("生成版本失败：%w", err)
	}

	// 评估和评分
	o.evaluateVersions(versions, req.Prompt)

	// 生成建议
	var suggestions []OptimizationSuggestion
	if req.IncludeSuggestions {
		suggestions = o.GenerateSuggestions(req.Prompt)
	}

	// 计算改进分数
	originalLength := len(req.Prompt)
	optimizedLength := 0
	for _, v := range versions {
		optimizedLength += len(v.Content)
	}
	if len(versions) > 0 {
		optimizedLength = optimizedLength / len(versions)
	}

	improvementScore := 0
	if originalLength > 0 {
		improvementScore = ((optimizedLength - originalLength) * 100) / originalLength
		if improvementScore < 0 {
			improvementScore = 0
		}
	}

	return &OptimizationResult{
		Versions:         versions,
		Suggestions:      suggestions,
		OriginalLength:   originalLength,
		OptimizedLength:  optimizedLength,
		ImprovementScore: improvementScore,
	}, nil
}

// generateVersions 生成多个提示词版本
func (o *Optimizer) generateVersions(ctx context.Context, prompt string, count int) ([]PromptVersion, error) {
	versions := make([]PromptVersion, 0, count)

	// 版本 1: 结构化版本
	versions = append(versions, o.generateStructuredVersion(prompt))

	// 版本 2: 详细版本
	versions = append(versions, o.generateDetailedVersion(prompt))

	// 版本 3: 简洁版本
	versions = append(versions, o.generateConciseVersion(prompt))

	// 如果还需要更多版本，生成变体
	for len(versions) < count {
		versions = append(versions, o.generateVariantVersion(prompt, len(versions)+1))
	}

	return versions[:count], nil
}

// generateStructuredVersion 生成结构化版本
func (o *Optimizer) generateStructuredVersion(prompt string) PromptVersion {
	return PromptVersion{
		ID:      "1",
		Content: o.formatStructuredPrompt(prompt),
		Score:   85,
		Features: []string{
			"结构清晰",
			"逻辑完整",
			"易于理解",
		},
		Scenarios: []string{
			"复杂任务",
			"多步骤流程",
			"需要明确输出格式",
		},
		IsRecommended: true,
	}
}

// generateDetailedVersion 生成详细版本
func (o *Optimizer) generateDetailedVersion(prompt string) PromptVersion {
	return PromptVersion{
		ID:      "2",
		Content: o.formatDetailedPrompt(prompt),
		Score:   78,
		Features: []string{
			"细节丰富",
			"示例充足",
			"覆盖全面",
		},
		Scenarios: []string{
			"需要高精度结果",
			"复杂场景",
			"专业领域",
		},
	}
}

// generateConciseVersion 生成简洁版本
func (o *Optimizer) generateConciseVersion(prompt string) PromptVersion {
	return PromptVersion{
		ID:      "3",
		Content: o.formatConcisePrompt(prompt),
		Score:   72,
		Features: []string{
			"简洁明了",
			"快速执行",
			"重点突出",
		},
		Scenarios: []string{
			"简单任务",
			"快速原型",
			"日常使用",
		},
	}
}

// generateVariantVersion 生成变体版本
func (o *Optimizer) generateVariantVersion(prompt string, index int) PromptVersion {
	templates := []string{
		"请作为专业助手，%s。请确保回答准确、全面。",
		"我需要你帮助完成以下任务：%s。请提供详细的步骤和解释。",
		"任务描述：%s。请按照最佳实践来完成这个任务。",
	}

	template := templates[(index-1)%len(templates)]
	content := fmt.Sprintf(template, prompt)

	return PromptVersion{
		ID:      fmt.Sprintf("%d", index),
		Content: content,
		Score:   70 + rand.Intn(15),
		Features: []string{
			"平衡版本",
			"通用性强",
		},
		Scenarios: []string{
			"通用场景",
			"日常任务",
		},
	}
}

// formatStructuredPrompt 格式化结构化提示词
func (o *Optimizer) formatStructuredPrompt(prompt string) string {
	var sb strings.Builder

	sb.WriteString("## 角色设定\n")
	sb.WriteString("你是一位专业助手，擅长完成各种任务。\n\n")

	sb.WriteString("## 任务描述\n")
	sb.WriteString(prompt + "\n\n")

	sb.WriteString("## 输出要求\n")
	sb.WriteString("1. 回答准确、专业\n")
	sb.WriteString("2. 结构清晰、逻辑完整\n")
	sb.WriteString("3. 提供必要的示例和解释\n\n")

	sb.WriteString("## 注意事项\n")
	sb.WriteString("- 如有不确定之处，请明确说明\n")
	sb.WriteString("- 优先提供可执行的建议")

	return sb.String()
}

// formatDetailedPrompt 格式化详细提示词
func (o *Optimizer) formatDetailedPrompt(prompt string) string {
	var sb strings.Builder

	sb.WriteString("请作为该领域的专家助手，帮助我完成以下任务：\n\n")
	sb.WriteString("**任务背景**：\n")
	sb.WriteString(prompt + "\n\n")

	sb.WriteString("**详细要求**：\n")
	sb.WriteString("1. 请提供完整的解决方案，包括所有必要步骤\n")
	sb.WriteString("2. 对每个步骤进行详细解释，说明原因和方法\n")
	sb.WriteString("3. 提供至少 2-3 个具体示例\n")
	sb.WriteString("4. 指出可能的陷阱和注意事项\n")
	sb.WriteString("5. 如有替代方案，请一并说明\n\n")

	sb.WriteString("**输出格式**：\n")
	sb.WriteString("- 使用清晰的标题和分段\n")
	sb.WriteString("- 重要内容使用加粗标注\n")
	sb.WriteString("- 代码示例使用代码块格式\n")
	sb.WriteString("- 列表项使用项目符号")

	return sb.String()
}

// formatConcisePrompt 格式化简洁提示词
func (o *Optimizer) formatConcisePrompt(prompt string) string {
	return fmt.Sprintf("请帮我完成：%s。要求：简洁明了，直接给出核心内容，避免冗余。", prompt)
}

// evaluateVersions 评估版本质量
func (o *Optimizer) evaluateVersions(versions []PromptVersion, originalPrompt string) {
	originalLen := len(originalPrompt)

	for i := range versions {
		contentLen := len(versions[i].Content)

		// 基于长度、结构等因素调整评分
		lengthScore := 0
		if contentLen > originalLen && contentLen < originalLen*3 {
			lengthScore = 20
		}

		structureScore := 0
		if strings.Contains(versions[i].Content, "##") {
			structureScore = 15
		}

		clarityScore := 0
		if strings.Contains(versions[i].Content, "要求") || strings.Contains(versions[i].Content, "步骤") {
			clarityScore = 15
		}

		// 更新评分
		versions[i].Score += lengthScore + structureScore + clarityScore
		if versions[i].Score > 100 {
			versions[i].Score = 100
		}
	}

	// 重新确定推荐版本
	maxScore := 0
	recommendedIdx := 0
	for i, v := range versions {
		if v.Score > maxScore {
			maxScore = v.Score
			recommendedIdx = i
		}
	}

	for i := range versions {
		versions[i].IsRecommended = (i == recommendedIdx)
	}
}

// generateSuggestions 生成优化建议
func (o *Optimizer) GenerateSuggestions(prompt string) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	// 检查具体性
	if len(prompt) < 30 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:       "specificity",
			Message:    "描述可以更具体一些",
			Suggestion: "添加更多细节，如：具体场景、期望结果、限制条件等",
		})
	}

	// 检查是否有示例
	if !strings.Contains(prompt, "例如") && !strings.Contains(prompt, "比如") {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:       "example",
			Message:    "可以添加示例",
			Suggestion: "提供 1-2 个具体示例能帮助 AI 更好地理解你的需求",
		})
	}

	// 检查语气
	if !strings.Contains(prompt, "请") && !strings.Contains(prompt, "谢谢") {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:       "tone",
			Message:    "语气可以更友好",
			Suggestion: "使用礼貌用语（如'请'、'谢谢'）可以获得更好的交互体验",
		})
	}

	// 检查完整性
	words := strings.Fields(prompt)
	if len(words) < 10 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:       "completeness",
			Message:    "信息可能不够完整",
			Suggestion: "考虑补充：背景信息、目标受众、输出格式、时间限制等",
		})
	}

	return suggestions
}

// LogOptimization 记录优化历史（用于学习机制）
func (o *Optimizer) LogOptimization(ctx context.Context, originalPrompt, selectedVersion string, userID string) error {
	// TODO: 实现持久化存储
	// 记录用户选择，用于优化推荐算法
	fmt.Printf("[%s] 用户 %s 选择了版本：%s\n", time.Now().Format(time.RFC3339), userID, selectedVersion)
	return nil
}

// CollectQualityCase 收集优质案例
func (o *Optimizer) CollectQualityCase(ctx context.Context, prompt, optimizedPrompt string, rating int) error {
	// TODO: 实现优质案例收集
	// 用于持续改进生成质量
	fmt.Printf("收集优质案例：评分 %d\n", rating)
	return nil
}
