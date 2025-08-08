package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	SSH      SSHConfig      `yaml:"ssh"`
	Deploy   DeployConfig   `yaml:"deploy"`
	Log      LogConfig      `yaml:"log"`
	Storage  StorageConfig  `yaml:"storage"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string    `yaml:"host"`
	Port         int       `yaml:"port"`
	Mode         string    `yaml:"mode"`         // debug, release, test
	ReadTimeout  int       `yaml:"read_timeout"`
	WriteTimeout int       `yaml:"write_timeout"`
	MaxHeaderMB  int       `yaml:"max_header_mb"`
	TLS          TLSConfig `yaml:"tls"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            string `yaml:"type"`             // mysql, postgres, sqlite
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	LogLevel        string `yaml:"log_level"`        // silent, error, warn, info
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireTime int    `yaml:"expire_time"` // 小时
	Issuer     string `yaml:"issuer"`
}

// SSHConfig SSH配置
type SSHConfig struct {
	KeysPath    string `yaml:"keys_path"`
	Timeout     int    `yaml:"timeout"`     // 秒
	MaxRetries  int    `yaml:"max_retries"`
	DefaultUser string `yaml:"default_user"`
	DefaultPort int    `yaml:"default_port"`
}

// DeployConfig 部署配置
type DeployConfig struct {
	WorkspaceDir      string `yaml:"workspace_dir"`
	MaxConcurrent     int    `yaml:"max_concurrent"`
	Timeout           int    `yaml:"timeout"`           // 秒
	RetryCount        int    `yaml:"retry_count"`
	CleanupAfterDays  int    `yaml:"cleanup_after_days"`
	EnableWebhook     bool   `yaml:"enable_webhook"`
	WebhookSecret     string `yaml:"webhook_secret"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`       // debug, info, warn, error
	Format     string `yaml:"format"`      // json, text
	Output     string `yaml:"output"`      // stdout, file
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`    // MB
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`     // 天
	Compress   bool   `yaml:"compress"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type  string      `yaml:"type"`       // local, s3, oss
	Local LocalConfig `yaml:"local"`
	S3    S3Config    `yaml:"s3"`
	OSS   OSSConfig   `yaml:"oss"`
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	Path string `yaml:"path"`
}

// S3Config S3存储配置
type S3Config struct {
	Region          string `yaml:"region"`
	Bucket          string `yaml:"bucket"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Endpoint        string `yaml:"endpoint"`
	UseSSL          bool   `yaml:"use_ssl"`
}

// OSSConfig 阿里云OSS配置
type OSSConfig struct {
	Endpoint        string `yaml:"endpoint"`
	Bucket          string `yaml:"bucket"`
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
}

var (
	AppConfig *Config
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML配置
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 从环境变量覆盖配置
	overrideFromEnv(&config)

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	// 设置默认值
	setDefaults(&config)

	AppConfig = &config
	return &config, nil
}

// overrideFromEnv 从环境变量覆盖配置
func overrideFromEnv(config *Config) {
	// 服务器配置
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		config.Server.Mode = mode
	}

	// 数据库配置
	if dbType := os.Getenv("DB_TYPE"); dbType != "" {
		config.Database.Type = dbType
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		if p, err := strconv.Atoi(dbPort); err == nil {
			config.Database.Port = p
		}
	}
	if dbUser := os.Getenv("DB_USERNAME"); dbUser != "" {
		config.Database.Username = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		config.Database.Password = dbPass
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.Name = dbName
	}

	// JWT配置
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWT.Secret = jwtSecret
	}
	if jwtExpire := os.Getenv("JWT_EXPIRE_TIME"); jwtExpire != "" {
		if e, err := strconv.Atoi(jwtExpire); err == nil {
			config.JWT.ExpireTime = e
		}
	}

	// 部署配置
	if workspaceDir := os.Getenv("DEPLOY_WORKSPACE_DIR"); workspaceDir != "" {
		config.Deploy.WorkspaceDir = workspaceDir
	}
	if webhookSecret := os.Getenv("WEBHOOK_SECRET"); webhookSecret != "" {
		config.Deploy.WebhookSecret = webhookSecret
	}
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("无效的服务器端口: %d", config.Server.Port)
	}

	validModes := []string{"debug", "release", "test"}
	if !contains(validModes, config.Server.Mode) {
		return fmt.Errorf("无效的服务器模式: %s", config.Server.Mode)
	}

	// 验证数据库配置
	validDBTypes := []string{"mysql", "postgres", "sqlite"}
	if !contains(validDBTypes, config.Database.Type) {
		return fmt.Errorf("不支持的数据库类型: %s", config.Database.Type)
	}

	if config.Database.Type != "sqlite" {
		if config.Database.Host == "" {
			return fmt.Errorf("数据库主机不能为空")
		}
		if config.Database.Username == "" {
			return fmt.Errorf("数据库用户名不能为空")
		}
		if config.Database.Name == "" {
			return fmt.Errorf("数据库名不能为空")
		}
	}

	// 验证JWT配置
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT密钥不能为空")
	}
	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT密钥长度不能少于32位")
	}

	// 验证存储配置
	validStorageTypes := []string{"local", "s3", "oss"}
	if !contains(validStorageTypes, config.Storage.Type) {
		return fmt.Errorf("不支持的存储类型: %s", config.Storage.Type)
	}

	return nil
}

