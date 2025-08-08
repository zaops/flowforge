package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"flowforge/pkg/config"
	"flowforge/pkg/database"
	"flowforge/pkg/git"
	"flowforge/pkg/models"
	"flowforge/pkg/scripts"
)

// Engine 流水线执行引擎
type Engine struct {
	config        *config.Config
	scriptManager *scripts.Manager
	gitManager    *git.Manager
	runningJobs   map[uint]*JobContext
	mu            sync.RWMutex
}

// JobContext 任务上下文
type JobContext struct {
	PipelineRun *models.PipelineRun
	Pipeline    *models.Pipeline
	Project     *models.Project
	Context     context.Context
	Cancel      context.CancelFunc
	LogChan     chan string
}

// NewEngine 创建流水线执行引擎
func NewEngine(cfg *config.Config, scriptMgr *scripts.Manager, gitMgr *git.Manager) *Engine {
	return &Engine{
		config:        cfg,
		scriptManager: scriptMgr,
		gitManager:    gitMgr,
		runningJobs:   make(map[uint]*JobContext),
	}
}

// RunPipeline 运行流水线
func (e *Engine) RunPipeline(pipelineID uint, triggerType models.TriggerType, triggerBy uint) (*models.PipelineRun, error) {
	// 获取流水线信息
	var pipeline models.Pipeline
	if err := database.DB.Preload("Project").First(&pipeline, pipelineID).Error; err != nil {
		return nil, fmt.Errorf("获取流水线失败: %w", err)
	}

	// 创建流水线运行记录
	pipelineRun := &models.PipelineRun{
		PipelineID:  pipelineID,
		Status:      models.RunStatusRunning,
		TriggerType: triggerType,
		TriggerBy:   triggerBy,
		StartTime:   time.Now(),
	}

	if err := database.DB.Create(pipelineRun).Error; err != nil {
		return nil, fmt.Errorf("创建流水线运行记录失败: %w", err)
	}

	// 创建任务上下文
	ctx, cancel := context.WithCancel(context.Background())
	jobCtx := &JobContext{
		PipelineRun: pipelineRun,
		Pipeline:    &pipeline,
		Project:     &pipeline.Project,
		Context:     ctx,
		Cancel:      cancel,
		LogChan:     make(chan string, 100),
	}

	// 添加到运行中的任务
	e.mu.Lock()
	e.runningJobs[pipelineRun.ID] = jobCtx
	e.mu.Unlock()

	// 异步执行流水线
	go e.executePipeline(jobCtx)

	return pipelineRun, nil
}

// executePipeline 执行流水线
func (e *Engine) executePipeline(jobCtx *JobContext) {
	defer func() {
		// 清理任务上下文
		e.mu.Lock()
		delete(e.runningJobs, jobCtx.PipelineRun.ID)
		e.mu.Unlock()
		close(jobCtx.LogChan)
	}()

	// 解析流水线配置
	var config models.PipelineConfig
	if err := json.Unmarshal([]byte(jobCtx.Pipeline.Config), &config); err != nil {
		e.finishPipelineRun(jobCtx, models.RunStatusFailed, fmt.Sprintf("解析流水线配置失败: %v", err))
		return
	}

	// 记录开始日志
	e.logMessage(jobCtx, fmt.Sprintf("开始执行流水线: %s", jobCtx.Pipeline.Name))

	// 执行各个阶段
	for i, stage := range config.Stages {
		e.logMessage(jobCtx, fmt.Sprintf("执行阶段 %d: %s", i+1, stage.Name))

		if err := e.executeStage(jobCtx, &stage); err != nil {
			e.finishPipelineRun(jobCtx, models.RunStatusFailed, fmt.Sprintf("阶段 %s 执行失败: %v", stage.Name, err))
			return
		}

		e.logMessage(jobCtx, fmt.Sprintf("阶段 %s 执行完成", stage.Name))
	}

	// 流水线执行成功
	e.finishPipelineRun(jobCtx, models.RunStatusSuccess, "流水线执行成功")
}

