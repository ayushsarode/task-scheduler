package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/ayushsarode/task-scheduler/internal/utils"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	utils.SuccessResponse(c, http.StatusOK, gin.H{
"status":  "healthy",
"service": "task-scheduler",
"version": "1.0.0",
})
}
