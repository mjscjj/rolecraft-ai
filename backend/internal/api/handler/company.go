package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

type CompanyHandler struct {
	db *gorm.DB
}

func NewCompanyHandler(db *gorm.DB) *CompanyHandler {
	return &CompanyHandler{db: db}
}

type CompanyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *CompanyHandler) List(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	var companies []models.Company
	if err := h.db.Where("owner_id = ?", userIDStr).Order("created_at DESC").Find(&companies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": companies})
}

func (h *CompanyHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	var req CompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company := models.Company{
		ID:          models.NewUUID(),
		OwnerID:     userIDStr,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.db.Create(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 200, "message": "success", "data": company})
}

func (h *CompanyHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var company models.Company
	if err := h.db.Where("id = ? AND owner_id = ?", id, userIDStr).First(&company).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	var roleCount int64
	var workCount int64
	var outcomeCount int64
	var docCount int64
	h.db.Model(&models.Role{}).Where("company_id = ?", company.ID).Count(&roleCount)
	h.db.Model(&models.Work{}).Where("company_id = ?", company.ID).Count(&workCount)
	h.db.Model(&models.Work{}).Where("company_id = ? AND result_summary <> ''", company.ID).Count(&outcomeCount)
	h.db.Model(&models.Document{}).Where("company_id = ?", company.ID).Count(&docCount)

	var recentWorkspaces []models.Work
	h.db.
		Where("company_id = ? AND result_summary <> ''", company.ID).
		Order("updated_at DESC").
		Limit(10).
		Find(&recentWorkspaces)

	outcomes := make([]gin.H, 0, len(recentWorkspaces))
	for _, item := range recentWorkspaces {
		outcomes = append(outcomes, gin.H{
			"id":            item.ID,
			"name":          item.Name,
			"type":          item.Type,
			"asyncStatus":   item.AsyncStatus,
			"resultSummary": item.ResultSummary,
			"updatedAt":     item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"company": company,
			"stats": gin.H{
				"roleCount":      roleCount,
				"workCount":      workCount, // 兼容旧字段
				"workspaceCount": workCount,
				"outcomeCount":   outcomeCount,
				"docCount":       docCount,
			},
			"recentOutcomes": outcomes,
		},
	})
}

func (h *CompanyHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var req CompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	if err := h.db.Where("id = ? AND owner_id = ?", id, userIDStr).First(&company).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	company.Name = req.Name
	company.Description = req.Description
	company.UpdatedAt = time.Now()

	if err := h.db.Save(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": company})
}

func (h *CompanyHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	var company models.Company
	if err := h.db.Where("id = ? AND owner_id = ?", id, userIDStr).First(&company).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Role{}).Where("company_id = ?", id).Update("company_id", "").Error; err != nil {
			return err
		}
		if err := tx.Where("company_id = ?", id).Delete(&models.Work{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Document{}).Where("company_id = ?", id).Update("company_id", "").Error; err != nil {
			return err
		}
		return tx.Delete(&models.Company{}, "id = ? AND owner_id = ?", id, userIDStr).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}
