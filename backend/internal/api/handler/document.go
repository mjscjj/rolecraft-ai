package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

// AnythingLLMConfig AnythingLLM 配置
type AnythingLLMConfig struct {
	BaseURL    string
	APIKey     string
	Workspace  string
}

// DocumentHandler 文档处理器
type DocumentHandler struct {
	db          *gorm.DB
	uploadDir   string
	maxFileSize int64
	config      AnythingLLMConfig
}

// NewDocumentHandler 创建文档处理器
func NewDocumentHandler(db *gorm.DB) *DocumentHandler {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	os.MkdirAll(uploadDir, 0755)

	// AnythingLLM 配置
	config := AnythingLLMConfig{
		BaseURL:   os.Getenv("ANYTHINGLLM_BASE_URL"),
		APIKey:    os.Getenv("ANYTHINGLLM_API_KEY"),
		Workspace: os.Getenv("ANYTHINGLLM_WORKSPACE"),
	}
	if config.BaseURL == "" {
		config.BaseURL = "http://150.109.21.115:3001/api/v1"
	}
	if config.Workspace == "" {
		config.Workspace = "user_001"
	}

	return &DocumentHandler{
		db:          db,
		uploadDir:   uploadDir,
		maxFileSize: 50 * 1024 * 1024,
		config:      config,
	}
}

var allowedTypes = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".txt":  true,
	".md":   true,
}

// List 获取文档列表
func (h *DocumentHandler) List(c *gin.Context) {
	userId, _ := c.Get("userId")

	var documents []models.Document
	query := h.db.Where("user_id = ?", userId)

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if fileType := c.Query("type"); fileType != "" {
		query = query.Where("file_type = ?", fileType)
	}

	if result := query.Order("created_at DESC").Find(&documents); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    documents,
	})
}

// Upload 上传文档 (异步处理)
func (h *DocumentHandler) Upload(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIdStr, ok := userId.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	if header.Size > h.maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 50MB)"})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file type not allowed"})
		return
	}

	// 生成临时文件 ID
	fileId := models.NewUUID()
	fileName := fileId + ext
	tempFilePath := filepath.Join(h.uploadDir, "temp_"+fileName)

	// 临时保存到本地
	dst, err := os.Create(tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// 创建文档记录 (状态为 processing)
	document := models.Document{
		ID:        fileId,
		UserID:    userIdStr,
		Name:      header.Filename,
		FileType:  ext[1:],
		FilePath:  tempFilePath,
		FileSize:  header.Size,
		Status:    "processing",
		CreatedAt: time.Now(),
	}

	if result := h.db.Create(&document); result.Error != nil {
		os.Remove(tempFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 异步处理文档上传到 AnythingLLM
	go h.processDocumentAsync(document.ID, tempFilePath, userIdStr)

	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "document uploaded and processing",
		"data":    document,
	})
}

// processDocumentAsync 异步处理文档上传到 AnythingLLM
func (h *DocumentHandler) processDocumentAsync(docId, tempFilePath, userId string) {
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Done()

	// 1. 上传到 AnythingLLM
	anythingLLMFileId, hash, err := h.uploadToAnythingLLM(tempFilePath, userId)
	if err != nil {
		h.updateDocumentStatus(docId, "failed", err.Error())
		os.Remove(tempFilePath)
		return
	}

	// 2. 等待处理完成并更新 embeddings
	err = h.updateEmbeddings(userId)
	if err != nil {
		h.updateDocumentStatus(docId, "failed", "embedding update failed: "+err.Error())
		os.Remove(tempFilePath)
		return
	}

	// 3. 更新文档状态为 completed
	finalFilePath := filepath.Join(h.uploadDir, docId+filepath.Ext(tempFilePath))
	err = os.Rename(tempFilePath, finalFilePath)
	if err != nil {
		// 如果重命名失败，至少更新状态
		h.updateDocumentStatusWithMetadata(docId, "completed", "", map[string]interface{}{
			"anythingLLMFileId": anythingLLMFileId,
			"anythingLLMHash":   hash,
		})
		return
	}

	h.updateDocumentStatusWithMetadata(docId, "completed", finalFilePath, map[string]interface{}{
		"anythingLLMFileId": anythingLLMFileId,
		"anythingLLMHash":   hash,
	})
}

// uploadToAnythingLLM 上传文档到 AnythingLLM
func (h *DocumentHandler) uploadToAnythingLLM(filePath, userId string) (string, string, error) {
	// 准备 multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", "", fmt.Errorf("failed to copy file: %w", err)
	}

	// 添加 workspace 参数
	if err := writer.WriteField("addToWorkspaces", userId); err != nil {
		return "", "", fmt.Errorf("failed to write workspace field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close writer: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", h.config.BaseURL+"/document/upload", body)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.config.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("anythingLLM upload failed: %s", string(respBody))
	}

	// 解析响应获取文件 ID
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	// 提取文件 ID (根据实际情况调整)
	fileId, _ := result["filename"].(string)
	if fileId == "" {
		fileId = filepath.Base(filePath)
	}

	// 计算文件 hash
	hash, err := h.calculateFileHash(filePath)
	if err != nil {
		hash = ""
	}

	return fileId, hash, nil
}

