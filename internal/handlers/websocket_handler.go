package handlers

import (
	"log"
	"net/http"

	"flowforge/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	upgrader websocket.Upgrader
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// HandleDeploymentLogs 处理部署日志WebSocket连接
func (h *WebSocketHandler) HandleDeploymentLogs(c *gin.Context) {
	deploymentID := c.Param("deployment_id")
	if deploymentID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "缺少部署ID", "")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	for {
		err := conn.WriteMessage(websocket.TextMessage, []byte("部署日志实时推送"))
		if err != nil {
			break
		}
		
		_, _, err = conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// HandlePipelineLogs 处理流水线日志WebSocket连接
func (h *WebSocketHandler) HandlePipelineLogs(c *gin.Context) {
	runID := c.Param("run_id")
	if runID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "缺少运行ID", "")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	for {
		err := conn.WriteMessage(websocket.TextMessage, []byte("流水线日志实时推送"))
		if err != nil {
			break
		}
		
		_, _, err = conn.ReadMessage()
		if err != nil {
			break
		}
	}
}