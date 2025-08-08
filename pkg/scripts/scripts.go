package scripts

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"flowforge/pkg/config"
	"flowforge/pkg/models"
)

// Manager 脚本管理器
type Manager struct {
	config *config.Config
	mu     sync.RWMutex
}

// NewManager 创建脚本管理器
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

// ExecuteOptions 执行选项
type ExecuteOptions struct {
	WorkDir     string
	Env         map[string]string
	Timeout     time.Duration
	LogCallback func(string)
}

// ExecuteResult 执行结果
type ExecuteResult struct {
	ExitCode int
	Output   string
	Error    string
	Duration time.Duration
}

// Execute 执行脚本
func (m *Manager) Execute(ctx context.Context, script string, opts ExecuteOptions) (*ExecuteResult, error) {
	startTime := time.Now()
	
	// 创建临时脚本文件
	scriptFile, err := m.createTempScript(script)
	if err != nil {
		return nil, fmt.Errorf("创建临时脚本失败: %w", err)
	}
	defer os.Remove(scriptFile)

	// 设置超时上下文
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// 创建命令
	var cmd *exec.Cmd
	if strings.HasSuffix(scriptFile, ".sh") {
		cmd = exec.CommandContext(ctx, "bash", scriptFile)
	} else if strings.HasSuffix(scriptFile, ".ps1") {
		cmd = exec.CommandContext(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-File", scriptFile)
	} else {
		cmd = exec.CommandContext(ctx, scriptFile)
	}

	// 设置工作目录
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range opts.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// 创建管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("创建stdout管道失败: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("创建stderr管道失败: %w", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动命令失败: %w", err)
	}

	// 读取输出
	var outputBuilder strings.Builder
	var errorBuilder strings.Builder
	var wg sync.WaitGroup

	// 读取stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuilder.WriteString(line + "\n")
			if opts.LogCallback != nil {
				opts.LogCallback(line)
			}
		}
	}()

	// 读取stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			errorBuilder.WriteString(line + "\n")
			if opts.LogCallback != nil {
				opts.LogCallback("ERROR: " + line)
			}
		}
	}()

	// 等待命令完成
	err = cmd.Wait()
	wg.Wait()

	duration := time.Since(startTime)
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("命令执行失败: %w", err)
		}
	}

	return &ExecuteResult{
		ExitCode: exitCode,
		Output:   outputBuilder.String(),
		Error:    errorBuilder.String(),
		Duration: duration,
	}, nil
}

// createTempScript 创建临时脚本文件
func (m *Manager) createTempScript(script string) (string, error) {
	// 确保脚本目录存在
	scriptDir := filepath.Join(m.config.Deploy.WorkspaceDir, "scripts", "temp")
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return "", fmt.Errorf("创建脚本目录失败: %w", err)
	}

	// 根据操作系统选择脚本扩展名
	var ext string
	if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		ext = ".ps1"
	} else {
		ext = ".sh"
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp(scriptDir, "script_*"+ext)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer tempFile.Close()

	// 写入脚本内容
	if _, err := tempFile.WriteString(script); err != nil {
		return "", fmt.Errorf("写入脚本内容失败: %w", err)
	}

	// 设置执行权限
	if err := os.Chmod(tempFile.Name(), 0755); err != nil {
		return "", fmt.Errorf("设置执行权限失败: %w", err)
	}

	return tempFile.Name(), nil
}

// ValidateScript 验证脚本语法
func (m *Manager) ValidateScript(script string, scriptType string) error {
	switch scriptType {
	case models.ScriptTypeBash:
		return m.validateBashScript(script)
	case models.ScriptTypePowerShell:
		return m.validatePowerShellScript(script)
	case models.ScriptTypePython:
		return m.validatePythonScript(script)
	default:
		return fmt.Errorf("不支持的脚本类型: %s", scriptType)
	}
}

// validateBashScript 验证Bash脚本
func (m *Manager) validateBashScript(script string) error {
	// 创建临时脚本文件
	scriptFile, err := m.createTempScript(script)
	if err != nil {
		return err
	}
	defer os.Remove(scriptFile)

	// 使用bash -n检查语法
	cmd := exec.Command("bash", "-n", scriptFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Bash脚本语法错误: %w", err)
	}

	return nil
}

