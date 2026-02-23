package document

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Processor 文档处理器
type Processor struct {
	maxChunkSize    int // 最大块大小（字符）
	chunkOverlap    int // 块重叠大小
	minChunkSize    int // 最小块大小
}

// ProcessorConfig 配置
type ProcessorConfig struct {
	MaxChunkSize int
	ChunkOverlap int
	MinChunkSize int
}

// Chunk 文本块
type Chunk struct {
	ID      string            `json:"id"`
	Content string            `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewProcessor 创建处理器
func NewProcessor(config ProcessorConfig) *Processor {
	maxChunkSize := config.MaxChunkSize
	if maxChunkSize == 0 {
		maxChunkSize = 1000 // 默认 1000 字符
	}

	chunkOverlap := config.ChunkOverlap
	if chunkOverlap == 0 {
		chunkOverlap = 200 // 默认重叠 200 字符
	}

	minChunkSize := config.MinChunkSize
	if minChunkSize == 0 {
		minChunkSize = 100 // 默认最小 100 字符
	}

	return &Processor{
		maxChunkSize: maxChunkSize,
		chunkOverlap: chunkOverlap,
		minChunkSize: minChunkSize,
	}
}

// ExtractText 提取文本（支持 PDF, DOCX, TXT）
// 实际项目中应使用专门的库：unioffice, pdfcpu 等
func (p *Processor) ExtractText(fileType string, data []byte) (string, error) {
	switch strings.ToLower(fileType) {
	case "txt", "md":
		return string(data), nil
		
	case "pdf":
		// TODO: 使用 pdfcpu 或 unidoc 解析 PDF
		// 这里返回占位文本
		return "PDF content extraction not implemented. Use pdfcpu or unidoc library.", nil
		
	case "doc", "docx":
		// TODO: 使用 unioffice 解析 Word 文档
		return "DOCX content extraction not implemented. Use unioffice library.", nil
		
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileType)
	}
}

// ChunkText 智能文本分块
func (p *Processor) ChunkText(text string) []Chunk {
	// 清理文本
	text = p.cleanText(text)
	
	// 按段落分割
	paragraphs := p.splitParagraphs(text)
	
	var chunks []Chunk
	var currentChunk strings.Builder
	currentSize := 0
	chunkID := 0
	
	for _, para := range paragraphs {
		paraSize := len(para)
		
		// 如果当前块加入段落后超过最大大小
		if currentSize+paraSize > p.maxChunkSize && currentSize > 0 {
			// 保存当前块
			if currentSize >= p.minChunkSize {
				chunks = append(chunks, Chunk{
					ID:       fmt.Sprintf("chunk-%d", chunkID),
					Content:  strings.TrimSpace(currentChunk.String()),
					Metadata: map[string]interface{}{"index": chunkID},
				})
				chunkID++
			}
			
			// 开始新块，保留重叠部分
			currentChunk.Reset()
			currentSize = 0
		}
		
		// 添加段落到当前块
		if currentSize > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(para)
		currentSize += paraSize
	}
	
	// 保存最后一个块
	if currentSize >= p.minChunkSize {
		chunks = append(chunks, Chunk{
			ID:       fmt.Sprintf("chunk-%d", chunkID),
			Content:  strings.TrimSpace(currentChunk.String()),
			Metadata: map[string]interface{}{"index": chunkID},
		})
	}
	
	return chunks
}

// ChunkTextWithOverlap 带重叠的分块
func (p *Processor) ChunkTextWithOverlap(text string) []Chunk {
	text = p.cleanText(text)
	
	var chunks []Chunk
	runes := []rune(text)
	totalLen := len(runes)
	
	for start := 0; start < totalLen; {
		end := start + p.maxChunkSize
		if end > totalLen {
			end = totalLen
		}
		
		// 尝试在句子边界处分割
		if end < totalLen {
			// 查找最近的句子结束符
			for i := end; i > start+p.minChunkSize; i-- {
				if runes[i-1] == '。' || runes[i-1] == '！' || runes[i-1] == '？' ||
					runes[i-1] == '.' || runes[i-1] == '!' || runes[i-1] == '?' {
					end = i
					break
				}
			}
		}
		
		chunk := string(runes[start:end])
		chunks = append(chunks, Chunk{
			ID:       fmt.Sprintf("chunk-%d", len(chunks)),
			Content:  strings.TrimSpace(chunk),
			Metadata: map[string]interface{}{"start": start, "end": end},
		})
		
		// 下一个块的起始位置（考虑重叠）
		start = end - p.chunkOverlap
		if start < 0 {
			start = 0
		}
		if start >= totalLen {
			break
		}
	}
	
	return chunks
}

// ProcessDocument 完整处理流程
func (p *Processor) ProcessDocument(fileType string, data []byte) ([]Chunk, error) {
	// 1. 提取文本
	text, err := p.ExtractText(fileType, data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}
	
	// 2. 分块
	chunks := p.ChunkTextWithOverlap(text)
	
	return chunks, nil
}

// cleanText 清理文本
func (p *Processor) cleanText(text string) string {
	// 移除多余的空白
	spaceRegex := regexp.MustCompile(`[ \t]+`)
	text = spaceRegex.ReplaceAllString(text, " ")
	
	// 移除多余的换行
	newlineRegex := regexp.MustCompile(`\n{3,}`)
	text = newlineRegex.ReplaceAllString(text, "\n\n")
	
	return strings.TrimSpace(text)
}

// splitParagraphs 按段落分割
func (p *Processor) splitParagraphs(text string) []string {
	var paragraphs []string
	lines := strings.Split(text, "\n")
	
	var currentPara strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentPara.Len() > 0 {
				paragraphs = append(paragraphs, currentPara.String())
				currentPara.Reset()
			}
		} else {
			if currentPara.Len() > 0 {
				currentPara.WriteString(" ")
			}
			currentPara.WriteString(line)
		}
	}
	
	if currentPara.Len() > 0 {
		paragraphs = append(paragraphs, currentPara.String())
	}
	
	return paragraphs
}

// ExtractTextFromReader 从 Reader 提取文本
func (p *Processor) ExtractTextFromReader(fileType string, reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read: %w", err)
	}
	return p.ExtractText(fileType, data)
}

// ProcessDocumentFromReader 从 Reader 处理文档
func (p *Processor) ProcessDocumentFromReader(fileType string, reader io.Reader) ([]Chunk, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}
	return p.ProcessDocument(fileType, data)
}

// CountTokens 估算 Token 数量（简化版）
func (p *Processor) CountTokens(text string) int {
	// 简化估算：英文约 4 字符 = 1 token，中文约 1.5 字符 = 1 token
	// 实际应使用 tiktoken 库
	runes := []rune(text)
	englishCount := 0
	chineseCount := 0
	
	for _, r := range runes {
		if r < 128 {
			englishCount++
		} else {
			chineseCount++
		}
	}
	
	tokens := englishCount/4 + chineseCount/2
	if tokens == 0 {
		tokens = 1
	}
	
	return tokens
}
