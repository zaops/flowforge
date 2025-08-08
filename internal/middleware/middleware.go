package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求头中是否已有请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置请求ID到上下文和响应头
		c.Set("requestId", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 获取请求ID
		requestID, exists := c.Get("requestId")
		if !exists {
			requestID = "unknown"
		}

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重置请求体，以便后续处理
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 获取响应状态
		status := c.Writer.Status()

		// 记录日志
		fmt.Printf("[%s] %s | %3d | %13v | %15s | %s | %s\n",
			end.Format("2006/01/02 - 15:04:05"),
			requestID,
			status,
			latency,
			c.ClientIP(),
			c.Request.Method,
			c.Request.URL.Path,
		)
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}