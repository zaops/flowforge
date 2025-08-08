package database

import (
	"fmt"
	"log"
	"time"

	"flowforge/pkg/config"
	"flowforge/pkg/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.Config) error {
	var err error
	var dialector gorm.Dialector

	// 根据配置选择数据库驱动
	switch cfg.Database.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
		)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Database.Host,
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Name,
			cfg.Database.Port,
		)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.Database.Name)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", cfg.Database.Type)
	}

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(getLogLevel(cfg.Database.LogLevel)),
	}

	// 连接数据库
	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	log.Println("数据库连接成功")
	return nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 定义需要迁移的模型
	models := []interface{}{
		&models.User{},
		&models.Project{},
		&models.SSHKey{},
		&models.Deployment{},
		&models.Pipeline{},
		&models.PipelineRun{},
		&models.PipelineStep{},
		&models.Environment{},
		&models.Webhook{},
		&models.SystemConfig{},
	}

	// 执行自动迁移
	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("迁移模型 %T 失败: %v", model, err)
		}
	}

	log.Println("数据库表结构迁移完成")
	return nil
}

// SeedData 初始化种子数据
func SeedData() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 创建默认管理员用户
	if err := createDefaultAdmin(); err != nil {
		return fmt.Errorf("创建默认管理员失败: %v", err)
	}

	// 创建默认系统配置
	if err := createDefaultSystemConfig(); err != nil {
		return fmt.Errorf("创建默认系统配置失败: %v", err)
	}

	log.Println("种子数据初始化完成")
	return nil
}

// createDefaultAdmin 创建默认管理员用户
func createDefaultAdmin() error {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)
	
	if count > 0 {
		log.Println("管理员用户已存在，跳过创建")
		return nil
	}

	// 创建默认管理员
	admin := models.User{
		Username: "admin",
		Email:    "admin@flowforge.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		Role:     models.RoleAdmin,
		Status:   models.StatusActive,
	}

	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	log.Printf("默认管理员用户创建成功: %s", admin.Username)
	return nil
}

// createDefaultSystemConfig 创建默认系统配置
func createDefaultSystemConfig() error {
	configs := []models.SystemConfig{
		{
			Key:         "site_name",
			Value:       "FlowForge",
			Description: "网站名称",
			Category:    "general",
			IsPublic:    true,
		},
		{
			Key:         "site_description",
			Value:       "现代化的部署工具",
			Description: "网站描述",
			Category:    "general",
			IsPublic:    true,
		},
		{
			Key:         "max_concurrent_deployments",
			Value:       "5",
			Description: "最大并发部署数量",
			Category:    "deployment",
			IsPublic:    false,
		},
		{
			Key:         "deployment_timeout",
			Value:       "1800",
			Description: "部署超时时间（秒）",
			Category:    "deployment",
			IsPublic:    false,
		},
		{
			Key:         "log_retention_days",
			Value:       "30",
			Description: "日志保留天数",
			Category:    "system",
			IsPublic:    false,
		},
		{
			Key:         "enable_webhook",
			Value:       "true",
			Description: "启用Webhook功能",
			Category:    "integration",
			IsPublic:    false,
		},
	}

	for _, config := range configs {
		var existing models.SystemConfig
		result := DB.Where("key = ?", config.Key).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			if err := DB.Create(&config).Error; err != nil {
				return err
			}
			log.Printf("创建系统配置: %s", config.Key)
		}
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// getLogLevel 获取日志级别
func getLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}

// Transaction 事务处理
func Transaction(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}

// Paginate 分页查询
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		if pageSize > 100 {
			pageSize = 100
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// Search 搜索查询
func Search(fields []string, keyword string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" {
			return db
		}

		var conditions []string
		var values []interface{}

		for _, field := range fields {
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
			values = append(values, "%"+keyword+"%")
		}

		if len(conditions) > 0 {
			query := fmt.Sprintf("(%s)", fmt.Sprintf("%s", conditions[0]))
			for i := 1; i < len(conditions); i++ {
				query += fmt.Sprintf(" OR (%s)", conditions[i])
			}
			return db.Where(query, values...)
		}

		return db
	}
}

// OrderBy 排序查询
func OrderBy(sort, order string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if sort == "" {
			sort = "id"
		}
		if order == "" {
			order = "desc"
		}
		if order != "asc" && order != "desc" {
			order = "desc"
		}

		return db.Order(fmt.Sprintf("%s %s", sort, order))
	}
}

// HealthCheck 数据库健康检查
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	return nil
}

// GetStats 获取数据库统计信息
func GetStats() (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}, nil
}