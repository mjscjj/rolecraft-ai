package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

// DocumentHandler 文档处理器
type DocumentHandler struct {
	db          *gorm.DB
	uploadDir   string
	maxFileSize int64
}

// NewDocumentHandler 创建文档处理器
func NewDocumentHandler(db *gorm.DB) *DocumentHandler {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	os.MkdirAll(uploadDir, 0755)

	return &DocumentHandler{
		db:          db,
		uploadDir:   uploadDir,
		maxFileSize: 50 * 1024 * 1024,
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
	query := h.db.Where("workspace_id = ?", userId)

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

// Upload 上传文档
func (h *DocumentHandler) Upload(c *gin.Context) {
	userId, _ := c.Get("userId")

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

	fileId := models.NewUUID()
	fileName := fileId + ext
	filePath := filepath.Join(h.uploadDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	document := models.Document{
		ID:          fileId,
		WorkspaceID: userId.(string),
		Name:        header.Filename,
		FileType:    ext[1:],
		FilePath:    filePath,
		FileSize:    header.Size,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	if result := h.db.Create(&document); result.Error != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "success",
		"data":    document,
	})
}

// Get 获取文档详情
func (h *DocumentHandler) Get(c *gin.Context) {
	userId, _ := c.Get("userId")
	docId := c.Param("id")

	var document models.Document
	if result := h.db.Where("id = ? AND workspace_id = ?", docId, userId).First(&document); result.Error != nil {
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
	userId, _ := c.Get("userId")
	docId := c.Param("id")

	var document models.Document
	if result := h.db.Where("id = ? AND workspace_id = ?", docId, userId).First(&document); result.Error != nil {
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
	userId, _ := c.Get("userId")
	docId := c.Param("id")

	var document models.Document
	if result := h.db.Where("id = ? AND workspace_id = ?", docId, userId).First(&document); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	if document.FilePath != "" {
		os.Remove(document.FilePath)
	}

	if result := h.db.Delete(&document); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}
