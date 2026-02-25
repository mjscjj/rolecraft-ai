package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"rolecraft-ai/internal/api/middleware"
	"rolecraft-ai/internal/models"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	db *gorm.DB
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// RegisterRequest 注册请求
// @Description 用户注册请求体
type RegisterRequest struct {
	// 用户邮箱
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// 用户密码 (至少 6 位)
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	// 用户名称
	Name string `json:"name" binding:"required" example:"张三"`
}

// LoginRequest 登录请求
// @Description 用户登录请求体
type LoginRequest struct {
	// 用户邮箱
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// 用户密码
	Password string `json:"password" binding:"required" example:"password123"`
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户并返回 JWT token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 201 {object} map[string]interface{} "注册成功，返回用户信息和 token"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 409 {object} map[string]string "邮箱已存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查邮箱是否已存在
	var existingUser models.User
	if result := h.db.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	// 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// 创建用户
	user := models.User{
		ID:           models.NewUUID(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
	}

	if result := h.db.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// 生成 JWT
	token, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"user":  user,
			"token": token,
		},
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用邮箱和密码登录，返回 JWT token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} "登录成功，返回用户信息和 token"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "邮箱或密码错误"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	var user models.User
	if result := h.db.Where("email = ?", req.Email).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// 生成 JWT
	token, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"user":  user,
			"token": token,
		},
	})
}

// RefreshRequest 刷新 Token 请求
// @Description Token 刷新请求体
type RefreshRequest struct {
	// 过期的 JWT token
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// Refresh 刷新令牌
// @Summary 刷新 JWT Token
// @Description 使用过期的 JWT token 换取新的 token (支持 Authorization header 或 JSON body)
// @Tags 认证
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer {token}"
// @Param request body RefreshRequest false "Token (如果 header 中没有提供)"
// @Success 200 {object} map[string]interface{} "刷新成功，返回新 token"
// @Failure 400 {object} map[string]string "缺少 token"
// @Failure 401 {object} map[string]string "Token 无效或用户不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var tokenString string

	// 从 Authorization header 获取 token
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			tokenString = parts[1]
		}
	}

	// 如果 header 中没有，尝试从 JSON body 获取
	if tokenString == "" {
		var req struct {
			Token string `json:"token"`
		}
		if err := c.ShouldBindJSON(&req); err == nil && req.Token != "" {
			tokenString = req.Token
		}
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token is required"})
		return
	}

	// 解析 token (即使过期也能提取 claims)
	claims := &middleware.Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JWTSecret, nil
	})

	// 允许 token 过期，但必须是有效的签名
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// 验证用户仍然存在
	var user models.User
	if result := h.db.Where("id = ?", claims.UserID).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// 生成新 token
	newToken, err := middleware.GenerateToken(claims.UserID, claims.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"token": newToken,
		},
	})
}