// executeStage 执行阶段
func (e *Engine) executeStage(jobCtx *JobContext, stage *models.PipelineStage) error {
	// 执行阶段中的所有步骤
	for _, step := range stage.Steps {
		if err := e.executeStep(jobCtx, &step); err != nil {
			return fmt.Errorf("步骤 %s 执行失败: %w", step.Name, err)
		}
	}
	return nil
}

// executeStep 执行步骤
func (e *Engine) executeStep(jobCtx *JobContext, step *models.PipelineStep) error {
	e.logMessage(jobCtx, fmt.Sprintf("执行步骤: %s", step.Name))

	switch step.Type {
	case "git_clone":
		return e.executeGitClone(jobCtx, step)
	case "script":
		return e.executeScript(jobCtx, step)
	case "build":
		return e.executeBuild(jobCtx, step)
	case "deploy":
		return e.executeDeploy(jobCtx, step)
	default:
		return fmt.Errorf("不支持的步骤类型: %s", step.Type)
	}
}

// executeGitClone 执行Git克隆
func (e *Engine) executeGitClone(jobCtx *JobContext, step *models.PipelineStep) error {
	project := jobCtx.Project
	workDir := fmt.Sprintf("%s/workspaces/%d", e.config.App.DataPath, project.ID)

	// 克隆或更新代码
	if err := e.gitManager.CloneOrPull(project.RepoURL, project.Branch, workDir); err != nil {
		return fmt.Errorf("代码拉取失败: %w", err)
	}

	e.logMessage(jobCtx, "代码拉取完成")
	return nil
}

