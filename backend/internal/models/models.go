package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JSON 通用 JSON 类型 (SQLite 兼容 - 存储为 TEXT)
type JSON string

// User 用户模型
type User struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	Email           string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash    string    `json:"-" gorm:"not null"`
	Name            string    `json:"name"`
	Avatar          string    `json:"avatar"`
	AnythingLLMSlug string    `json:"anythingLLMSlug" gorm:"index"` // 新增：Workspace slug
	EmailVerified   bool      `json:"emailVerified" gorm:"default:false"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Workspace 工作空间
type Workspace struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Type        string    `json:"type" gorm:"default:'personal'"`
	OwnerID     string    `json:"ownerId"`
	Description string    `json:"description"`
	Logo        string    `json:"logo"`
	Settings    JSON      `json:"settings" gorm:"type:text"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Role AI 角色 - 简化模型
type Role struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	UserID         string    `json:"userId" gorm:"index"` // 关联用户（模板角色可为空）
	Name           string    `json:"name"`
	Avatar         string    `json:"avatar"`
	Description    string    `json:"description"`
	Category       string    `json:"category"`
	SystemPrompt   string    `json:"systemPrompt"`
	WelcomeMessage string    `json:"welcomeMessage"`
	ModelConfig    JSON      `json:"modelConfig" gorm:"type:text"`
	IsTemplate     bool      `json:"isTemplate" gorm:"default:false"`
	IsPublic       bool      `json:"isPublic" gorm:"default:false"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// Skill 技能
type Skill struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Config      JSON      `json:"config" gorm:"type:text"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Document 文档 - 添加 AnythingLLM 关联
type Document struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	UserID          string    `json:"userId" gorm:"index;not null"`
	Name            string    `json:"name"`
	FileType        string    `json:"fileType"`
	FileSize        int64     `json:"fileSize"`
	FilePath        string    `json:"filePath"` // 临时存储路径
	FolderID        string    `json:"folderId" gorm:"index"` // 文件夹 ID
	AnythingLLMHash string    `json:"anythingLLMHash" gorm:"index"` // 新增：AnythingLLM 文档 hash
	Status          string    `json:"status" gorm:"default:'pending'"` // pending/processing/completed/failed
	ChunkCount      int       `json:"chunkCount"`
	ErrorMessage    string    `json:"errorMessage"`
	Similarity      float64   `json:"similarity" gorm:"-"` // 搜索相似度 (不存储到数据库)
	Metadata        JSON      `json:"metadata" gorm:"type:text"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Folder 文件夹
type Folder struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"userId" gorm:"index;not null"`
	Name      string    `json:"name"`
	ParentID  string    `json:"parentId" gorm:"index"` // 父文件夹 ID，空表示根目录
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ChatSession 对话会话 - 添加关联
type ChatSession struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	UserID          string    `json:"userId" gorm:"index;not null"`
	RoleID          string    `json:"roleId" gorm:"index"`
	Title           string    `json:"title"`
	Mode            string    `json:"mode" gorm:"default:'quick'"` // quick/task
	AnythingLLMSlug string    `json:"anythingLLMSlug" gorm:"index"` // 新增：关联 Workspace
	ModelConfig     JSON      `json:"modelConfig" gorm:"type:text"` // 新增：存储元数据（归档状态等）
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Message 消息
type Message struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	SessionID  string    `json:"sessionId" gorm:"index"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	Likes      int       `json:"likes" gorm:"default:0"`
	Dislikes   int       `json:"dislikes" gorm:"default:0"`
	IsEdited   bool      `json:"isEdited" gorm:"default:false"`
	Sources    JSON      `json:"sources" gorm:"type:text"`
	TokensUsed int       `json:"tokensUsed"`
	UpdatedAt  time.Time `json:"updatedAt"`
	CreatedAt  time.Time `json:"createdAt"`
}

// TableName 指定表名
func (User) TableName() string        { return "users" }
func (Workspace) TableName() string   { return "workspaces" }
func (Role) TableName() string        { return "roles" }
func (Skill) TableName() string       { return "skills" }
func (Document) TableName() string    { return "documents" }
func (Folder) TableName() string      { return "folders" }
func (ChatSession) TableName() string { return "chat_sessions" }
func (Message) TableName() string     { return "messages" }

// NewUUID 生成新 UUID 字符串
func NewUUID() string {
	return uuid.New().String()
}

// ToJSON 转换为 JSON 字符串
func ToJSON(v interface{}) JSON {
	b, _ := json.Marshal(v)
	return JSON(b)
}

// FromJSON 从 JSON 字符串解析
func (j *JSON) FromJSON(v interface{}) error {
	return json.Unmarshal([]byte(*j), v)
}
