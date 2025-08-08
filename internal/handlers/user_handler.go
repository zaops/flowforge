package handlers

import (
	"net/http"
	"strconv"

	"flowforge/pkg/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserHandler 用户处理器
type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler 创建用户处理器
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

// List 获取用户列表
func (h *UserHandler) List(c *gin.Context) {
	var users []models.User
	result := h.db.Preload("Role").Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Get 获取单个用户
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var user models.User
	result := h.db.Preload("Role").First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name"`
	RoleID   uint   `json:"role_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 检查用户名是否已存在
	var count int64
	h.db.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// 检查邮箱是否已存在
	h.db.Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已存在"})
		return
	}

	// 检查角色是否存在
	var role models.Role
	result := h.db.First(&role, req.RoleID)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色不存在"})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		FullName: req.FullName,
		RoleID:   req.RoleID,
		IsActive: req.IsActive,
	}

	if result := h.db.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 记录审计日志
	auditLog := models.AuditLog{
		UserID:      &user.ID,
		Action:      "create_user",
		Description: "创建用户",
		IP:          c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
	}
	h.db.Create(&auditLog)

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
		"user_id": user.ID,
	})
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	FullName string `json:"full_name"`
	RoleID   *uint  `json:"role_id"`
	IsActive *bool  `json:"is_active"`
	Password string `json:"password"`
}

// Update 更新用户
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 查找用户
	var user models.User
	result := h.db.First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 检查邮箱是否已被其他用户使用
	if req.Email != "" && req.Email != user.Email {
		var count int64
		h.db.Model(&models.User{}).Where("email = ? AND id != ?", req.Email, id).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被其他用户使用"})
			return
		}
		user.Email = req.Email
	}

	// 更新其他字段
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.RoleID != nil {
		// 检查角色是否存在
		var role models.Role
		result := h.db.First(&role, *req.RoleID)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "角色不存在"})
			return
		}
		user.RoleID = *req.RoleID
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// 更新密码（如果提供）
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		user.Password = string(hashedPassword)
	}

	// 保存更新
	if result := h.db.Save(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
		return
	}

	// 记录审计日志
	auditLog := models.AuditLog{
		UserID:      &user.ID,
		Action:      "update_user",
		Description: "更新用户",
		IP:          c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
	}
	h.db.Create(&auditLog)

	c.JSON(http.StatusOK, gin.H{
		"message": "用户更新成功",
	})
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 查找用户
	var user models.User
	result := h.db.First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 删除用户（软删除）
	if result := h.db.Delete(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	// 记录审计日志
	auditLog := models.AuditLog{
		UserID:      &user.ID,
		Action:      "delete_user",
		Description: "删除用户",
		IP:          c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
	}
	h.db.Create(&auditLog)

	c.JSON(http.StatusOK, gin.H{
		"message": "用户删除成功",
	})
}

// GetCurrentUser 获取当前用户信息
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var user models.User
	result := h.db.Preload("Role").First(&user, userID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateCurrentUserRequest 更新当前用户请求
type UpdateCurrentUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

// UpdateCurrentUser 更新当前用户信息
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var req UpdateCurrentUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 查找用户
	var user models.User
	result := h.db.First(&user, userID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 检查邮箱是否已被其他用户使用
	if req.Email != "" && req.Email != user.Email {
		var count int64
		h.db.Model(&models.User{}).Where("email = ? AND id != ?", req.Email, userID).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被其他用户使用"})
			return
		}
		user.Email = req.Email
	}

	// 更新其他字段
	if req.FullName != "" {
		user.FullName = req.FullName
	}

	// 更新密码（如果提供）
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		user.Password = string(hashedPassword)
	}

	// 保存更新
	if result := h.db.Save(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
		return
	}

	// 记录审计日志
	auditLog := models.AuditLog{
		UserID:      &user.ID,
		Action:      "update_profile",
		Description: "更新个人资料",
		IP:          c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
	}
	h.db.Create(&auditLog)

	c.JSON(http.StatusOK, gin.H{
		"message": "个人资料更新成功",
	})
}