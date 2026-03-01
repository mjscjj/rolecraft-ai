package collab

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"rolecraft-ai/internal/config"
	"rolecraft-ai/internal/service/ai"
)

type AgentStep struct {
	Agent      string `json:"agent"`
	Purpose    string `json:"purpose"`
	Output     string `json:"output"`
	DurationMs int64  `json:"durationMs"`
}

type RunRequest struct {
	TaskName        string
	TaskDescription string
	TaskType        string
	InputSource     string
	ReportRule      string
	ExecutionMode   string
}

type RunResult struct {
	Summary     string      `json:"summary"`
	FinalAnswer string      `json:"finalAnswer"`
	Confidence  float64     `json:"confidence"`
	NextActions []string    `json:"nextActions"`
	Evidence    []string    `json:"evidence"`
	Steps       []AgentStep `json:"steps"`
}

type Orchestrator struct {
	openrouter *ai.OpenRouterClient
}

func NewOrchestrator(cfg *config.Config) *Orchestrator {
	var openrouter *ai.OpenRouterClient
	if strings.TrimSpace(cfg.OpenRouterKey) != "" {
		openrouter = ai.NewOpenRouterClient(ai.OpenRouterConfig{
			APIKey:  strings.TrimSpace(cfg.OpenRouterKey),
			BaseURL: strings.TrimSpace(cfg.OpenRouterURL),
			Model:   strings.TrimSpace(cfg.OpenRouterModel),
		})
	}
	return &Orchestrator{openrouter: openrouter}
}

func (o *Orchestrator) Run(ctx context.Context, req RunRequest) (*RunResult, error) {
	steps := make([]AgentStep, 0, 4)

	plannerInput := fmt.Sprintf(
		"任务名称：%s\n任务类型：%s\n任务描述：%s\n输入源：%s\n汇报规则：%s\n请给出执行计划、里程碑和验收标准。",
		req.TaskName, req.TaskType, req.TaskDescription, req.InputSource, req.ReportRule,
	)
	plannerOutput, plannerCost, err := o.ask(ctx, plannerSystemPrompt, plannerInput)
	if err != nil {
		return nil, err
	}
	plannerOutput = sanitizeText(plannerOutput)
	steps = append(steps, AgentStep{
		Agent:      "Planner",
		Purpose:    "任务拆解与执行计划",
		Output:     plannerOutput,
		DurationMs: plannerCost,
	})

	mode := strings.ToLower(strings.TrimSpace(req.ExecutionMode))
	if mode != "parallel" {
		mode = "serial"
	}

	var researcherOutput string
	var researcherCost int64
	var criticOutput string
	var criticCost int64
	if mode == "parallel" {
		researcherOutput, researcherCost, criticOutput, criticCost, err = o.runParallelPhase(ctx, plannerOutput)
	} else {
		researcherOutput, researcherCost, criticOutput, criticCost, err = o.runSerialPhase(ctx, plannerOutput)
	}
	if err != nil {
		return nil, err
	}
	steps = append(steps, AgentStep{
		Agent:      "Researcher",
		Purpose:    "信息补充与证据检索",
		Output:     sanitizeText(researcherOutput),
		DurationMs: researcherCost,
	})
	steps = append(steps, AgentStep{
		Agent:      "Critic",
		Purpose:    "质量审查与反例校验",
		Output:     sanitizeText(criticOutput),
		DurationMs: criticCost,
	})

	synthInput := fmt.Sprintf(
		"任务信息：\n%s\n\nPlanner:\n%s\n\nResearcher:\n%s\n\nCritic:\n%s\n\n请输出 JSON：{\"summary\":\"\",\"finalAnswer\":\"\",\"confidence\":0.0,\"nextActions\":[],\"evidence\":[]}",
		plannerInput, plannerOutput, researcherOutput, criticOutput,
	)
	synthOutput, synthCost, err := o.ask(ctx, synthesizerSystemPrompt, synthInput)
	if err != nil {
		return nil, err
	}
	synthOutput = sanitizeText(synthOutput)
	steps = append(steps, AgentStep{
		Agent:      "Synthesizer",
		Purpose:    "综合决议与结果产出",
		Output:     synthOutput,
		DurationMs: synthCost,
	})

	result := parseSynthResult(synthOutput)
	if strings.TrimSpace(result.FinalAnswer) == "" {
		result.FinalAnswer = synthOutput
	}
	if strings.TrimSpace(result.Summary) == "" {
		result.Summary = clipText(result.FinalAnswer, 180)
	}
	if result.Confidence <= 0 {
		result.Confidence = 0.72
	}
	result.NextActions = sanitizeList(result.NextActions)
	result.Evidence = sanitizeList(result.Evidence)
	result.Steps = steps
	return &result, nil
}

