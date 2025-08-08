package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flowforge/internal/handlers"
	"flowforge/internal/middleware"
	"flowforge/pkg/config"
	"flowforge/pkg/database"
	"flowforge/pkg/deploy"
	"flowforge/pkg/git"
	"flowforge/pkg/pipeline"
	"flowforge/pkg/scripts"
	"flowforge/pkg/ssh"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server API服务器结构
type Server struct {
	router         *gin.Engine
	httpServer     *http.Server
	config         *config.Config
	pipelineEngine *pipeline.Engine
	scriptManager  *scripts.Manager
	gitManager     *git.Manager
	sshManager     *ssh.Manager
	deployManager  *deploy.DeployManager
}

// NewServer 创建新的API服务器
func NewServer(cfg *config.Config, pipelineEngine *pipeline.Engine, scriptManager *scripts.Manager, gitManager *git.Manager, sshManager *ssh.Manager, deployManager *deploy.DeployManager) *Server {
	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Gin路由器
	router := gin.New()

	// 创建HTTP服务器
	httpServer := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderMB << 20, // MB to bytes
	}

	return &Server{
		router:         router,
		httpServer:     httpServer,
		config:         cfg,
		pipelineEngine: pipelineEngine,
		scriptManager:  scriptManager,
		gitManager:     gitManager,
		sshManager:     sshManager,
		deployManager:  deployManager,
	}
}

// setupMiddleware 设置中间件
func (s *Server) setupMiddleware() {
	// 恢复中间件
	s.router.Use(gin.Recovery())

	// 日志中间件
	if s.config.Server.Mode == "debug" {
		s.router.Use(gin.Logger())
	} else {
		s.router.Use(middleware.Logger())
	}

	// CORS中间件
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 限流中间件
	s.router.Use(middleware.RateLimit())

	// 请求ID中间件
	s.router.Use(middleware.RequestID())

	// 安全头中间件
	s.router.Use(middleware.Security())
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// API版本组
	v1 := s.router.Group("/api/v1")

	// 认证路由（无需JWT验证）
	authGroup := v1.Group("/auth")
	{
		authHandler := handlers.NewAuthHandler()
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/refresh", authHandler.RefreshToken)
	}

	// 需要JWT验证的路由
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth())

	// 用户管理路由
	userGroup := protected.Group("/users")
	{
		userHandler := handlers.NewUserHandler()
		userGroup.GET("", userHandler.GetUsers)
		userGroup.GET("/:id", userHandler.GetUser)
		userGroup.PUT("/:id", userHandler.UpdateUser)
		userGroup.DELETE("/:id", userHandler.DeleteUser)
		userGroup.GET("/profile", userHandler.GetProfile)
		userGroup.PUT("/profile", userHandler.UpdateProfile)
		userGroup.PUT("/password", userHandler.ChangePassword)
	}

	// 项目管理路由
	projectGroup := protected.Group("/projects")
	{
		projectHandler := handlers.NewProjectHandler()
		projectGroup.GET("", projectHandler.GetProjects)
		projectGroup.POST("", projectHandler.CreateProject)
		projectGroup.GET("/:id", projectHandler.GetProject)
		projectGroup.PUT("/:id", projectHandler.UpdateProject)
		projectGroup.DELETE("/:id", projectHandler.DeleteProject)
		
		// 项目部署相关
		projectGroup.POST("/:id/deploy", projectHandler.DeployProject)
		projectGroup.GET("/:id/deployments", projectHandler.GetDeployments)
		projectGroup.GET("/:id/deployments/:deployment_id", projectHandler.GetDeployment)
		projectGroup.DELETE("/:id/deployments/:deployment_id", projectHandler.DeleteDeployment)
		
		// 项目环境变量
		projectGroup.GET("/:id/environments", projectHandler.GetEnvironments)
		projectGroup.POST("/:id/environments", projectHandler.CreateEnvironment)
		projectGroup.PUT("/:id/environments/:env_id", projectHandler.UpdateEnvironment)
		projectGroup.DELETE("/:id/environments/:env_id", projectHandler.DeleteEnvironment)
	}

	// SSH密钥管理路由
	sshGroup := protected.Group("/ssh-keys")
	{
		sshHandler := handlers.NewSSHHandler(s.sshManager)
		sshGroup.GET("", sshHandler.GetSSHKeys)
		sshGroup.POST("", sshHandler.CreateSSHKey)
		sshGroup.GET("/:id", sshHandler.GetSSHKey)
		sshGroup.PUT("/:id", sshHandler.UpdateSSHKey)
		sshGroup.DELETE("/:id", sshHandler.DeleteSSHKey)
		sshGroup.POST("/:id/test", sshHandler.TestSSHConnection)
	}

	// 流水线管理路由
	pipelineGroup := protected.Group("/pipelines")
	{
		pipelineHandler := handlers.NewPipelineHandler(s.pipelineEngine)
		pipelineGroup.GET("", pipelineHandler.GetPipelines)
		pipelineGroup.POST("", pipelineHandler.CreatePipeline)
		pipelineGroup.GET("/:id", pipelineHandler.GetPipeline)
		pipelineGroup.PUT("/:id", pipelineHandler.UpdatePipeline)
		pipelineGroup.DELETE("/:id", pipelineHandler.DeletePipeline)
		
		// 流水线执行
		pipelineGroup.POST("/:id/run", pipelineHandler.RunPipeline)
		pipelineGroup.GET("/:id/runs", pipelineHandler.GetPipelineRuns)
		pipelineGroup.GET("/:id/runs/:runId", pipelineHandler.GetPipelineRun)
		pipelineGroup.POST("/:id/runs/:runId/cancel", pipelineHandler.CancelPipelineRun)
		pipelineGroup.GET("/:id/runs/:runId/logs", pipelineHandler.GetPipelineRunLogs)
	}

	// 文件上传路由
	uploadGroup := protected.Group("/upload")
	{
		uploadHandler := handlers.NewUploadHandler()
		uploadGroup.POST("/avatar", uploadHandler.UploadAvatar)
		uploadGroup.POST("/file", uploadHandler.UploadFile)
	}

	// WebSocket路由（实时日志）
	wsGroup := protected.Group("/ws")
	{
		wsHandler := handlers.NewWebSocketHandler()
		wsGroup.GET("/logs/:deployment_id", wsHandler.HandleDeploymentLogs)
		wsGroup.GET("/pipeline/:run_id", wsHandler.HandlePipelineLogs)
	}
}

