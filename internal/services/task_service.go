package services

import (
	"fmt"
	"time"

	"github.com/ayushsarode/task-scheduler/internal/db"
	"github.com/ayushsarode/task-scheduler/internal/models"
	"github.com/ayushsarode/task-scheduler/internal/scheduler"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type TaskService struct {
	repo      *db.Repository
	scheduler *scheduler.Scheduler
}

func NewTaskService(repo *db.Repository, scheduler *scheduler.Scheduler) *TaskService {
	return &TaskService{
		repo:      repo,
		scheduler: scheduler,
	}
}

func (s *TaskService) GetRepository() *db.Repository {
	return s.repo
}

func (s *TaskService) CreateTask(req models.CreateTaskRequest) (*models.Task, error) {
	// Validate trigger
	if err := s.validateTrigger(req.Trigger); err != nil {
		return nil, err
	}

	now := time.Now()
	task := &models.Task{
		ID:        uuid.New(),
		Name:      req.Name,
		Trigger:   req.Trigger,
		Action:    req.Action,
		Status:    models.StatusScheduled,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set next run time
	if task.Trigger.Type == models.TriggerOneOff && task.Trigger.DateTime != nil {
		task.NextRun = task.Trigger.DateTime
	}

	// Save to database
	if err := s.repo.CreateTask(task); err != nil {
		return nil, err
	}

	// Schedule the task
	if err := s.scheduler.ScheduleTask(task); err != nil {
		return nil, fmt.Errorf("failed to schedule task: %w", err)
	}

	return task, nil
}

func (s *TaskService) GetTask(id uuid.UUID) (*models.Task, error) {
	return s.repo.GetTaskByID(id)
}

func (s *TaskService) ListTasks(params models.ListTasksParams) ([]models.Task, int, error) {
	return s.repo.ListTasks(params)
}

func (s *TaskService) UpdateTask(id uuid.UUID, req models.UpdateTaskRequest) (*models.Task, error) {
	// Get existing task
	task, err := s.repo.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		task.Name = *req.Name
	}

	if req.Trigger != nil {
		if err := s.validateTrigger(*req.Trigger); err != nil {
			return nil, err
		}
		task.Trigger = *req.Trigger

		// Update next run time
		if task.Trigger.Type == models.TriggerOneOff && task.Trigger.DateTime != nil {
			task.NextRun = task.Trigger.DateTime
		}
	}

	if req.Action != nil {
		task.Action = *req.Action
	}

	if req.Status != nil {
		task.Status = *req.Status
	}

	task.UpdatedAt = time.Now()

	// Save to database
	if err := s.repo.UpdateTask(task); err != nil {
		return nil, err
	}

	// Reschedule the task if still scheduled
	if task.Status == models.StatusScheduled {
		if err := s.scheduler.ScheduleTask(task); err != nil {
			return nil, fmt.Errorf("failed to reschedule task: %w", err)
		}
	} else {
		s.scheduler.RemoveTask(task.ID)
	}

	return task, nil
}

func (s *TaskService) DeleteTask(id uuid.UUID) error {
	// Remove from scheduler
	s.scheduler.RemoveTask(id)

	// Mark as cancelled in database
	return s.repo.DeleteTask(id)
}

func (s *TaskService) validateTrigger(trigger models.Trigger) error {
	switch trigger.Type {
	case models.TriggerOneOff:
		if trigger.DateTime == nil {
			return fmt.Errorf("datetime is required for one-off trigger")
		}
		if trigger.DateTime.Before(time.Now()) {
			return fmt.Errorf("datetime must be in the future")
		}
	case models.TriggerCron:
		if trigger.Cron == nil {
			return fmt.Errorf("cron expression is required for cron trigger")
		}
		// Validate cron expression
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		if _, err := parser.Parse(*trigger.Cron); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
		if _, err := parser.Parse(*trigger.Cron); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
	default:
		return fmt.Errorf("invalid trigger type: %s", trigger.Type)
	}
	return nil
}
