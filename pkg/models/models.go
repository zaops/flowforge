package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Username string `json:"username" gorm:"uniqueIndex;not null" binding:"required"`
	Email    string `json:"email" gorm:"uniqueIndex;not null" binding:"required,email"`
	Password string `json:"-" gorm:"not null"`
	Role     string `json:"role" gorm:"default:user"`
	Avatar   string `json:"avatar"`
	Status   string `json:"status" gorm:"default:active"`
	
	// 关联关系
	Projects []Project `json:"projects,omitempty" gorm:"foreignKey:UserID"`
}

// Project 项目模型
type Project struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name        string `json:"name" gorm:"not null" binding:"required"`
	Description string `json:"description"`
	RepoURL     string `json:"repo_url" gorm:"not null" binding:"required"`
	Branch      string `json:"branch" gorm:"default:main"`
	BuildPath   string `json:"build_path" gorm:"default:./"`
	DeployPath  string `json:"deploy_path"`
	Status      string `json:"status" gorm:"default:inactive"`
	
	// SSH配置
	SSHKeyID     *uint   `json:"ssh_key_id"`
	SSHKey       *SSHKey `json:"ssh_key,omitempty" gorm:"foreignKey:SSHKeyID"`
	
	// 用户关联
	UserID uint `json:"user_id" gorm:"not null"`
	User   User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	
	// 关联关系
	Deployments []Deployment `json:"deployments,omitempty" gorm:"foreignKey:ProjectID"`
	Pipelines   []Pipeline   `json:"pipelines,omitempty" gorm:"foreignKey:ProjectID"`
}

// SSHKey SSH密钥模型
type SSHKey struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name       string `json:"name" gorm:"not null" binding:"required"`
	PublicKey  string `json:"public_key" gorm:"type:text"`
	PrivateKey string `json:"-" gorm:"type:text"`
	Host       string `json:"host"`
	Port       int    `json:"port" gorm:"default:22"`
	Username   string `json:"username" gorm:"default:root"`
	Status     string `json:"status" gorm:"default:active"`
	
	// 用户关联
	UserID uint `json:"user_id" gorm:"not null"`
	User   User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Deployment 部署记录模型
type Deployment struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Version     string `json:"version"`
	CommitHash  string `json:"commit_hash"`
	Status      string `json:"status" gorm:"default:pending"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int64  `json:"duration"` // 部署耗时（秒）
	LogOutput   string `json:"log_output" gorm:"type:text"`
	ErrorMsg    string `json:"error_msg" gorm:"type:text"`
	
	// 项目关联
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	
	// 用户关联
	UserID uint `json:"user_id" gorm:"not null"`
	User   User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Pipeline 流水线模型
type Pipeline struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name        string `json:"name" gorm:"not null" binding:"required"`
	Description string `json:"description"`
	Config      string `json:"config" gorm:"type:text"` // YAML配置
	Status      string `json:"status" gorm:"default:active"`
	Trigger     string `json:"trigger" gorm:"default:manual"` // manual, webhook, schedule
	CronExpr    string `json:"cron_expr"` // 定时触发表达式
	
	// 项目关联
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	
	// 关联关系
	PipelineRuns []PipelineRun `json:"pipeline_runs,omitempty" gorm:"foreignKey:PipelineID"`
}

// PipelineRun 流水线执行记录
type PipelineRun struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	RunNumber   int        `json:"run_number"`
	Status      string     `json:"status" gorm:"default:pending"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int64      `json:"duration"` // 执行耗时（秒）
	LogOutput   string     `json:"log_output" gorm:"type:text"`
	ErrorMsg    string     `json:"error_msg" gorm:"type:text"`
	TriggerType string     `json:"trigger_type"` // manual, webhook, schedule
	
	// 流水线关联
	PipelineID uint     `json:"pipeline_id" gorm:"not null"`
	Pipeline   Pipeline `json:"pipeline,omitempty" gorm:"foreignKey:PipelineID"`
	
	// 用户关联
	UserID uint `json:"user_id" gorm:"not null"`
	User   User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	
	// 关联关系
	Steps []PipelineStep `json:"steps,omitempty" gorm:"foreignKey:PipelineRunID"`
}

