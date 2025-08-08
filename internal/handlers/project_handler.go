package handlers

import (
	"net/http"
	"strconv"

	"flowforge/pkg/database"
	"flowforge/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	db *gorm.DB
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{
		db: database.DB,
	}
}

// List 获取项目列表
func (h *ProjectHandler) List(c *gin.Context) {
	var projects []models.Project
	result := h.db.Find(&projects)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目列表失败"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// Get 获取单个项目
func (h *ProjectHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	var project models.Project
	result := h.db.Preload("SSHKey").Preload("Pipelines").Preload("Schedules").First(&project, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	GitURL      string `json:"git_url" binding:"required"`
	GitBranch   string `json:"git_branch"`
	GitUsername string `json:"git_username"`
	GitPassword string `json:"git_password"`
	SSHKeyID    *uint  `json:"ssh_key_id"`
	WorkDir     string `json:"work_dir"`
}

// Create 创建项目
func (h *ProjectHandler) Create(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 如果提供了SSH密钥ID，检查它是否存在
	if req.SSHKeyID != nil {
		var sshKey models.SSHKey
		result := h.db.First(&sshKey, *req.SSHKeyID)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "SSH密钥不存在"})
			return
		}
	}

	// 获取用户ID
	userID, _ := c.Get("user_id")
	
	// 创建项目
	project := models.Project{
		Name:        req.Name,
		Description: req.Description,
		RepoURL:     req.GitURL,
		Branch:      req.GitBranch,
		BuildPath:   req.WorkDir,
		SSHKeyID:    req.SSHKeyID,
		UserID:      userID.(uint),
		Status:      models.ProjectStatusActive,
	}

	// 设置默认分支（如果未提供）
	if project.Branch == "" {
		project.Branch = "main"
	}

	if result := h.db.Create(&project); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建项目失败"})
		return
	}

	// 审计日志功能暂时移除，因为AuditLog模型不存在
	// TODO: 实现审计日志功能

	c.JSON(http.StatusCreated, gin.H{
		"message":    "项目创建成功",
		"project_id": project.ID,
	})
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	GitURL      string `json:"git_url"`
	GitBranch   string `json:"git_branch"`
	GitUsername string `json:"git_username"`
	GitPassword string `json:"git_password"`
	SSHKeyID    *uint  `json:"ssh_key_id"`
	WorkDir     string `json:"work_dir"`
}

// Update 更新项目
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 查找项目
	var project models.Project
	result := h.db.First(&project, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	// 如果提供了SSH密钥ID，检查它是否存在
	if req.SSHKeyID != nil {
		var sshKey models.SSHKey
		result := h.db.First(&sshKey, *req.SSHKeyID)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "SSH密钥不存在"})
			return
		}
	}

	// 更新字段
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.GitURL != "" {
		project.RepoURL = req.GitURL
	}
	if req.GitBranch != "" {
		project.Branch = req.GitBranch
	}
	if req.SSHKeyID != nil {
		project.SSHKeyID = req.SSHKeyID
	}
	if req.WorkDir != "" {
		project.BuildPath = req.WorkDir
	}

	// 保存更新
	if result := h.db.Save(&project); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新项目失败"})
		return
	}

	// 审计日志功能暂时移除，因为AuditLog模型不存在
	// TODO: 实现审计日志功能

	c.JSON(http.StatusOK, gin.H{
		"message": "项目更新成功",
	})
}

// Delete 删除项目
func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 查找项目
	var project models.Project
	result := h.db.First(&project, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	// 删除项目（软删除）
	if result := h.db.Delete(&project); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除项目失败"})
		return
	}

	// 审计日志功能暂时移除，因为AuditLog模型不存在
	// TODO: 实现审计日志功能

	c.JSON(http.StatusOK, gin.H{
		"message": "项目删除成功",
	})
}