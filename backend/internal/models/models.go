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
	CompanyID      string    `json:"companyId" gorm:"index"`
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
	CompanyID       string    `json:"companyId" gorm:"index"`
	WorkID          string    `json:"workId" gorm:"index"`
	Name            string    `json:"name"`
	FileType        string    `json:"fileType"`
	FileSize        int64     `json:"fileSize"`
	FilePath        string    `json:"filePath"`                        // 临时存储路径
	FolderID        string    `json:"folderId" gorm:"index"`           // 文件夹 ID
	AnythingLLMHash string    `json:"anythingLLMHash" gorm:"index"`    // 新增：AnythingLLM 文档 hash
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
	Mode            string    `json:"mode" gorm:"default:'quick'"`  // quick/task
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

// Company 公司（组织空间）
type Company struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	OwnerID     string    `json:"ownerId" gorm:"index;not null"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Work 工作区任务（异步执行单元）
type Work struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	UserID        string     `json:"userId" gorm:"index;not null"`
	CompanyID     string     `json:"companyId" gorm:"index"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Status        string     `json:"status" gorm:"default:'todo'"` // todo/in_progress/done
	Priority      string     `json:"priority" gorm:"default:'medium'"`
	RoleID        string     `json:"roleId" gorm:"index"`
	Type          string     `json:"type" gorm:"default:'general'"`           // general/report/analyze
	TriggerType   string     `json:"triggerType" gorm:"default:'manual'"`     // manual/once/daily/interval_hours
	TriggerValue  string     `json:"triggerValue"`                            // 例如 09:00 / 4 / 2026-03-01T09:00:00+08:00
	Timezone      string     `json:"timezone" gorm:"default:'Asia/Shanghai'"` // 时区
	NextRunAt     *time.Time `json:"nextRunAt"`                               // 下次执行时间
	LastRunAt     *time.Time `json:"lastRunAt"`                               // 最近执行时间
	AsyncStatus   string     `json:"asyncStatus" gorm:"default:'idle'"`       // idle/scheduled/running/completed/failed
	InputSource   string     `json:"inputSource"`                             // 输入源（如文档/文件夹）
	ReportRule    string     `json:"reportRule"`                              // 汇报规则
	ResultSummary string     `json:"resultSummary"`                           // 最近产出摘要
	Config        JSON       `json:"config" gorm:"type:text"`                 // 扩展配置
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// RoleInstall 角色安装记录（市场 -> 个人/公司）
type RoleInstall struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	TemplateID      string    `json:"templateId" gorm:"index;not null"`
	InstalledRoleID string    `json:"installedRoleId" gorm:"index;not null"`
	InstallerUserID string    `json:"installerUserId" gorm:"index;not null"`
	TargetType      string    `json:"targetType" gorm:"index;not null"` // personal/company
	TargetID        string    `json:"targetId" gorm:"index;not null"`   // userId/companyId
	CreatedAt       time.Time `json:"createdAt"`
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
func (Company) TableName() string     { return "companies" }
func (Work) TableName() string        { return "works" }
func (RoleInstall) TableName() string { return "role_installs" }

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
