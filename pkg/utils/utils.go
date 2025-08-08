package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 返回错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

// SuccessResponse 返回成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// EnsureDir 确保目录存在
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// GetFileExtension 获取文件扩展名
func GetFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// GetFilenameWithoutExtension 获取不带扩展名的文件名
func GetFilenameWithoutExtension(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// GetFileSize 获取文件大小
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// IsFileExists 检查文件是否存在
func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDirExists 检查目录是否存在
func IsDirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}