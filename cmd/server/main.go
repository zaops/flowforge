package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"flowforge/pkg/api"
	"flowforge/pkg/config"
	"flowforge/pkg/database"
	"flowforge/pkg/deploy"
	"flowforge/pkg/git"
	"flowforge/pkg/pipeline"
	"flowforge/pkg/scheduler"
	"flowforge/pkg/scripts"
	"flowforge/pkg/ssh"
	
	"github.com/gin-gonic/gin"
)

var (
	configPath = flag.String("config", "config.yaml", "配置文件路径")
	version    = flag.Bool("version", false, "显示版本信息")
	help       = flag.Bool("help", false, "显示帮助信息")
)

const (
	AppName    = "FlowForge"
	AppVersion = "1.0.0"
	AppDesc    = "现代化的部署工具"
)

func main() {
	flag.Parse()

	// 显示版本信息
	if *version {
		showVersion()
		return
	}

	// 显示帮助信息
	if *help {
		showHelp()
		return
	}

	// 初始化应用
	if err := initApp(); err != nil {
		log.Fatalf("应用初始化失败: %v", err)
	}

	log.Printf("%s v%s 启动成功", AppName, AppVersion)
}

// initApp 初始化应用
func initApp() error {
	// 1. 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		return err
	}

	// 2. 初始化数据库
	if err := database.InitDatabase(cfg); err != nil {
		return err
	}

	// 3. 自动迁移数据库表结构
	if err := database.AutoMigrate(); err != nil {
		return err
	}

	// 4. 初始化种子数据
	if err := database.SeedData(); err != nil {
		return err
	}

	// 5. 创建必要的目录
	if err := createDirectories(cfg); err != nil {
		return err
	}

	// 6. 初始化各种管理器
	scriptManager := scripts.NewManager(cfg)
	gitManager := git.NewManager(cfg)
	sshManager := ssh.NewManager(cfg)
	deployManager := deploy.NewDeployManager(cfg)
	pipelineEngine := pipeline.NewEngine(cfg, scriptManager, gitManager)

	// 7. 启动部署管理器
	if err := deployManager.Start(); err != nil {
		return err
	}

	// 8. 初始化调度器
	scheduler := scheduler.NewScheduler()
	if err := scheduler.Start(); err != nil {
		return err
	}

	// 9. 创建并启动API服务器
	server := api.NewServer(cfg, pipelineEngine, scriptManager, gitManager, sshManager, deployManager)
	
	// 设置静态文件服务
	server.Static("/static", "./web/dist")
	server.StaticFile("/", "./web/dist/index.html")
	
	// 设置404处理
	server.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})

	// 启动服务器（带优雅关闭）
	return server.Run()
}

// createDirectories 创建必要的目录
func createDirectories(cfg *config.Config) error {
	dirs := []string{
		cfg.Deploy.WorkspaceDir,
		cfg.SSH.KeysPath,
		cfg.Storage.Local.Path,
		filepath.Dir(cfg.Log.Filename),
		filepath.Join(cfg.App.DataPath, "workspaces"),
		filepath.Join(cfg.App.DataPath, "scripts"),
		"./web/dist",
		"./logs",
		"./tmp",
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// showVersion 显示版本信息
func showVersion() {
	log.Printf("%s v%s", AppName, AppVersion)
	log.Printf("Description: %s", AppDesc)
}

// showHelp 显示帮助信息
func showHelp() {
	log.Printf("%s v%s - %s", AppName, AppVersion, AppDesc)
	log.Println()
	log.Println("Usage:")
	log.Printf("  %s [options]", os.Args[0])
	log.Println()
	log.Println("Options:")
	flag.PrintDefaults()
	log.Println()
	log.Println("Examples:")
	log.Printf("  %s -config=config.yaml", os.Args[0])
	log.Printf("  %s -version", os.Args[0])
	log.Printf("  %s -help", os.Args[0])
}