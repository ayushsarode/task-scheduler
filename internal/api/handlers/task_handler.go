package handlers

import (
"net/http"
"strconv"

"github.com/gin-gonic/gin"
"github.com/google/uuid"
"github.com/ayushsarode/task-scheduler/internal/models"
"github.com/ayushsarode/task-scheduler/internal/services"
"github.com/ayushsarode/task-scheduler/internal/utils"
)

type TaskHandler struct {
	taskService *services.TaskService
}

func NewTaskHandler(taskService *services.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	task, err := h.taskService.CreateTask(req)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, task)
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	var params models.ListTasksParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Set defaults
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = 10
	}

	tasks, total, err := h.taskService.ListTasks(params)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	meta := utils.CalculatePaginationMeta(params.Page, params.Limit, total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, tasks, meta)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := h.taskService.GetTask(id)
	if err != nil {
		utils.NotFoundResponse(c, "Task not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, task)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	task, err := h.taskService.UpdateTask(id, req)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	if err := h.taskService.DeleteTask(id); err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "Task cancelled successfully"})
}

func (h *TaskHandler) GetTaskResults(c *gin.Context) {
	idStr := c.Param("id")
	taskID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get task results from result service
	resultService := services.NewResultService(h.taskService.GetRepository())
	results, total, err := resultService.GetTaskResults(taskID, page, limit)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	meta := utils.CalculatePaginationMeta(page, limit, total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, results, meta)
}
