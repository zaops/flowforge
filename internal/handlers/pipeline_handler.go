package handlers

import (
	"net/http"
	"strconv"

	"flowforge/pkg/database"
	"flowforge/pkg/models"
	"flowforge/pkg/pipeline"
	"flowforge/pkg/utils"

	"github.com/gin-gonic/gin"
)

// PipelineHandler 流水线处理器
type PipelineHandler struct{
	engine *pipeline.Engine
}

// NewPipelineHandler 创建流水线处理器
func NewPipelineHandler(engine *pipeline.Engine) *PipelineHandler {
	return &PipelineHandler{
		engine: engine,
	}
}

// GetPipelines 获取流水线列表
func (h *PipelineHandler) GetPipelines(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var pipelines []models.Pipeline
	var total int64

	query := database.DB.Model(&models.Pipeline{}).Preload("Project")
	
	// 非管理员只能查看自己的流水线
	if role, exists := c.Get("role"); !exists || role != models.RoleAdmin {
		query = query.Joins("JOIN projects ON pipelines.project_id = projects.id").
			Where("projects.user_id = ?", userID)
	}

	query.Count(&total)
	query.Scopes(database.Paginate(page, pageSize)).Find(&pipelines)

	utils.SuccessResponse(c, models.PaginationResponse{
		Data:       pipelines,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}

// CreatePipeline 创建流水线
func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
	var req models.CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	userID, _ := c.Get("user_id")

	// 检查项目是否存在且属于当前用户
	var project models.Project
	if err := database.DB.Where("id = ? AND user_id = ?", req.ProjectID, userID).First(&project).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "项目不存在")
		return
	}

	pipeline := models.Pipeline{
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
		Trigger:     req.Trigger,
		CronExpr:    req.CronExpr,
		ProjectID:   req.ProjectID,
		Status:      models.PipelineStatusActive,
	}

	if err := database.DB.Create(&pipeline).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建流水线失败")
		return
	}

	utils.SuccessResponse(c, pipeline)
}

// GetPipeline 获取流水线详情
func (h *PipelineHandler) GetPipeline(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var pipeline models.Pipeline
	query := database.DB.Preload("Project").Preload("PipelineRuns")

	// 非管理员只能查看自己的流水线
	if role, exists := c.Get("role"); !exists || role != models.RoleAdmin {
		query = query.Joins("JOIN projects ON pipelines.project_id = projects.id").
			Where("projects.user_id = ?", userID)
	}

	if err := query.First(&pipeline, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "流水线不存在")
		return
	}

	utils.SuccessResponse(c, pipeline)
}

// UpdatePipeline 更新流水线
func (h *PipelineHandler) UpdatePipeline(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var pipeline models.Pipeline
	query := database.DB.Joins("JOIN projects ON pipelines.project_id = projects.id")

	// 非管理员只能更新自己的流水线
	if role, exists := c.Get("role"); !exists || role != models.RoleAdmin {
		query = query.Where("projects.user_id = ?", userID)
	}

	if err := query.First(&pipeline, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "流水线不存在")
		return
	}

	var req models.CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	pipeline.Name = req.Name
	pipeline.Description = req.Description
	pipeline.Config = req.Config
	pipeline.Trigger = req.Trigger
	pipeline.CronExpr = req.CronExpr

	if err := database.DB.Save(&pipeline).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新流水线失败")
		return
	}

	utils.SuccessResponse(c, pipeline)
}

// DeletePipeline 删除流水线
func (h *PipelineHandler) DeletePipeline(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var pipeline models.Pipeline
	query := database.DB.Joins("JOIN projects ON pipelines.project_id = projects.id")

	// 非管理员只能删除自己的流水线
	if role, exists := c.Get("role"); !exists || role != models.RoleAdmin {
		query = query.Where("projects.user_id = ?", userID)
	}

	if err := query.First(&pipeline, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "流水线不存在")
		return
	}

	if err := database.DB.Delete(&pipeline).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除流水线失败")
		return
	}

	utils.SuccessResponse(c, nil)
}

