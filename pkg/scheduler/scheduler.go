package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"flowforge/pkg/models"
	
	"github.com/robfig/cron/v3"
)

// Scheduler 调度器
type Scheduler struct {
	cron    *cron.Cron
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex
	running bool
	jobs    map[string]cron.EntryID
}

// Job 调度任务
type Job struct {
	ID       string
	Name     string
	Spec     string
	Func     func()
	Enabled  bool
	LastRun  *time.Time
	NextRun  *time.Time
}

// NewScheduler 创建调度器
func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 创建带有秒级精度的cron调度器
	c := cron.New(cron.WithSeconds())
	
	return &Scheduler{
		cron:   c,
		ctx:    ctx,
		cancel: cancel,
		jobs:   make(map[string]cron.EntryID),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.cron.Start()
	s.running = true
	
	log.Println("Scheduler started")
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	s.cancel()
	s.cron.Stop()
	s.running = false
	
	log.Println("Scheduler stopped")
	return nil
}

// AddJob 添加定时任务
func (s *Scheduler) AddJob(jobID, spec string, cmd func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果任务已存在，先删除
	if entryID, exists := s.jobs[jobID]; exists {
		s.cron.Remove(entryID)
	}

	// 添加新任务
	entryID, err := s.cron.AddFunc(spec, cmd)
	if err != nil {
		return fmt.Errorf("failed to add job %s: %v", jobID, err)
	}

	s.jobs[jobID] = entryID
	log.Printf("Job %s added with spec: %s", jobID, spec)
	return nil
}

// RemoveJob 删除定时任务
func (s *Scheduler) RemoveJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	s.cron.Remove(entryID)
	delete(s.jobs, jobID)
	
	log.Printf("Job %s removed", jobID)
	return nil
}

// GetJobs 获取所有任务
func (s *Scheduler) GetJobs() []Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]Job, 0, len(s.jobs))
	
	for jobID, entryID := range s.jobs {
		entry := s.cron.Entry(entryID)
		job := Job{
			ID:      jobID,
			Name:    jobID,
			Enabled: true,
		}
		
		if !entry.Next.IsZero() {
			job.NextRun = &entry.Next
		}
		if !entry.Prev.IsZero() {
			job.LastRun = &entry.Prev
		}
		
		jobs = append(jobs, job)
	}
	
	return jobs
}

// AddPipelineJob 添加流水线定时任务
func (s *Scheduler) AddPipelineJob(pipeline *models.Pipeline) error {
	if pipeline.CronExpr == "" {
		return fmt.Errorf("pipeline cron expression is empty")
	}

	jobID := fmt.Sprintf("pipeline_%d", pipeline.ID)
	
	return s.AddJob(jobID, pipeline.CronExpr, func() {
		log.Printf("Executing scheduled pipeline: %s (ID: %d)", pipeline.Name, pipeline.ID)
		// 这里应该调用流水线执行逻辑
		s.executePipeline(pipeline)
	})
}

// RemovePipelineJob 删除流水线定时任务
func (s *Scheduler) RemovePipelineJob(pipelineID uint) error {
	jobID := fmt.Sprintf("pipeline_%d", pipelineID)
	return s.RemoveJob(jobID)
}

// executePipeline 执行流水线
func (s *Scheduler) executePipeline(pipeline *models.Pipeline) {
	// 模拟流水线执行
	log.Printf("Starting pipeline execution: %s", pipeline.Name)
	
	// 这里应该包含实际的流水线执行逻辑
	// 例如：克隆代码、构建、测试、部署等步骤
	
	steps := []string{
		"Preparing environment",
		"Cloning repository",
		"Installing dependencies", 
		"Running tests",
		"Building application",
		"Deploying to target",
	}
	
	for _, step := range steps {
		log.Printf("Pipeline %s: %s", pipeline.Name, step)
		time.Sleep(1 * time.Second) // 模拟执行时间
	}
	
	log.Printf("Pipeline execution completed: %s", pipeline.Name)
}

// AddCleanupJob 添加清理任务
func (s *Scheduler) AddCleanupJob() error {
	// 每天凌晨2点执行清理任务
	return s.AddJob("cleanup", "0 0 2 * * *", func() {
		log.Println("Starting cleanup job")
		s.performCleanup()
	})
}

// performCleanup 执行清理操作
func (s *Scheduler) performCleanup() {
	// 清理临时文件
	log.Println("Cleaning up temporary files...")
	
	// 清理过期的部署记录
	log.Println("Cleaning up expired deployment records...")
	
	// 清理过期的日志文件
	log.Println("Cleaning up expired log files...")
	
	log.Println("Cleanup job completed")
}

// IsRunning 检查调度器是否运行中
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetJobCount 获取任务数量
func (s *Scheduler) GetJobCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.jobs)
}