// calculateFileHash 计算文件 SHA256 hash
func (h *DocumentHandler) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// updateEmbeddings 更新 AnythingLLM 工作空间的 embeddings
func (h *DocumentHandler) updateEmbeddings(workspace string) error {
	reqBody := map[string]interface{}{}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/workspace/%s/update-embeddings", h.config.BaseURL, workspace),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update embeddings failed: %s", string(respBody))
	}

	return nil
}

// updateDocumentStatus 更新文档状态
func (h *DocumentHandler) updateDocumentStatus(docId, status, errorMessage string) {
	h.db.Model(&models.Document{}).Where("id = ?", docId).Updates(map[string]interface{}{
		"status":         status,
		"error_message":  errorMessage,
		"updated_at":     time.Now(),
	})
}

// updateDocumentStatusWithMetadata 更新文档状态和元数据
func (h *DocumentHandler) updateDocumentStatusWithMetadata(docId, status, filePath string, metadata map[string]interface{}) {
	updateData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if filePath != "" {
		updateData["file_path"] = filePath
	}

	if metadata != nil {
		// 获取现有元数据
		var doc models.Document
		if err := h.db.Where("id = ?", docId).First(&doc).Error; err == nil {
			var existingMetadata map[string]interface{}
			if doc.Metadata != "" {
				json.Unmarshal([]byte(doc.Metadata), &existingMetadata)
			}
			if existingMetadata == nil {
				existingMetadata = make(map[string]interface{})
			}
			// 合并新元数据
			for k, v := range metadata {
				existingMetadata[k] = v
			}
			data, _ := json.Marshal(existingMetadata)
			updateData["metadata"] = models.JSON(data)
		}
	}

	h.db.Model(&models.Document{}).Where("id = ?", docId).Updates(updateData)
}

// Search 向量搜索文档
func (h *DocumentHandler) Search(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIdStr, ok := userId.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Query string `json:"query" binding:"required"`
		TopN  int    `json:"topN" binding:"omitempty,min=1,max=20"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopN == 0 {
		req.TopN = 4
	}

	// 调用 AnythingLLM 向量搜索
	results, err := h.vectorSearch(req.Query, req.TopN, userIdStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"query":   req.Query,
			"results": results,
		},
	})
}

// vectorSearch 调用 AnythingLLM 向量搜索 API
func (h *DocumentHandler) vectorSearch(query string, topN int, workspace string) ([]interface{}, error) {
	reqBody := map[string]interface{}{
		"query": query,
		"topN":  topN,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/workspace/%s/vector-search", h.config.BaseURL, workspace),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vector search failed: %s", string(respBody))
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 提取结果 (根据实际情况调整)
	responses, ok := result["responses"].([]interface{})
	if !ok {
		// 尝试其他可能的字段
		if items, ok := result["items"].([]interface{}); ok {
			return items, nil
		}
		return []interface{}{result}, nil
	}

	return responses, nil
}

// Get 获取文档详情
func (h *DocumentHandler) Get(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIdStr, ok := userId.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	docId := c.Param("id")

	var document models.Document
	if result := h.db.Where("id = ? AND user_id = ?", docId, userIdStr).First(&document); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    document,
	})
}

// Update 更新文档信息
func (h *DocumentHandler) Update(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIdStr, ok := userId.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	docId := c.Param("id")

	var document models.Document
	if result := h.db.Where("id = ? AND user_id = ?", docId, userIdStr).First(&document); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		document.Name = req.Name
	}
	if req.Description != "" {
		if document.Metadata == "" {
			document.Metadata = `{"description":""}`
		}
		var metadata map[string]interface{}
		json.Unmarshal([]byte(document.Metadata), &metadata)
		metadata["description"] = req.Description
		data, _ := json.Marshal(metadata)
		document.Metadata = models.JSON(data)
	}

	document.UpdatedAt = time.Now()
	if result := h.db.Save(&document); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    document,
	})
}

// Delete 删除文档
func (h *DocumentHandler) Delete(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIdStr, ok := userId.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	docId := c.Param("id")

	var document models.Document
	if result := h.db.Where("id = ? AND user_id = ?", docId, userIdStr).First(&document); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	// 1. 从 AnythingLLM 删除文档 (如果有 anythingLLMFileId)
	var metadata map[string]interface{}
	if document.Metadata != "" {
		json.Unmarshal([]byte(document.Metadata), &metadata)
	}

	if anythingLLMFileId, ok := metadata["anythingLLMFileId"].(string); ok {
		h.deleteFromAnythingLLM(anythingLLMFileId, userIdStr)
	}

	// 2. 删除本地文件
	if document.FilePath != "" {
		os.Remove(document.FilePath)
	}

	// 3. 删除数据库记录
	if result := h.db.Delete(&document); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// deleteFromAnythingLLM 从 AnythingLLM 删除文档
func (h *DocumentHandler) deleteFromAnythingLLM(filename, workspace string) error {
	reqBody := map[string]interface{}{
		"filename": filename,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/workspace/%s/remove-document", h.config.BaseURL, workspace),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 不检查错误，即使删除失败也继续清理本地数据
	return nil
}