// PipelineStep 流水线步骤
type PipelineStep struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name        string     `json:"name" gorm:"not null"`
	StepOrder   int        `json:"step_order"`
	Status      string     `json:"status" gorm:"default:pending"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int64      `json:"duration"` // 步骤耗时（秒）
	Command     string     `json:"command" gorm:"type:text"`
	LogOutput   string     `json:"log_output" gorm:"type:text"`
	ErrorMsg    string     `json:"error_msg" gorm:"type:text"`
	
	// 流水线执行关联
	PipelineRunID uint        `json:"pipeline_run_id" gorm:"not null"`
	PipelineRun   PipelineRun `json:"pipeline_run,omitempty" gorm:"foreignKey:PipelineRunID"`
}

// Environment 环境变量模型
type Environment struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Key         string `json:"key" gorm:"not null" binding:"required"`
	Value       string `json:"value" gorm:"type:text"`
	Description string `json:"description"`
	IsSecret    bool   `json:"is_secret" gorm:"default:false"`
	
	// 项目关联
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// Webhook Webhook模型
type Webhook struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name        string `json:"name" gorm:"not null" binding:"required"`
	URL         string `json:"url" gorm:"not null"`
	Secret      string `json:"secret"`
	Events      string `json:"events" gorm:"default:push"` // push, pull_request, etc.
	Status      string `json:"status" gorm:"default:active"`
	LastTrigger *time.Time `json:"last_trigger"`
	
	// 项目关联
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Key         string `json:"key" gorm:"uniqueIndex;not null"`
	Value       string `json:"value" gorm:"type:text"`
	Description string `json:"description"`
	Category    string `json:"category" gorm:"default:general"`
	IsPublic    bool   `json:"is_public" gorm:"default:false"`
}

// 常量定义
const (
	// 用户角色
	RoleAdmin = "admin"
	RoleUser  = "user"
	
	// 用户状态
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusBlocked  = "blocked"
	
	// 项目状态
	ProjectStatusActive   = "active"
	ProjectStatusInactive = "inactive"
	ProjectStatusArchived = "archived"
	
	// 部署状态
	DeployStatusPending    = "pending"
	DeployStatusRunning    = "running"
	DeployStatusSuccess    = "success"
	DeployStatusFailed     = "failed"
	DeployStatusCancelled  = "cancelled"
	
	// 流水线状态
	PipelineStatusActive   = "active"
	PipelineStatusInactive = "inactive"
	PipelineStatusArchived = "archived"
	
	// 流水线触发类型
	TriggerManual     = "manual"
	TriggerWebhook    = "webhook"
	TriggerSchedule   = "schedule"
	TriggerTypeManual = "manual" // 兼容性别名
	
	// 脚本类型
	ScriptTypeBash       = "bash"
	ScriptTypePowerShell = "powershell"
	ScriptTypePython     = "python"
	ScriptTypeShell      = "shell"
	
	// 流水线执行状态
	RunStatusPending   = "pending"
	RunStatusRunning   = "running"
	RunStatusSuccess   = "success"
	RunStatusFailed    = "failed"
	RunStatusCancelled = "cancelled"
	
	// 步骤状态
	StepStatusPending   = "pending"
	StepStatusRunning   = "running"
	StepStatusSuccess   = "success"
	StepStatusFailed    = "failed"
	StepStatusSkipped   = "skipped"
)

// 请求和响应结构体

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	RepoURL     string `json:"repo_url" binding:"required"`
	Branch      string `json:"branch"`
	BuildPath   string `json:"build_path"`
	DeployPath  string `json:"deploy_path"`
	SSHKeyID    *uint  `json:"ssh_key_id"`
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	RepoURL     *string `json:"repo_url"`
	Branch      *string `json:"branch"`
	BuildPath   *string `json:"build_path"`
	DeployPath  *string `json:"deploy_path"`
	SSHKeyID    *uint   `json:"ssh_key_id"`
	Status      *string `json:"status"`
}

// CreateSSHKeyRequest 创建SSH密钥请求
type CreateSSHKeyRequest struct {
	Name     string `json:"name" binding:"required"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
}

// CreatePipelineRequest 创建流水线请求
type CreatePipelineRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Config      string `json:"config" binding:"required"`
	Trigger     string `json:"trigger"`
	CronExpr    string `json:"cron_expr"`
	ProjectID   uint   `json:"project_id" binding:"required"`
}

// DeployRequest 部署请求
type DeployRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	Version   string `json:"version"`
	Branch    string `json:"branch"`
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Search   string `json:"search" form:"search"`
	Sort     string `json:"sort" form:"sort"`
	Order    string `json:"order" form:"order"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// 辅助方法

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Project) TableName() string {
	return "projects"
}

func (SSHKey) TableName() string {
	return "ssh_keys"
}

func (Deployment) TableName() string {
	return "deployments"
}

func (Pipeline) TableName() string {
	return "pipelines"
}

func (PipelineRun) TableName() string {
	return "pipeline_runs"
}

func (PipelineStep) TableName() string {
	return "pipeline_steps"
}

func (Environment) TableName() string {
	return "environments"
}

func (Webhook) TableName() string {
	return "webhooks"
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// IsValidRole 验证用户角色
func IsValidRole(role string) bool {
	return role == RoleAdmin || role == RoleUser
}

// IsValidStatus 验证用户状态
func IsValidStatus(status string) bool {
	return status == StatusActive || status == StatusInactive || status == StatusBlocked
}

// IsValidProjectStatus 验证项目状态
func IsValidProjectStatus(status string) bool {
	return status == ProjectStatusActive || status == ProjectStatusInactive || status == ProjectStatusArchived
}

// IsValidDeployStatus 验证部署状态
func IsValidDeployStatus(status string) bool {
	validStatuses := []string{
		DeployStatusPending, DeployStatusRunning, DeployStatusSuccess,
		DeployStatusFailed, DeployStatusCancelled,
	}
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

// IsValidTriggerType 验证触发类型
func IsValidTriggerType(trigger string) bool {
	return trigger == TriggerManual || trigger == TriggerWebhook || trigger == TriggerSchedule
}