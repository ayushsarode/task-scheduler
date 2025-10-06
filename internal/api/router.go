package api

import (
"github.com/gin-gonic/gin"
"github.com/ayushsarode/task-scheduler/internal/api/handlers"
"github.com/ayushsarode/task-scheduler/internal/services"
)

func SetupRoutes(router *gin.Engine, taskService *services.TaskService, resultService *services.ResultService) {
	// Health check
	healthHandler := handlers.NewHealthHandler()
	router.GET("/health", healthHandler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Task handlers
		taskHandler := handlers.NewTaskHandler(taskService)
		v1.POST("/tasks", taskHandler.CreateTask)
		v1.GET("/tasks", taskHandler.ListTasks)
		v1.GET("/tasks/:id", taskHandler.GetTask)
		v1.PUT("/tasks/:id", taskHandler.UpdateTask)
		v1.DELETE("/tasks/:id", taskHandler.DeleteTask)
		v1.GET("/tasks/:id/results", taskHandler.GetTaskResults)

		// Result handlers
		resultHandler := handlers.NewResultHandler(resultService)
		v1.GET("/results", resultHandler.ListResults)
	}
}
