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
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

// AnythingLLMConfig AnythingLLM 配置
type AnythingLLMConfig struct {
	BaseURL   string
	APIKey    string
	Workspace string
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
		BaseURL:   normalizeAnythingLLMBaseURL(firstNonEmpty(os.Getenv("ANYTHINGLLM_BASE_URL"), os.Getenv("ANYTHINGLLM_URL"))),
		APIKey:    firstNonEmpty(os.Getenv("ANYTHINGLLM_API_KEY"), os.Getenv("ANYTHINGLLM_KEY")),
		Workspace: os.Getenv("ANYTHINGLLM_WORKSPACE"),
	}

	return &DocumentHandler{
		db:          db,
		uploadDir:   uploadDir,
		maxFileSize: 50 * 1024 * 1024,
		config:      config,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func normalizeAnythingLLMBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	if base == "" {
		return ""
	}
	base = strings.TrimSuffix(base, "/")
	if strings.HasSuffix(base, "/api/v1") {
		return base
	}
	return base + "/api/v1"
}

func (h *DocumentHandler) anythingLLMEnabled() bool {
	return strings.TrimSpace(h.config.BaseURL) != "" && strings.TrimSpace(h.config.APIKey) != ""
}

var allowedTypes = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".txt":  true,
	".md":   true,
}

// List 获取文档列表 (支持多条件过滤)
func (h *DocumentHandler) List(c *gin.Context) {
	userId, _ := c.Get("userId")

	var documents []models.Document
	query := h.db.Where("user_id = ?", userId)

	// 多条件过滤
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if fileType := c.Query("type"); fileType != "" {
		query = query.Where("file_type = ?", fileType)
	}

	if folderID := c.Query("folder"); folderID != "" {
		query = query.Where("folder_id = ?", folderID)
	}

	// 日期范围过滤
	if dateFrom := c.Query("dateFrom"); dateFrom != "" {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo := c.Query("dateTo"); dateTo != "" {
		query = query.Where("created_at <= ?", dateTo)
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

// Upload 上传文档 (支持多文件)
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

	// 获取文件夹 ID (可选)
	folderID := c.PostForm("folderId")

	// 支持多文件上传
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form"})
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		// 兼容单文件上传
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
			return
		}
		defer file.Close()
		files = []*multipart.FileHeader{header}
	}

	var uploadedDocs []models.Document

	for _, fileHeader := range files {
		doc, err := h.processSingleFile(fileHeader, userIdStr, folderID)
		if err != nil {
			continue // 跳过失败的文件
		}
		uploadedDocs = append(uploadedDocs, *doc)
	}

	if len(uploadedDocs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files uploaded successfully"})
		return
	}

	// 如果是单文件，返回单个对象；多文件返回数组
	if len(uploadedDocs) == 1 {
		c.JSON(http.StatusCreated, gin.H{
			"code":    200,
			"message": "document uploaded and processing",
			"data":    uploadedDocs[0],
		})
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"code":    200,
			"message": "documents uploaded and processing",
			"data":    uploadedDocs,
		})
	}
}

