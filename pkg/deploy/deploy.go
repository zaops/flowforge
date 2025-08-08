package deploy

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"flowforge/pkg/config"
	"flowforge/pkg/models"
)

// DeployManager 部署管理器
type DeployManager struct {
	config   *config.Config
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
	running  bool
	tasks    map[string]*DeployTask
}

// DeployTask 部署任务
type DeployTask struct {
	ID        string
	ProjectID uint
	Status    string
	StartTime time.Time
	EndTime   *time.Time
	Logs      []string
	mu        sync.RWMutex
}

// NewDeployManager 创建部署管理器
func NewDeployManager(cfg *config.Config) *DeployManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DeployManager{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
		tasks:  make(map[string]*DeployTask),
	}
}

// Start 启动部署管理器
func (dm *DeployManager) Start() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.running {
		return fmt.Errorf("deploy manager is already running")
	}

	dm.running = true
	log.Println("Deploy manager started")
	return nil
}

// Stop 停止部署管理器
func (dm *DeployManager) Stop() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if !dm.running {
		return fmt.Errorf("deploy manager is not running")
	}

	dm.cancel()
	dm.running = false
	log.Println("Deploy manager stopped")
	return nil
}

// CreateDeployTask 创建部署任务
func (dm *DeployManager) CreateDeployTask(projectID uint) (*DeployTask, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	taskID := fmt.Sprintf("deploy_%d_%d", projectID, time.Now().Unix())
	task := &DeployTask{
		ID:        taskID,
		ProjectID: projectID,
		Status:    "pending",
		StartTime: time.Now(),
		Logs:      make([]string, 0),
	}

	dm.tasks[taskID] = task
	return task, nil
}

// GetDeployTask 获取部署任务
func (dm *DeployManager) GetDeployTask(taskID string) (*DeployTask, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	task, exists := dm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("deploy task not found: %s", taskID)
	}

	return task, nil
}

// ExecuteDeploy 执行部署
func (dm *DeployManager) ExecuteDeploy(project *models.Project) error {
	task, err := dm.CreateDeployTask(project.ID)
	if err != nil {
		return err
	}

	go dm.runDeployTask(task, project)
	return nil
}

// runDeployTask 运行部署任务
func (dm *DeployManager) runDeployTask(task *DeployTask, project *models.Project) {
	task.mu.Lock()
	task.Status = "running"
	task.mu.Unlock()

	// 模拟部署过程
	steps := []string{
		"Initializing deployment...",
		"Cloning repository...",
		"Installing dependencies...",
		"Building application...",
		"Deploying to server...",
		"Deployment completed successfully",
	}

	for i, step := range steps {
		task.mu.Lock()
		task.Logs = append(task.Logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), step))
		task.mu.Unlock()

		// 模拟每个步骤的执行时间
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	task.mu.Lock()
	task.Status = "completed"
	now := time.Now()
	task.EndTime = &now
	task.mu.Unlock()

	log.Printf("Deploy task %s completed for project %d", task.ID, project.ID)
}

// AddLog 添加日志
func (dt *DeployTask) AddLog(message string) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	
	logEntry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message)
	dt.Logs = append(dt.Logs, logEntry)
}

// GetLogs 获取日志
func (dt *DeployTask) GetLogs() []string {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	
	logs := make([]string, len(dt.Logs))
	copy(logs, dt.Logs)
	return logs
}

// GetStatus 获取状态
func (dt *DeployTask) GetStatus() string {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return dt.Status
}