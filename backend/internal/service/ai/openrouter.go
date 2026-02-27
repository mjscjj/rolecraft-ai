package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenRouterClient OpenRouter API 客户端
type OpenRouterClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

// OpenRouterConfig OpenRouter 配置
type OpenRouterConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// NewOpenRouterClient 创建 OpenRouter 客户端
func NewOpenRouterClient(config OpenRouterConfig) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:  config.APIKey,
		baseURL: config.BaseURL,
		model:   config.Model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// OpenRouterRequest OpenRouter 请求体
type OpenRouterRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// OpenRouterResponse OpenRouter 响应体
type OpenRouterResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		Delta        ChatMessage `json:"delta,omitempty"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// OpenRouterStreamChunk OpenRouter 流式数据块
type OpenRouterStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Delta        ChatMessage `json:"delta"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
}

// ChatCompletion 聊天完成
func (c *OpenRouterClient) ChatCompletion(ctx context.Context, messages []ChatMessage, temperature float64) (*ChatResponse, error) {
	reqBody := OpenRouterRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: temperature,
		Stream:      false,
		MaxTokens:   4096,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://rolecraft.ai")
	req.Header.Set("X-Title", "RoleCraft AI")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var openrouterResp OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openrouterResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openrouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openrouterResp.Choices[0]
	return &ChatResponse{
		ID:      openrouterResp.ID,
		Object:  openrouterResp.Object,
		Created: openrouterResp.Created,
		Model:   openrouterResp.Model,
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
					Content: choice.Message.Content,
				},
				FinishReason: choice.FinishReason,
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     openrouterResp.Usage.PromptTokens,
			CompletionTokens: openrouterResp.Usage.CompletionTokens,
			TotalTokens:      openrouterResp.Usage.TotalTokens,
		},
	}, nil
}

// ChatCompletionStream 流式聊天完成
func (c *OpenRouterClient) ChatCompletionStream(ctx context.Context, messages []ChatMessage, temperature float64) (<-chan *StreamChunk, <-chan error) {
	chunkChan := make(chan *StreamChunk, 100)
	errChan := make(chan error, 1)

	reqBody := OpenRouterRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: temperature,
		Stream:      true,
		MaxTokens:   4096,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		errChan <- fmt.Errorf("failed to marshal request: %w", err)
		close(chunkChan)
		close(errChan)
		return chunkChan, errChan
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		errChan <- fmt.Errorf("failed to create request: %w", err)
		close(chunkChan)
		close(errChan)
		return chunkChan, errChan
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://rolecraft.ai")
	req.Header.Set("X-Title", "RoleCraft AI")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		errChan <- fmt.Errorf("failed to send request: %w", err)
		close(chunkChan)
		close(errChan)
		return chunkChan, errChan
	}

	go func() {
		defer resp.Body.Close()
		defer close(chunkChan)
		defer close(errChan)

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
			return
		}

		decoder := json.NewDecoder(resp.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var chunk OpenRouterStreamChunk
				if err := decoder.Decode(&chunk); err != nil {
					if err == io.EOF {
						return
					}
					errChan <- fmt.Errorf("failed to decode chunk: %w", err)
					return
				}

				if len(chunk.Choices) > 0 {
					streamChunk := &StreamChunk{
						ID:      chunk.ID,
						Object:  chunk.Object,
						Created: chunk.Created,
						Model:   chunk.Model,
						Choices: []struct {
							Index        int         `json:"index"`
							Delta        ChatMessage `json:"delta"`
							FinishReason string      `json:"finish_reason"`
						}{
							{
								Index: 0,
								Delta: ChatMessage{
									Role:    "assistant",
									Content: chunk.Choices[0].Delta.Content,
								},
								FinishReason: chunk.Choices[0].FinishReason,
							},
						},
					}
					chunkChan <- streamChunk
				}
			}
		}
	}()

	return chunkChan, errChan
}

// SetModel 设置模型
func (c *OpenRouterClient) SetModel(model string) {
	c.model = model
}

// GetModel 获取当前模型
func (c *OpenRouterClient) GetModel() string {
	return c.model
}