// setDefaults 设置默认值
func setDefaults(config *Config) {
	// 服务器默认值
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "release"
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 60
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 60
	}
	if config.Server.MaxHeaderMB == 0 {
		config.Server.MaxHeaderMB = 1
	}

	// 数据库默认值
	if config.Database.Type == "" {
		config.Database.Type = "sqlite"
	}
	if config.Database.Name == "" {
		if config.Database.Type == "sqlite" {
			config.Database.Name = "flowforge.db"
		} else {
			config.Database.Name = "flowforge"
		}
	}
	if config.Database.Port == 0 {
		switch config.Database.Type {
		case "mysql":
			config.Database.Port = 3306
		case "postgres":
			config.Database.Port = 5432
		}
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 10
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 100
	}
	if config.Database.ConnMaxLifetime == 0 {
		config.Database.ConnMaxLifetime = 3600
	}
	if config.Database.LogLevel == "" {
		config.Database.LogLevel = "info"
	}

	// JWT默认值
	if config.JWT.ExpireTime == 0 {
		config.JWT.ExpireTime = 24
	}
	if config.JWT.Issuer == "" {
		config.JWT.Issuer = "flowforge"
	}

	// SSH默认值
	if config.SSH.KeysPath == "" {
		config.SSH.KeysPath = "./ssh_keys"
	}
	if config.SSH.Timeout == 0 {
		config.SSH.Timeout = 30
	}
	if config.SSH.MaxRetries == 0 {
		config.SSH.MaxRetries = 3
	}
	if config.SSH.DefaultUser == "" {
		config.SSH.DefaultUser = "root"
	}
	if config.SSH.DefaultPort == 0 {
		config.SSH.DefaultPort = 22
	}

	// 部署默认值
	if config.Deploy.WorkspaceDir == "" {
		config.Deploy.WorkspaceDir = "./workspace"
	}
	if config.Deploy.MaxConcurrent == 0 {
		config.Deploy.MaxConcurrent = 5
	}
	if config.Deploy.Timeout == 0 {
		config.Deploy.Timeout = 1800
	}
	if config.Deploy.RetryCount == 0 {
		config.Deploy.RetryCount = 3
	}
	if config.Deploy.CleanupAfterDays == 0 {
		config.Deploy.CleanupAfterDays = 7
	}

	// 日志默认值
	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	if config.Log.Format == "" {
		config.Log.Format = "json"
	}
	if config.Log.Output == "" {
		config.Log.Output = "stdout"
	}
	if config.Log.MaxSize == 0 {
		config.Log.MaxSize = 100
	}
	if config.Log.MaxBackups == 0 {
		config.Log.MaxBackups = 3
	}
	if config.Log.MaxAge == 0 {
		config.Log.MaxAge = 28
	}

	// 存储默认值
	if config.Storage.Type == "" {
		config.Storage.Type = "local"
	}
	if config.Storage.Local.Path == "" {
		config.Storage.Local.Path = "./storage"
	}
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

// GetConfig 获取应用配置
func GetConfig() *Config {
	return AppConfig
}

// IsProduction 是否为生产环境
func IsProduction() bool {
	return AppConfig != nil && AppConfig.Server.Mode == "release"
}

// IsDevelopment 是否为开发环境
func IsDevelopment() bool {
	return AppConfig != nil && AppConfig.Server.Mode == "debug"
}

// GetServerAddr 获取服务器地址
func GetServerAddr() string {
	if AppConfig == nil {
		return ":8080"
	}
	return fmt.Sprintf("%s:%d", AppConfig.Server.Host, AppConfig.Server.Port)
}

// GetDatabaseDSN 获取数据库连接字符串
func GetDatabaseDSN() string {
	if AppConfig == nil {
		return ""
	}

	cfg := AppConfig.Database
	switch cfg.Type {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	case "postgres":
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host, cfg.Username, cfg.Password, cfg.Name, cfg.Port)
	case "sqlite":
		return cfg.Name
	default:
		return ""
	}
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// ReloadConfig 重新加载配置
func ReloadConfig(configPath string) error {
	config, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	AppConfig = config
	return nil
}

// GetEnvWithDefault 获取环境变量，如果不存在则返回默认值
func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvAsInt 获取环境变量并转换为整数
func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvAsBool 获取环境变量并转换为布尔值
func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEnvAsSlice 获取环境变量并转换为字符串切片
func GetEnvAsSlice(key, separator string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, separator)
	}
	return defaultValue
}