// healthCheck 健康检查处理器
func (s *Server) healthCheck(c *gin.Context) {
	// 检查数据库连接
	if err := database.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   "database connection failed",
			"details": err.Error(),
		})
		return
	}

	// 获取数据库统计信息
	dbStats, err := database.GetStats()
	if err != nil {
		log.Printf("获取数据库统计信息失败: %v", err)
		dbStats = map[string]interface{}{"error": err.Error()}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"database":  dbStats,
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	// 设置中间件
	s.setupMiddleware()

	// 设置路由
	s.setupRoutes()

	// 启动服务器
	log.Printf("服务器启动在 %s", s.httpServer.Addr)

	// 如果启用了TLS
	if s.config.Server.TLS.Enabled {
		if s.config.Server.TLS.CertFile == "" || s.config.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS已启用但证书文件未配置")
		}
		return s.httpServer.ListenAndServeTLS(s.config.Server.TLS.CertFile, s.config.Server.TLS.KeyFile)
	}

	return s.httpServer.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	log.Println("正在关闭服务器...")
	return s.httpServer.Shutdown(ctx)
}

// Run 运行服务器（带优雅关闭）
func (s *Server) Run() error {
	// 设置中间件和路由
	s.setupMiddleware()
	s.setupRoutes()

	// 创建一个通道来接收系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在goroutine中启动服务器
	go func() {
		log.Printf("服务器启动在 %s", s.httpServer.Addr)
		
		var err error
		if s.config.Server.TLS.Enabled {
			if s.config.Server.TLS.CertFile == "" || s.config.Server.TLS.KeyFile == "" {
				log.Fatal("TLS已启用但证书文件未配置")
			}
			err = s.httpServer.ListenAndServeTLS(s.config.Server.TLS.CertFile, s.config.Server.TLS.KeyFile)
		} else {
			err = s.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	<-quit
	log.Println("收到关闭信号...")

	// 创建一个超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
		return err
	}

	log.Println("服务器已关闭")
	return nil
}

// GetRouter 获取Gin路由器（用于测试）
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// RegisterCustomRoutes 注册自定义路由
func (s *Server) RegisterCustomRoutes(registerFunc func(*gin.Engine)) {
	registerFunc(s.router)
}

// SetTrustedProxies 设置信任的代理
func (s *Server) SetTrustedProxies(proxies []string) error {
	return s.router.SetTrustedProxies(proxies)
}

// LoadHTMLGlob 加载HTML模板
func (s *Server) LoadHTMLGlob(pattern string) {
	s.router.LoadHTMLGlob(pattern)
}

// Static 设置静态文件服务
func (s *Server) Static(relativePath, root string) {
	s.router.Static(relativePath, root)
}

// StaticFile 设置单个静态文件
func (s *Server) StaticFile(relativePath, filepath string) {
	s.router.StaticFile(relativePath, filepath)
}

// NoRoute 设置404处理器
func (s *Server) NoRoute(handlers ...gin.HandlerFunc) {
	s.router.NoRoute(handlers...)
}

// NoMethod 设置405处理器
func (s *Server) NoMethod(handlers ...gin.HandlerFunc) {
	s.router.NoMethod(handlers...)
}