// processSingleFile 处理单个文件上传
func (h *DocumentHandler) processSingleFile(fileHeader *multipart.FileHeader, userIdStr, folderID string) (*models.Document, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if fileHeader.Size > h.maxFileSize {
		return nil, fmt.Errorf("file too large")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedTypes[ext] {
		return nil, fmt.Errorf("file type not allowed")
	}

	// 生成临时文件 ID
	fileId := models.NewUUID()
	fileName := fileId + ext
	tempFilePath := filepath.Join(h.uploadDir, "temp_"+fileName)

	// 临时保存到本地
	dst, err := os.Create(tempFilePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(tempFilePath)
		return nil, err
	}

	// 创建文档记录
	document := models.Document{
		ID:        fileId,
		UserID:    userIdStr,
		Name:      fileHeader.Filename,
		FileType:  ext[1:],
		FilePath:  tempFilePath,
		FileSize:  fileHeader.Size,
		Status:    "processing",
		FolderID:  folderID,
		CreatedAt: time.Now(),
	}

	if result := h.db.Create(&document); result.Error != nil {
		os.Remove(tempFilePath)
		return nil, result.Error
	}

	// 异步处理文档上传到 AnythingLLM
	go h.processDocumentAsync(document.ID, tempFilePath, userIdStr)

	return &document, nil
}

// processDocumentAsync 异步处理文档上传到 AnythingLLM
func (h *DocumentHandler) processDocumentAsync(docId, tempFilePath, userId string) {
	if !h.anythingLLMEnabled() {
		targetPath := filepath.Join(h.uploadDir, docId+filepath.Ext(tempFilePath))
		finalPath := tempFilePath
		if err := os.Rename(tempFilePath, targetPath); err == nil {
			finalPath = targetPath
		}
		h.updateDocumentStatusWithMetadata(docId, "completed", finalPath, map[string]interface{}{
			"processingMode": "local",
		})
		return
	}

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
	if !h.anythingLLMEnabled() {
		return "", "", fmt.Errorf("anythingllm is not configured")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

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

	if err := writer.WriteField("addToWorkspaces", userId); err != nil {
		return "", "", fmt.Errorf("failed to write workspace field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", h.config.BaseURL+"/document/upload", body)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.config.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	fileId, _ := result["filename"].(string)
	if fileId == "" {
		fileId = filepath.Base(filePath)
	}

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
	if !h.anythingLLMEnabled() {
		return fmt.Errorf("anythingllm is not configured")
	}

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
		"status":        status,
		"error_message": errorMessage,
		"updated_at":    time.Now(),
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
		var doc models.Document
		if err := h.db.Where("id = ?", docId).First(&doc).Error; err == nil {
			var existingMetadata map[string]interface{}
			if doc.Metadata != "" {
				json.Unmarshal([]byte(doc.Metadata), &existingMetadata)
			}
			if existingMetadata == nil {
				existingMetadata = make(map[string]interface{})
			}
			for k, v := range metadata {
				existingMetadata[k] = v
			}
			data, _ := json.Marshal(existingMetadata)
			updateData["metadata"] = models.JSON(data)
		}
	}

	h.db.Model(&models.Document{}).Where("id = ?", docId).Updates(updateData)
}

// Search 高级搜索 (支持多条件、高亮、排序)
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
		Query     string            `json:"query"`
		TopN      int               `json:"topN"`
		Filters   map[string]string `json:"filters"`
		SortBy    string            `json:"sortBy"`
		SortOrder string            `json:"sortOrder"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// 1. 向量搜索 (如果有查询)
	var vectorResults []interface{}
	if req.Query != "" {
		results, err := h.vectorSearch(req.Query, req.TopN, userIdStr)
		if err != nil {
			// 向量搜索失败，降级到文本搜索
			vectorResults = nil
		} else {
			vectorResults = results
		}
	}

	// 2. 数据库搜索
	var documents []models.Document
	query := h.db.Where("user_id = ?", userIdStr)

	// 应用过滤器
	if req.Filters != nil {
		if fileType, ok := req.Filters["type"]; ok && fileType != "" {
			query = query.Where("file_type = ?", fileType)
		}
		if status, ok := req.Filters["status"]; ok && status != "" {
			query = query.Where("status = ?", status)
		}
		if folderID, ok := req.Filters["folder"]; ok && folderID != "" {
			query = query.Where("folder_id = ?", folderID)
		}
	}

	// 文本搜索
	if req.Query != "" {
		query = query.Where("name LIKE ?", "%"+req.Query+"%")
	}

	// 排序
	switch req.SortBy {
	case "name":
		if req.SortOrder == "asc" {
			query = query.Order("name ASC")
		} else {
			query = query.Order("name DESC")
		}
	case "size":
		if req.SortOrder == "asc" {
			query = query.Order("file_size ASC")
		} else {
			query = query.Order("file_size DESC")
		}
	case "relevance":
		// 相关度排序在内存中处理
	default:
		query = query.Order("created_at DESC")
	}

	if result := query.Find(&documents); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 3. 计算相似度分数 (如果有向量搜索结果)
	if vectorResults != nil && len(vectorResults) > 0 {
		docMap := make(map[string]*models.Document)
		for i := range documents {
			docMap[documents[i].ID] = &documents[i]
		}

		// 为匹配的文档添加相似度分数
		for _, vr := range vectorResults {
			if vrMap, ok := vr.(map[string]interface{}); ok {
				if id, ok := vrMap["id"].(string); ok {
					if doc, exists := docMap[id]; exists {
						if score, ok := vrMap["score"].(float64); ok {
							doc.Similarity = score
						}
					}
				}
			}
		}

		// 按相关度排序
		if req.SortBy == "relevance" {
			sortByRelevance(documents)
		}
	}

	// 4. 高亮搜索结果
	highlightedDocs := h.highlightResults(documents, req.Query)

	elapsed := time.Since(startTime)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"query":         req.Query,
			"documents":     highlightedDocs,
			"total":         len(documents),
			"searchTimeMs":  elapsed.Milliseconds(),
			"vectorResults": len(vectorResults),
		},
	})
}

// highlightResults 高亮搜索结果
func (h *DocumentHandler) highlightResults(docs []models.Document, query string) []map[string]interface{} {
	result := make([]map[string]interface{}, len(docs))

	for i, doc := range docs {
		docMap := map[string]interface{}{
			"id":         doc.ID,
			"name":       doc.Name,
			"fileType":   doc.FileType,
			"fileSize":   doc.FileSize,
			"status":     doc.Status,
			"createdAt":  doc.CreatedAt,
			"updatedAt":  doc.UpdatedAt,
			"folderId":   doc.FolderID,
			"similarity": doc.Similarity,
		}

		// 添加高亮名称
		if query != "" {
			docMap["highlightedName"] = h.highlightText(doc.Name, query)
		}

		result[i] = docMap
	}

	return result
}

// highlightText 高亮文本中的关键词
func (h *DocumentHandler) highlightText(text, query string) string {
	if query == "" {
		return text
	}

	// 简单的高亮标记 (前端会渲染)
	return strings.ReplaceAll(text, query, "<mark>"+query+"</mark>")
}

// sortByRelevance 按相关度排序
func sortByRelevance(docs []models.Document) {
	for i := 0; i < len(docs)-1; i++ {
		for j := i + 1; j < len(docs); j++ {
			if docs[i].Similarity < docs[j].Similarity {
				docs[i], docs[j] = docs[j], docs[i]
			}
		}
	}
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

// Update 更新文档信息 (支持标签)
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
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		FolderID    string   `json:"folderId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateData := make(map[string]interface{})

	if req.Name != "" {
		updateData["name"] = req.Name
	}

	if req.FolderID != "" {
		updateData["folder_id"] = req.FolderID
	}

	// 更新元数据 (标签、描述等)
	if req.Description != "" || len(req.Tags) > 0 {
		var metadata map[string]interface{}
		if document.Metadata != "" {
			json.Unmarshal([]byte(document.Metadata), &metadata)
		}
		if metadata == nil {
			metadata = make(map[string]interface{})
		}

		if req.Description != "" {
			metadata["description"] = req.Description
		}
		if len(req.Tags) > 0 {
			metadata["tags"] = req.Tags
		}

		data, _ := json.Marshal(metadata)
		updateData["metadata"] = models.JSON(data)
	}

	updateData["updated_at"] = time.Now()

	if result := h.db.Model(&document).Updates(updateData); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 重新加载文档
	h.db.First(&document, docId)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    document,
	})
}

