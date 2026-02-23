package ai

import (
	"context"
	"fmt"
	"strings"
)

// RAGService 检索增强生成服务
type RAGService struct {
	embeddingClient *EmbeddingClient
	vectorStore     VectorStore // 向量数据库接口
}

// VectorStore 向量数据库接口
type VectorStore interface {
	// 插入向量
	Insert(ctx context.Context, collection string, id string, vector []float32, metadata map[string]interface{}) error
	
	// 搜索相似向量
	Search(ctx context.Context, collection string, queryVector []float32, topK int) ([]SearchResult, error)
	
	// 删除向量
	Delete(ctx context.Context, collection string, id string) error
}

// SearchResult 搜索结果
type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// RAGConfig 配置
type RAGConfig struct {
	EmbeddingClient *EmbeddingClient
	VectorStore     VectorStore
}

// NewRAGService 创建 RAG 服务
func NewRAGService(config RAGConfig) *RAGService {
	return &RAGService{
		embeddingClient: config.EmbeddingClient,
		vectorStore:     config.VectorStore,
	}
}

// Retrieve 检索相关文档片段
func (s *RAGService) Retrieve(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	// 1. 向量化查询
	queryVector, err := s.embeddingClient.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// 2. 在向量数据库中搜索
	results, err := s.vectorStore.Search(ctx, "documents", queryVector, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return results, nil
}

// BuildPrompt 构建增强提示词
func (s *RAGService) BuildPrompt(systemPrompt string, query string, contexts []SearchResult) string {
	var contextBuilder strings.Builder
	
	if len(contexts) > 0 {
		contextBuilder.WriteString("\n\n参考信息：\n")
		for i, ctx := range contexts {
			if content, ok := ctx.Metadata["content"].(string); ok {
				contextBuilder.WriteString(fmt.Sprintf("\n[%d] %s", i+1, content))
			}
		}
		contextBuilder.WriteString("\n\n请基于以上参考信息回答用户问题。如果参考信息中没有相关内容，请根据你的知识回答。")
	}

	return fmt.Sprintf("%s%s\n\n用户问题：%s", systemPrompt, contextBuilder.String(), query)
}

// RAGChat RAG 对话流程
func (s *RAGService) RAGChat(ctx context.Context, systemPrompt string, query string, topK int) ([]SearchResult, string, error) {
	// 1. 检索相关文档
	contexts, err := s.Retrieve(ctx, query, topK)
	if err != nil {
		return nil, "", err
	}

	// 2. 构建增强提示词
	enhancedPrompt := s.BuildPrompt(systemPrompt, query, contexts)

	return contexts, enhancedPrompt, nil
}
