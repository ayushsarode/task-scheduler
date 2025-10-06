package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ayushsarode/task-scheduler/internal/api"
	"github.com/ayushsarode/task-scheduler/internal/config"
	"github.com/ayushsarode/task-scheduler/internal/db"
	"github.com/ayushsarode/task-scheduler/internal/scheduler"
	"github.com/ayushsarode/task-scheduler/internal/services"
	"github.com/ayushsarode/task-scheduler/internal/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	utils.InitLogger(cfg.Log.Level)
	utils.Info("Starting Task Scheduler Server...")

	// Connect to database
	database, err := db.NewPostgresDB(cfg.GetDSN())
	if err != nil {
		utils.Fatal("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run database migrations
	migrationsPath := filepath.Join("internal", "db", "migrations")
	if err := database.RunMigrations(migrationsPath); err != nil {
		utils.Fatal("Failed to run database migrations: %v", err)
	}

	// Initialize repository
	repo := db.NewRepository(database)

	// Initialize scheduler
	taskScheduler := scheduler.NewScheduler(repo)

	// Initialize services
	taskService := services.NewTaskService(repo, taskScheduler)
	resultService := services.NewResultService(repo)

	// Start scheduler
	if err := taskScheduler.Start(); err != nil {
		utils.Fatal("Failed to start scheduler: %v", err)
	}
	defer taskScheduler.Stop()

	// Setup Gin router
	if cfg.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Setup API routes
	api.SetupRoutes(router, taskService, resultService)

	// Create HTTP server
	server := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		utils.Info("Server starting on %s", cfg.GetServerAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Fatal("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Info("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		utils.Fatal("Server forced to shutdown: %v", err)
	}

	utils.Info("Server exited")
}
