package services

import (
	"github.com/google/uuid"
	"github.com/ayushsarode/task-scheduler/internal/db"
	"github.com/ayushsarode/task-scheduler/internal/models"
)

type ResultService struct {
	repo *db.Repository
}

func NewResultService(repo *db.Repository) *ResultService {
	return &ResultService{
		repo: repo,
	}
}

func (s *ResultService) GetTaskResults(taskID uuid.UUID, page, limit int) ([]models.TaskResult, int, error) {
	return s.repo.GetTaskResults(taskID, page, limit)
}

func (s *ResultService) ListAllResults(params models.ListResultsParams) ([]models.TaskResult, int, error) {
	return s.repo.ListAllResults(params)
}