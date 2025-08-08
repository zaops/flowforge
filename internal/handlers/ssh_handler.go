package handlers

import (
	"net/http"
	"strconv"

	"flowforge/pkg/database"
	"flowforge/pkg/models"
	"flowforge/pkg/ssh"
	"flowforge/pkg/utils"

	"github.com/gin-gonic/gin"
)

// SSHHandler SSH处理器
type SSHHandler struct {
	sshManager *ssh.Manager
}

// NewSSHHandler 创建SSH处理器
func NewSSHHandler(sshManager *ssh.Manager) *SSHHandler {
	return &SSHHandler{
		sshManager: sshManager,
	}
}

// GetSSHKeys 获取SSH密钥列表
func (h *SSHHandler) GetSSHKeys(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var sshKeys []models.SSHKey
	var total int64

	query := database.DB.Model(&models.SSHKey{}).Where("user_id = ?", userID)
	query.Count(&total)
	query.Scopes(database.Paginate(page, pageSize)).Find(&sshKeys)

	utils.SuccessResponse(c, "获取SSH密钥列表成功", models.PaginationResponse{
		Data:       sshKeys,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}

// CreateSSHKey 创建SSH密钥
func (h *SSHHandler) CreateSSHKey(c *gin.Context) {
	var req models.CreateSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	userID, _ := c.Get("user_id")

	// 生成SSH密钥对
	publicKey, privateKey, err := ssh.GenerateKeyPair()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "生成SSH密钥失败", err.Error())
		return
	}

	sshKey := models.SSHKey{
		Name:       req.Name,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		UserID:     userID.(uint),
		Status:     models.StatusActive,
	}

	if sshKey.Port == 0 {
		sshKey.Port = 22
	}
	if sshKey.Username == "" {
		sshKey.Username = "root"
	}

	if err := database.DB.Create(&sshKey).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建SSH密钥失败", err.Error())
		return
	}

	// 清除私钥字段
	sshKey.PrivateKey = ""

	utils.SuccessResponse(c, "创建SSH密钥成功", sshKey)
}

// GetSSHKey 获取SSH密钥详情
func (h *SSHHandler) GetSSHKey(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var sshKey models.SSHKey
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&sshKey).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "SSH密钥不存在", "")
		return
	}

	// 清除私钥字段
	sshKey.PrivateKey = ""

	utils.SuccessResponse(c, "获取SSH密钥详情成功", sshKey)
}

// UpdateSSHKey 更新SSH密钥
func (h *SSHHandler) UpdateSSHKey(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var sshKey models.SSHKey
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&sshKey).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "SSH密钥不存在", "")
		return
	}

	var req models.CreateSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	sshKey.Name = req.Name
	sshKey.Host = req.Host
	sshKey.Port = req.Port
	sshKey.Username = req.Username

	if err := database.DB.Save(&sshKey).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新SSH密钥失败", err.Error())
		return
	}

	// 清除私钥字段
	sshKey.PrivateKey = ""

	utils.SuccessResponse(c, "更新SSH密钥成功", sshKey)
}

// DeleteSSHKey 删除SSH密钥
func (h *SSHHandler) DeleteSSHKey(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var sshKey models.SSHKey
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&sshKey).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "SSH密钥不存在", "")
		return
	}

	if err := database.DB.Delete(&sshKey).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除SSH密钥失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "删除SSH密钥成功", nil)
}

// TestSSHConnection 测试SSH连接
func (h *SSHHandler) TestSSHConnection(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var sshKey models.SSHKey
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&sshKey).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "SSH密钥不存在", "")
		return
	}

	// 测试SSH连接
	config := ssh.SSHConfig{
		Host:       sshKey.Host,
		Port:       sshKey.Port,
		Username:   sshKey.Username,
		PrivateKey: sshKey.PrivateKey,
	}

	if err := ssh.TestConnection(config); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "SSH连接测试失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "SSH连接测试成功", nil)
}