// GetStatus 获取文档处理状态
func (h *DocumentHandler) GetStatus(c *gin.Context) {
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
		"data": gin.H{
			"id":        document.ID,
			"status":    document.Status,
			"progress":  calculateProgress(document.Status),
			"message":   getStatusMessage(document),
			"updatedAt": document.UpdatedAt,
		},
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

	// 1. 从 AnythingLLM 删除文档
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

// BatchDelete 批量删除文档
func (h *DocumentHandler) BatchDelete(c *gin.Context) {
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
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证所有文档都属于该用户
	var documents []models.Document
	if result := h.db.Where("id IN ? AND user_id = ?", req.IDs, userIdStr).Find(&documents); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if len(documents) != len(req.IDs) {
		c.JSON(http.StatusForbidden, gin.H{"error": "some documents not found or unauthorized"})
		return
	}

	// 批量删除
	for _, doc := range documents {
		// 从 AnythingLLM 删除
		var metadata map[string]interface{}
		if doc.Metadata != "" {
			json.Unmarshal([]byte(doc.Metadata), &metadata)
		}

		if anythingLLMFileId, ok := metadata["anythingLLMFileId"].(string); ok {
			h.deleteFromAnythingLLM(anythingLLMFileId, userIdStr)
		}

		// 删除本地文件
		if doc.FilePath != "" {
			os.Remove(doc.FilePath)
		}

		// 删除数据库记录
		h.db.Delete(&doc)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"deleted": len(documents),
		},
	})
}

// BatchMove 批量移动文档到文件夹
func (h *DocumentHandler) BatchMove(c *gin.Context) {
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
		IDs      []string `json:"ids" binding:"required"`
		FolderID string   `json:"folderId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新文档的文件夹
	result := h.db.Model(&models.Document{}).
		Where("id IN ? AND user_id = ?", req.IDs, userIdStr).
		Update("folder_id", req.FolderID)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"moved": result.RowsAffected,
		},
	})
}

// BatchUpdateTags 批量更新文档标签
func (h *DocumentHandler) BatchUpdateTags(c *gin.Context) {
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
		IDs  []string `json:"ids" binding:"required"`
		Tags []string `json:"tags" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取所有文档
	var documents []models.Document
	if result := h.db.Where("id IN ? AND user_id = ?", req.IDs, userIdStr).Find(&documents); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 批量更新
	for _, doc := range documents {
		var metadata map[string]interface{}
		if doc.Metadata != "" {
			json.Unmarshal([]byte(doc.Metadata), &metadata)
		}
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		metadata["tags"] = req.Tags

		data, _ := json.Marshal(metadata)
		h.db.Model(&doc).Updates(map[string]interface{}{
			"metadata":   models.JSON(data),
			"updated_at": time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"updated": len(documents),
		},
	})
}