func (o *Orchestrator) runSerialPhase(ctx context.Context, plannerOutput string) (string, int64, string, int64, error) {
	researcherInput := fmt.Sprintf(
		"任务上下文：\n%s\n\n请输出关键信息、外部依赖、可验证证据（可给出链接占位）和风险提示。",
		plannerOutput,
	)
	researcherOutput, researcherCost, err := o.ask(ctx, researcherSystemPrompt, researcherInput)
	if err != nil {
		return "", 0, "", 0, err
	}
	researcherOutput = sanitizeText(researcherOutput)

	criticInput := fmt.Sprintf(
		"计划：\n%s\n\n研究结果：\n%s\n\n请指出漏洞、冲突、遗漏，并给出修正建议。",
		plannerOutput, researcherOutput,
	)
	criticOutput, criticCost, err := o.ask(ctx, criticSystemPrompt, criticInput)
	if err != nil {
		return "", 0, "", 0, err
	}
	return researcherOutput, researcherCost, sanitizeText(criticOutput), criticCost, nil
}

func (o *Orchestrator) runParallelPhase(ctx context.Context, plannerOutput string) (string, int64, string, int64, error) {
	type askResult struct {
		output string
		cost   int64
		err    error
	}
	var wg sync.WaitGroup
	wg.Add(2)

	researcherCh := make(chan askResult, 1)
	criticCh := make(chan askResult, 1)

	go func() {
		defer wg.Done()
		researcherInput := fmt.Sprintf(
			"任务上下文：\n%s\n\n请输出关键信息、外部依赖、可验证证据（可给出链接占位）和风险提示。",
			plannerOutput,
		)
		output, cost, err := o.ask(ctx, researcherSystemPrompt, researcherInput)
		researcherCh <- askResult{
			output: sanitizeText(output),
			cost:   cost,
			err:    err,
		}
	}()

	go func() {
		defer wg.Done()
		criticInput := fmt.Sprintf(
			"计划：\n%s\n\n请从反例和风险审查角度指出漏洞、冲突、遗漏，并给出修正建议。",
			plannerOutput,
		)
		output, cost, err := o.ask(ctx, criticSystemPrompt, criticInput)
		criticCh <- askResult{
			output: sanitizeText(output),
			cost:   cost,
			err:    err,
		}
	}()

	wg.Wait()
	close(researcherCh)
	close(criticCh)

	researcherRes := <-researcherCh
	criticRes := <-criticCh
	if researcherRes.err != nil {
		return "", 0, "", 0, researcherRes.err
	}
	if criticRes.err != nil {
		return "", 0, "", 0, criticRes.err
	}
	return researcherRes.output, researcherRes.cost, criticRes.output, criticRes.cost, nil
}

func (o *Orchestrator) ask(ctx context.Context, systemPrompt, userPrompt string) (string, int64, error) {
	start := time.Now()
	if o.openrouter == nil {
		mock := fmt.Sprintf("系统未配置大模型，使用降级输出。\n系统角色：%s\n用户输入：%s", systemPrompt, userPrompt)
		return mock, time.Since(start).Milliseconds(), nil
	}

	callCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	resp, err := o.openrouter.ChatCompletion(callCtx, []ai.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}, 0.2)
	if err != nil {
		fallback := fmt.Sprintf(
			"模型调用失败，已降级为本地协商摘要。\n系统角色：%s\n任务输入：%s\n建议：先拆解任务、补充证据、进行风险复核，再汇总输出。",
			clipText(systemPrompt, 48),
			clipText(userPrompt, 220),
		)
		return fallback, time.Since(start).Milliseconds(), nil
	}
	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("empty llm response")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), time.Since(start).Milliseconds(), nil
}

func parseSynthResult(raw string) RunResult {
	type payload struct {
		Summary     string   `json:"summary"`
		FinalAnswer string   `json:"finalAnswer"`
		Confidence  float64  `json:"confidence"`
		NextActions []string `json:"nextActions"`
		Evidence    []string `json:"evidence"`
	}
	text := sanitizeText(raw)
	data := payload{}

	if err := json.Unmarshal([]byte(text), &data); err != nil {
		if matched := extractJSON(text); matched != "" {
			_ = json.Unmarshal([]byte(matched), &data)
		}
	}

	return RunResult{
		Summary:     sanitizeText(data.Summary),
		FinalAnswer: sanitizeText(data.FinalAnswer),
		Confidence:  data.Confidence,
		NextActions: sanitizeList(data.NextActions),
		Evidence:    sanitizeList(data.Evidence),
	}
}

func extractJSON(raw string) string {
	re := regexp.MustCompile(`\{[\s\S]*\}`)
	matched := re.FindString(raw)
	return strings.TrimSpace(matched)
}

func clipText(input string, limit int) string {
	text := sanitizeText(input)
	if len(text) <= limit {
		return text
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}

func sanitizeText(input string) string {
	text := strings.TrimSpace(input)
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		if r == '\n' || r == '\t' {
			b.WriteRune(r)
			continue
		}
		if r < 0x20 || r == 0x7F {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func sanitizeList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		text := sanitizeText(item)
		if text != "" {
			out = append(out, text)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

const plannerSystemPrompt = `你是 Planner Agent。你的目标是把任务拆解为可执行计划，强调里程碑、时间和验收标准。`
const researcherSystemPrompt = `你是 Researcher Agent。你的目标是补充关键信息、证据与潜在依赖，输出可验证依据。`
const criticSystemPrompt = `你是 Critic Agent。你的目标是找漏洞、找风险、找冲突，并给出具体修正建议。`
const synthesizerSystemPrompt = `你是 Synthesizer Agent。请整合各 Agent 观点，输出唯一 JSON，不要包含代码块。`
