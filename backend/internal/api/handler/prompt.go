package handler

import (
	"encoding/json"
	"net/http"

	"rolecraft-ai/internal/service/prompt"
)

// PromptHandler 提示词优化处理器
type PromptHandler struct {
	optimizer *prompt.Optimizer
}

// NewPromptHandler 创建提示词处理器
func NewPromptHandler(optimizer *prompt.Optimizer) *PromptHandler {
	return &PromptHandler{
		optimizer: optimizer,
	}
}

// OptimizeRequest API 请求结构
type OptimizeRequest struct {
	Prompt           string `json:"prompt"`
	GenerateVersions int    `json:"generateVersions"`
	IncludeSuggestions bool `json:"includeSuggestions"`
}

// APIResponse 通用 API 响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Optimize 优化提示词
// POST /api/prompt/optimize
func (h *PromptHandler) Optimize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OptimizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "请求参数无效", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		h.sendError(w, "提示词不能为空", http.StatusBadRequest)
		return
	}

	// 执行优化
	result, err := h.optimizer.Optimize(r.Context(), prompt.OptimizeRequest{
		Prompt:             req.Prompt,
		GenerateVersions:   req.GenerateVersions,
		IncludeSuggestions: req.IncludeSuggestions,
	})

	if err != nil {
		h.sendError(w, "优化失败："+err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, result)
}

// GetSuggestions 获取实时建议
// POST /api/prompt/suggestions
func (h *PromptHandler) GetSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Prompt string `json:"prompt"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "请求参数无效", http.StatusBadRequest)
		return
	}

	suggestions := h.optimizer.GenerateSuggestions(req.Prompt)
	h.sendSuccess(w, suggestions)
}

// LogSelection 记录用户选择（用于学习机制）
// POST /api/prompt/log
func (h *PromptHandler) LogSelection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OriginalPrompt  string `json:"originalPrompt"`
		SelectedVersion string `json:"selectedVersion"`
		UserID          string `json:"userID"`
		Rating          int    `json:"rating"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "请求参数无效", http.StatusBadRequest)
		return
	}

	// 记录优化历史
	if err := h.optimizer.LogOptimization(r.Context(), req.OriginalPrompt, req.SelectedVersion, req.UserID); err != nil {
		h.sendError(w, "记录失败", http.StatusInternalServerError)
		return
	}

	// 收集优质案例
	if req.Rating >= 4 {
		h.optimizer.CollectQualityCase(r.Context(), req.OriginalPrompt, req.SelectedVersion, req.Rating)
	}

	h.sendSuccess(w, nil)
}

// sendSuccess 发送成功响应
func (h *PromptHandler) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// sendError 发送错误响应
func (h *PromptHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Code:    status,
		Message: message,
	})
}