// Preview 预览文档内容
func (h *DocumentHandler) Preview(c *gin.Context) {
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

	// 根据文件类型返回不同内容
	switch document.FileType {
	case "pdf":
		// PDF 文件返回文件流
		if document.FilePath != "" {
			http.ServeFile(c.Writer, c.Request, document.FilePath)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		}
	case "txt", "md":
		// 文本文件返回内容
		if document.FilePath != "" {
			content, err := os.ReadFile(document.FilePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "success",
				"data": gin.H{
					"content": string(content),
				},
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "preview not supported for this file type"})
	}
}

// Download 下载文档
func (h *DocumentHandler) Download(c *gin.Context) {
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

	if document.FilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+document.Name+"\"")
	c.Header("Content-Type", "application/octet-stream")
	http.ServeFile(c.Writer, c.Request, document.FilePath)
}

// CreateFolder 创建文件夹
func (h *DocumentHandler) CreateFolder(c *gin.Context) {
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
		Name     string `json:"name" binding:"required"`
		ParentID string `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder := models.Folder{
		ID:       models.NewUUID(),
		UserID:   userIdStr,
		Name:     req.Name,
		ParentID: req.ParentID,
	}

	if result := h.db.Create(&folder); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "success",
		"data":    folder,
	})
}

// ListFolders 获取文件夹列表
func (h *DocumentHandler) ListFolders(c *gin.Context) {
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

	var folders []models.Folder
	if result := h.db.Where("user_id = ?", userIdStr).Order("created_at ASC").Find(&folders); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 计算每个文件夹的文档数量
	type FolderWithCount struct {
		models.Folder
		DocumentCount int64 `json:"documentCount"`
	}

	foldersWithCount := make([]FolderWithCount, len(folders))
	for i, folder := range folders {
		var count int64
		h.db.Model(&models.Document{}).Where("user_id = ? AND folder_id = ?", userIdStr, folder.ID).Count(&count)
		foldersWithCount[i] = FolderWithCount{
			Folder:        folder,
			DocumentCount: count,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    foldersWithCount,
	})
}

// DeleteFolder 删除文件夹
func (h *DocumentHandler) DeleteFolder(c *gin.Context) {
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

	folderId := c.Param("id")

	var folder models.Folder
	if result := h.db.Where("id = ? AND user_id = ?", folderId, userIdStr).First(&folder); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		return
	}

	// 删除文件夹 (不删除文档，文档 folder_id 设为空)
	h.db.Model(&models.Document{}).Where("folder_id = ?", folderId).Update("folder_id", nil)
	h.db.Delete(&folder)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// deleteFromAnythingLLM 从 AnythingLLM 删除文档
func (h *DocumentHandler) deleteFromAnythingLLM(filename, workspace string) error {
	if !h.anythingLLMEnabled() {
		return nil
	}

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

	return nil
}

// calculateProgress 根据状态计算进度百分比
func calculateProgress(status string) int {
	switch status {
	case "pending":
		return 0
	case "processing":
		return 50
	case "completed":
		return 100
	case "failed":
		return -1
	default:
		return 0
	}
}

// getStatusMessage 获取状态描述信息
func getStatusMessage(doc models.Document) string {
	switch doc.Status {
	case "pending":
		return "等待处理"
	case "processing":
		return "正在上传到知识库并建立索引..."
	case "completed":
		return "处理完成，可用于智能对话"
	case "failed":
		if doc.ErrorMessage != "" {
			return "处理失败：" + doc.ErrorMessage
		}
		return "处理失败"
	default:
		return "未知状态"
	}
}

// vectorSearch 调用 AnythingLLM 向量搜索 API
func (h *DocumentHandler) vectorSearch(query string, topN int, workspace string) ([]interface{}, error) {
	if !h.anythingLLMEnabled() {
		return nil, fmt.Errorf("anythingllm is not configured")
	}

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

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	responses, ok := result["responses"].([]interface{})
	if !ok {
		if items, ok := result["items"].([]interface{}); ok {
			return items, nil
		}
		return []interface{}{result}, nil
	}

	return responses, nil
}