// executeScript 执行脚本
func (e *Engine) executeScript(jobCtx *JobContext, step *models.PipelineStep) error {
	script, ok := step.Config["script"].(string)
	if !ok {
		return fmt.Errorf("脚本内容不能为空")
	}

	workDir := fmt.Sprintf("%s/workspaces/%d", e.config.App.DataPath, jobCtx.Project.ID)

	// 准备环境变量
	env := map[string]string{
		"PROJECT_NAME":    jobCtx.Project.Name,
		"PROJECT_ID":      fmt.Sprintf("%d", jobCtx.Project.ID),
		"PIPELINE_ID":     fmt.Sprintf("%d", jobCtx.Pipeline.ID),
		"PIPELINE_RUN_ID": fmt.Sprintf("%d", jobCtx.PipelineRun.ID),
		"BUILD_VERSION":   fmt.Sprintf("v%d", jobCtx.PipelineRun.ID),
	}

	// 添加自定义环境变量
	if envVars, ok := step.Config["env"].(map[string]interface{}); ok {
		for k, v := range envVars {
			if str, ok := v.(string); ok {
				env[k] = str
			}
		}
	}

	// 执行脚本
	opts := scripts.ExecuteOptions{
		WorkDir: workDir,
		Env:     env,
		Timeout: 30 * time.Minute,
		LogCallback: func(line string) {
			e.logMessage(jobCtx, line)
		},
	}

	result, err := e.scriptManager.Execute(jobCtx.Context, script, opts)
	if err != nil {
		return fmt.Errorf("脚本执行失败: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("脚本执行失败，退出码: %d", result.ExitCode)
	}

	return nil
}

// executeBuild 执行构建
func (e *Engine) executeBuild(jobCtx *JobContext, step *models.PipelineStep) error {
	buildType, ok := step.Config["type"].(string)
	if !ok {
		buildType = "auto"
	}

	var script string
	builtinScripts := e.scriptManager.GetBuiltinScripts()

	switch buildType {
	case "node":
		script = builtinScripts["node_build"]
	case "go":
		script = builtinScripts["go_build"]
	case "docker":
		script = builtinScripts["docker_build"]
	default:
		// 自动检测构建类型
		workDir := fmt.Sprintf("%s/workspaces/%d", e.config.App.DataPath, jobCtx.Project.ID)
		if e.fileExists(workDir + "/package.json") {
			script = builtinScripts["node_build"]
		} else if e.fileExists(workDir + "/go.mod") {
			script = builtinScripts["go_build"]
		} else if e.fileExists(workDir + "/Dockerfile") {
			script = builtinScripts["docker_build"]
		} else {
			return fmt.Errorf("无法自动检测构建类型")
		}
	}

	// 创建脚本步骤
	scriptStep := &models.PipelineStep{
		Name: "构建",
		Type: "script",
		Config: map[string]interface{}{
			"script": script,
			"env":    step.Config["env"],
		},
	}

	return e.executeScript(jobCtx, scriptStep)
}

// executeDeploy 执行部署
func (e *Engine) executeDeploy(jobCtx *JobContext, step *models.PipelineStep) error {
	deployType, ok := step.Config["type"].(string)
	if !ok {
		deployType = "script"
	}

	switch deployType {
	case "script":
		script := e.scriptManager.GetBuiltinScripts()["deploy_script"]
		scriptStep := &models.PipelineStep{
			Name: "部署",
			Type: "script",
			Config: map[string]interface{}{
				"script": script,
				"env":    step.Config["env"],
			},
		}
		return e.executeScript(jobCtx, scriptStep)
	default:
		return fmt.Errorf("不支持的部署类型: %s", deployType)
	}
}

// logMessage 记录日志消息
func (e *Engine) logMessage(jobCtx *JobContext, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s", timestamp, message)
	
	// 发送到日志通道
	select {
	case jobCtx.LogChan <- logLine:
	default:
		// 通道满了，丢弃日志
	}

	// 同时输出到控制台
	log.Printf("Pipeline %d: %s", jobCtx.Pipeline.ID, message)
}

// finishPipelineRun 完成流水线运行
func (e *Engine) finishPipelineRun(jobCtx *JobContext, status models.RunStatus, message string) {
	endTime := time.Now()
	duration := endTime.Sub(jobCtx.PipelineRun.StartTime)

	// 更新流水线运行记录
	updates := map[string]interface{}{
		"status":   status,
		"end_time": endTime,
		"duration": int(duration.Seconds()),
		"logs":     message,
	}

	if err := database.DB.Model(jobCtx.PipelineRun).Updates(updates).Error; err != nil {
		log.Printf("更新流水线运行记录失败: %v", err)
	}

	e.logMessage(jobCtx, fmt.Sprintf("流水线执行完成，状态: %s，耗时: %v", status, duration))
}

// CancelPipelineRun 取消流水线运行
func (e *Engine) CancelPipelineRun(runID uint) error {
	e.mu.RLock()
	jobCtx, exists := e.runningJobs[runID]
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("流水线运行不存在或已完成")
	}

	// 取消上下文
	jobCtx.Cancel()

	// 更新状态
	updates := map[string]interface{}{
		"status":   models.RunStatusCancelled,
		"end_time": time.Now(),
		"logs":     "流水线运行已被取消",
	}

	return database.DB.Model(jobCtx.PipelineRun).Updates(updates).Error
}

// GetRunningJobs 获取正在运行的任务
func (e *Engine) GetRunningJobs() map[uint]*JobContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[uint]*JobContext)
	for k, v := range e.runningJobs {
		result[k] = v
	}
	return result
}

// GetJobLogs 获取任务日志
func (e *Engine) GetJobLogs(runID uint) ([]string, error) {
	e.mu.RLock()
	jobCtx, exists := e.runningJobs[runID]
	e.mu.RUnlock()

	if !exists {
		// 从数据库获取历史日志
		var pipelineRun models.PipelineRun
		if err := database.DB.First(&pipelineRun, runID).Error; err != nil {
			return nil, fmt.Errorf("流水线运行不存在")
		}
		return []string{pipelineRun.Logs}, nil
	}

	// 获取实时日志
	var logs []string
	for {
		select {
		case log := <-jobCtx.LogChan:
			logs = append(logs, log)
		default:
			return logs, nil
		}
	}
}

// fileExists 检查文件是否存在
func (e *Engine) fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}