// validatePowerShellScript 验证PowerShell脚本
func (m *Manager) validatePowerShellScript(script string) error {
	// 创建临时脚本文件
	scriptFile, err := m.createTempScript(script)
	if err != nil {
		return err
	}
	defer os.Remove(scriptFile)

	// 使用PowerShell检查语法
	cmd := exec.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("Get-Command -Syntax (Get-Content '%s' -Raw)", scriptFile))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("PowerShell脚本语法错误: %w", err)
	}

	return nil
}

// validatePythonScript 验证Python脚本
func (m *Manager) validatePythonScript(script string) error {
	// 创建临时脚本文件
	scriptFile, err := m.createTempScript(script)
	if err != nil {
		return err
	}
	defer os.Remove(scriptFile)

	// 使用python -m py_compile检查语法
	cmd := exec.Command("python", "-m", "py_compile", scriptFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Python脚本语法错误: %w", err)
	}

	return nil
}

// GetBuiltinScripts 获取内置脚本模板
func (m *Manager) GetBuiltinScripts() map[string]string {
	return map[string]string{
		"node_build": `#!/bin/bash
# Node.js 项目构建脚本
set -e

echo "开始构建 Node.js 项目..."

# 安装依赖
if [ -f "package.json" ]; then
    echo "安装 npm 依赖..."
    npm install
fi

# 运行构建
if [ -f "package.json" ] && npm run | grep -q "build"; then
    echo "运行构建命令..."
    npm run build
fi

echo "Node.js 项目构建完成"
`,
		"go_build": `#!/bin/bash
# Go 项目构建脚本
set -e

echo "开始构建 Go 项目..."

# 下载依赖
echo "下载 Go 模块依赖..."
go mod download

# 运行测试
echo "运行测试..."
go test ./...

# 构建项目
echo "构建项目..."
go build -o app ./cmd/server

echo "Go 项目构建完成"
`,
		"docker_build": `#!/bin/bash
# Docker 构建脚本
set -e

echo "开始 Docker 构建..."

# 构建镜像
if [ -f "Dockerfile" ]; then
    echo "构建 Docker 镜像..."
    docker build -t $PROJECT_NAME:$BUILD_VERSION .
    
    echo "Docker 镜像构建完成: $PROJECT_NAME:$BUILD_VERSION"
else
    echo "未找到 Dockerfile"
    exit 1
fi
`,
		"deploy_script": `#!/bin/bash
# 部署脚本
set -e

echo "开始部署应用..."

# 停止旧服务
echo "停止旧服务..."
sudo systemctl stop $SERVICE_NAME || true

# 备份旧版本
if [ -f "$DEPLOY_PATH/$APP_NAME" ]; then
    echo "备份旧版本..."
    sudo cp "$DEPLOY_PATH/$APP_NAME" "$DEPLOY_PATH/$APP_NAME.backup.$(date +%Y%m%d_%H%M%S)"
fi

# 复制新版本
echo "复制新版本..."
sudo cp ./app "$DEPLOY_PATH/$APP_NAME"
sudo chmod +x "$DEPLOY_PATH/$APP_NAME"

# 启动新服务
echo "启动新服务..."
sudo systemctl start $SERVICE_NAME
sudo systemctl enable $SERVICE_NAME

echo "部署完成"
`,
	}
}

// ExecuteBuiltinScript 执行内置脚本
func (m *Manager) ExecuteBuiltinScript(ctx context.Context, scriptName string, opts ExecuteOptions) (*ExecuteResult, error) {
	builtinScripts := m.GetBuiltinScripts()
	script, exists := builtinScripts[scriptName]
	if !exists {
		return nil, fmt.Errorf("内置脚本不存在: %s", scriptName)
	}

	return m.Execute(ctx, script, opts)
}

// StreamExecute 流式执行脚本
func (m *Manager) StreamExecute(ctx context.Context, script string, opts ExecuteOptions, output io.Writer) error {
	// 创建临时脚本文件
	scriptFile, err := m.createTempScript(script)
	if err != nil {
		return fmt.Errorf("创建临时脚本失败: %w", err)
	}
	defer os.Remove(scriptFile)

	// 设置超时上下文
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// 创建命令
	var cmd *exec.Cmd
	if strings.HasSuffix(scriptFile, ".sh") {
		cmd = exec.CommandContext(ctx, "bash", scriptFile)
	} else if strings.HasSuffix(scriptFile, ".ps1") {
		cmd = exec.CommandContext(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-File", scriptFile)
	} else {
		cmd = exec.CommandContext(ctx, scriptFile)
	}

	// 设置工作目录
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range opts.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// 设置输出
	cmd.Stdout = output
	cmd.Stderr = output

	// 执行命令
	return cmd.Run()
}