package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JSON 通用JSON类型 (SQLite 兼容 - 存储为 TEXT)
type JSON string

// User 用户模型
type User struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	Email         string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash  string    `json:"-" gorm:"not null"`
	Name          string    `json:"name"`
	Avatar        string    `json:"avatar"`
	EmailVerified bool      `json:"emailVerified" gorm:"default:false"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
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

// Role AI角色
type Role struct {
	ID             string     `json:"id" gorm:"primaryKey"`
	WorkspaceID    string     `json:"workspaceId"`
	Name           string     `json:"name"`
	Avatar         string     `json:"avatar"`
	Description    string     `json:"description"`
	Category       string     `json:"category"`
	SystemPrompt   string     `json:"systemPrompt"`
	WelcomeMessage string     `json:"welcomeMessage"`
	ModelConfig    JSON       `json:"modelConfig" gorm:"type:text"`
	IsTemplate     bool       `json:"isTemplate" gorm:"default:false"`
	IsPublic       bool       `json:"isPublic" gorm:"default:false"`
	Skills         []Skill    `json:"skills" gorm:"many2many:role_skills;"`
	Documents      []Document `json:"documents" gorm:"many2many:role_documents;"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// Skill 技能
type Skill struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Config      JSON      `json:"config" gorm:"type:text"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Document 文档
type Document struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	WorkspaceID  string    `json:"workspaceId"`
	Name         string    `json:"name"`
	FileType     string    `json:"fileType"`
	FileSize     int64     `json:"fileSize"`
	FilePath     string    `json:"filePath"`
	Status       string    `json:"status" gorm:"default:'pending'"`
	ChunkCount   int       `json:"chunkCount"`
	ErrorMessage string    `json:"errorMessage"`
	Metadata     JSON      `json:"metadata" gorm:"type:text"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ChatSession 对话会话
type ChatSession struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	RoleID    string    `json:"roleId"`
	UserID    string    `json:"userId"`
	Title     string    `json:"title"`
	Mode      string    `json:"mode" gorm:"default:'quick'"`
	Role      *Role     `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Messages  []Message `json:"messages" gorm:"foreignKey:SessionID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Message 消息
type Message struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	SessionID  string    `json:"sessionId"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	Sources    JSON      `json:"sources" gorm:"type:text"`
	TokensUsed int       `json:"tokensUsed"`
	CreatedAt  time.Time `json:"createdAt"`
}

// TableName 指定表名
func (User) TableName() string        { return "users" }
func (Workspace) TableName() string   { return "workspaces" }
func (Role) TableName() string        { return "roles" }
func (Skill) TableName() string       { return "skills" }
func (Document) TableName() string    { return "documents" }
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