// RunPipeline 运行流水线
func (h *PipelineHandler) RunPipeline(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	// 检查流水线是否存在且有权限
	var pipeline models.Pipeline
	query := database.DB.Preload("Project")

	// 非管理员只能运行自己的流水线
	if role, exists := c.Get("role"); !exists || role != models.RoleAdmin {
		query = query.Joins("JOIN projects ON pipelines.project_id = projects.id").
			Where("projects.user_id = ?", userID)
	}

	if err := query.First(&pipeline, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "流水线不存在")
		return
	}

	// 运行流水线
	pipelineRun, err := h.engine.RunPipeline(pipeline.ID, models.TriggerTypeManual, userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "启动流水线失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, pipelineRun)
}

// GetPipelineRuns 获取流水线运行记录
func (h *PipelineHandler) GetPipelineRuns(c *gin.Context) {
	pipelineID := c.Param("id")
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 检查流水线权限
	var pipeline models.Pipeline
	query := database.DB.Joins("JOIN projects ON pipelines.project_id = projects.id")

	if role, exists := c.Get("role"); !exists || role != models.RoleAdmin {
		query = query.Where("projects.user_id = ?", userID)
	}

	if err := query.First(&pipeline, pipelineID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "流水线不存在")
		return
	}

	// 获取运行记录
	var runs []models.PipelineRun
	var total int64

	runQuery := database.DB.Model(&models.PipelineRun{}).Where("pipeline_id = ?", pipelineID)
	runQuery.Count(&total)
	runQuery.Order("created_at DESC").Scopes(database.Paginate(page, pageSize)).Find(&runs)

	utils.SuccessResponse(c, models.PaginationResponse{
		Data:       runs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}

// GetPipelineRun 获取流水线运行详情
func (h *PipelineHandler) GetPipelineRun(c *gin.Context) {
	runID := c.Param("runId")
	userID, _ := c.Get("user_id")

	var pipelineRun models.PipelineRun
	query := database.DB.Preload("Pipeline.Project")

	// 非管理员只能查看自己的流水线运行
	if role, exists := c.Get("role"); !exists || role != models.RoleAdmin {
		query = query.Joins("JOIN pipelines ON pipeline_runs.pipeline_id = pipelines.id").
			Joins("JOIN projects ON pipelines.project_id = projects.id").
			Where("projects.user_id = ?", userID)
	}

	if err := query.First(&pipelineRun, runID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "流水线运行记录不存在")
		return
	}

	utils.SuccessResponse(c, pipelineRun)
}

// CancelPipelineRun 取消流水线运行
func (h *PipelineHandler) CancelPipelineRun(c *gin.Context) {
	runID := c.Param("runId")
	userID, _ := c.Get("user_id")

	// 检查权限
	var pipelineRun models.PipelineRun
	query := database.DB.Preload("Pipeline.Project")

	if role, exists := c.Get("role"); !exists || role != models.RoleAdmin {
		query = query.Joins("JOIN pipelines ON pipeline_runs.pipeline_id = pipelines.id").
			Joins("JOIN projects ON pipelines.project_id = projects.id").
			Where("projects.user_id = ?", userID)
	}

	if err := query.First(&pipelineRun, runID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "流水线运行记录不存在")
		return
	}

	// 取消运行
	runIDUint, _ := strconv.ParseUint(runID, 10, 32)
	if err := h.engine.CancelPipelineRun(uint(runIDUint)); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "取消流水线运行失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

// GetPipelineRunLogs 获取流水线运行日志
func (h *PipelineHandler) GetPipelineRunLogs(c *gin.Context) {
	runID := c.Param("runId")
	userID, _ := c.Get("user_id")

	// 检查权限
	var pipelineRun models.PipelineRun
	query := database.DB.Preload("Pipeline.Project")

	if role, exists := c.Get("role"); !exists || role != models.RoleAdmin {
		query = query.Joins("JOIN pipelines ON pipeline_runs.pipeline_id = pipelines.id").
			Joins("JOIN projects ON pipelines.project_id = projects.id").
			Where("projects.user_id = ?", userID)
	}

	if err := query.First(&pipelineRun, runID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "流水线运行记录不存在")
		return
	}

	// 获取日志
	runIDUint, _ := strconv.ParseUint(runID, 10, 32)
	logs, err := h.engine.GetJobLogs(uint(runIDUint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"logs": logs,
	})
}
