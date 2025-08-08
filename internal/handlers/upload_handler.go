package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"flowforge/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadHandler 上传处理器
type UploadHandler struct{}

// NewUploadHandler 创建上传处理器
func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadAvatar 上传头像
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("avatar")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "获取上传文件失败", err.Error())
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif"}
	if !contains(allowedExts, ext) {
		utils.ErrorResponse(c, http.StatusBadRequest, "不支持的文件类型", "")
		return
	}

	// 检查文件大小（2MB）
	if file.Size > 2*1024*1024 {
		utils.ErrorResponse(c, http.StatusBadRequest, "文件大小不能超过2MB", "")
		return
	}

	// 生成唯一文件名
	filename := uuid.New().String() + ext
	savePath := filepath.Join("./storage/avatars", filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "保存文件失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "头像上传成功", gin.H{
		"filename": filename,
		"url":      "/static/avatars/" + filename,
	})
}

// UploadFile 上传文件
func (h *UploadHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "获取上传文件失败", err.Error())
		return
	}

	// 检查文件大小（10MB）
	if file.Size > 10*1024*1024 {
		utils.ErrorResponse(c, http.StatusBadRequest, "文件大小不能超过10MB", "")
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext
	savePath := filepath.Join("./storage/files", filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "保存文件失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "文件上传成功", gin.H{
		"filename":     filename,
		"original_name": file.Filename,
		"size":        file.Size,
		"url":         "/static/files/" + filename,
	})
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}