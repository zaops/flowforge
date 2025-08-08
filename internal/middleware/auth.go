package middleware

import (
	"net/http"
	"strings"

	"flowforge/pkg/auth"
	"flowforge/pkg/config"
	"github.com/gin-gonic/gin"
)

// Auth 认证中间件
func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未提供认证信息",
			})
			c.Abort()
			return
		}

		// 检查Token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "认证格式无效",
			})
			c.Abort()
			return
		}

		// 验证Token
		token := parts[1]
		claims, err := auth.ValidateToken(token, cfg.JWT.Secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证令牌",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roleId", claims.RoleID)

		c.Next()
	}
}