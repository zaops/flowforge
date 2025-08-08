package handlers

import (
	"net/http"
	"time"

	"flowforge/pkg/auth"
	"flowforge/pkg/config"
	"flowforge/pkg/database"
	"flowforge/pkg/models"
	"flowforge/pkg/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler 认证处理器
type AuthHandler struct{}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 查找用户
	var user models.User
	if err := database.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "数据库查询失败")
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 检查用户状态
	if user.Status != models.StatusActive {
		utils.ErrorResponse(c, http.StatusForbidden, "用户账户已被禁用")
		return
	}

	// 获取配置
	cfg := config.GetConfig()
	
	// 生成JWT令牌 - 需要转换Role字符串为RoleID
	var roleID uint = 2 // 默认用户角色
	if user.Role == models.RoleAdmin {
		roleID = 1
	}
	
		expirationTime := time.Now().Add(time.Duration(cfg.JWT.ExpireTime) * time.Hour)
	token, err := auth.GenerateToken(user.ID, user.Username, roleID, cfg.JWT.Secret, expirationTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "生成令牌失败")
		return
	}

	// 更新最后登录时间
	database.DB.Model(&user).Update("updated_at", time.Now())

	// 返回登录响应
	response := models.LoginResponse{
		Token: token,
		User:  user,
	}

	utils.SuccessResponse(c, response)
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := database.DB.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "用户名已存在")
		return
	}

	// 检查邮箱是否已存在
	if err := database.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "邮箱已存在")
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "密码加密失败")
		return
	}
	user.Password = string(hashedPassword)

	// 设置默认值
	if user.Role == "" {
		user.Role = models.RoleUser
	}
	user.Status = models.StatusActive

	// 创建用户
	if err := database.DB.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建用户失败")
		return
	}

	// 清除密码字段
	user.Password = ""

	utils.SuccessResponse(c, user)
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从请求头获取当前令牌
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "缺少认证令牌")
		return
	}

	// 移除 "Bearer " 前缀
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// 获取配置
	cfg := config.GetConfig()

	// 解析令牌
	claims, err := auth.ValidateToken(tokenString, cfg.JWT.Secret)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "无效的令牌")
		return
	}

	// 查找用户
	var user models.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 检查用户状态
	if user.Status != models.StatusActive {
		utils.ErrorResponse(c, http.StatusForbidden, "用户账户已被禁用")
		return
	}

	// 生成新的令牌
	var roleID uint = 2 // 默认用户角色
	if user.Role == models.RoleAdmin {
		roleID = 1
	}
	
	expirationTime := time.Now().Add(time.Duration(cfg.JWT.ExpireTime) * time.Hour)
	newToken, err := auth.GenerateToken(user.ID, user.Username, roleID, cfg.JWT.Secret, expirationTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "生成令牌失败")
		return
	}

	utils.SuccessResponse(c, gin.H{"token": newToken})
}
