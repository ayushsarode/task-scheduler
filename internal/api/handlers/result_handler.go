package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/google/uuid"
"github.com/ayushsarode/task-scheduler/internal/models"
"github.com/ayushsarode/task-scheduler/internal/services"
"github.com/ayushsarode/task-scheduler/internal/utils"
)

type ResultHandler struct {
	resultService *services.ResultService
}

func NewResultHandler(resultService *services.ResultService) *ResultHandler {
	return &ResultHandler{
		resultService: resultService,
	}
}

func (h *ResultHandler) ListResults(c *gin.Context) {
	var params models.ListResultsParams
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

	// Parse task_id if provided
	if taskIDStr := c.Query("task_id"); taskIDStr != "" {
		if taskID, err := uuid.Parse(taskIDStr); err == nil {
			params.TaskID = &taskID
		}
	}

	results, total, err := h.resultService.ListAllResults(params)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	meta := utils.CalculatePaginationMeta(params.Page, params.Limit, total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, results, meta)
}
