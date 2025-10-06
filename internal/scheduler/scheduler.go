package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/ayushsarode/task-scheduler/internal/db"
	"github.com/ayushsarode/task-scheduler/internal/models"
)

type Scheduler struct {
	cron     *cron.Cron
	repo     *db.Repository
	executor *Executor
	jobs     map[uuid.UUID]cron.EntryID
	mu       sync.RWMutex
	stopCh   chan struct{}
}

func NewScheduler(repo *db.Repository) *Scheduler {
	return &Scheduler{
		cron:     cron.New(cron.WithSeconds()),
		repo:     repo,
		executor: NewExecutor(repo),
		jobs:     make(map[uuid.UUID]cron.EntryID),
		stopCh:   make(chan struct{}),
	}
}

func (s *Scheduler) Start() error {
	log.Println("Starting scheduler...")

	// Load existing scheduled tasks
	if err := s.loadTasks(); err != nil {
		return err
	}

	// Start cron scheduler
	s.cron.Start()

	// Start background worker for one-off tasks
	go s.runOneOffWorker()

	log.Println("Scheduler started successfully")
	return nil
}

func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	close(s.stopCh)
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) loadTasks() error {
	tasks, err := s.repo.GetScheduledTasks()
	if err != nil {
		return err
	}

	log.Printf("Loading %d scheduled tasks", len(tasks))

	for _, task := range tasks {
		if err := s.ScheduleTask(&task); err != nil {
			log.Printf("Failed to schedule task %s: %v", task.ID, err)
		}
	}

	return nil
}

func (s *Scheduler) ScheduleTask(task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing job if any
	if entryID, exists := s.jobs[task.ID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, task.ID)
	}

	switch task.Trigger.Type {
	case models.TriggerCron:
		return s.scheduleCronTask(task)
	case models.TriggerOneOff:
		return s.scheduleOneOffTask(task)
	default:
		log.Printf("Unknown trigger type: %s", task.Trigger.Type)
	}

	return nil
}

func (s *Scheduler) scheduleCronTask(task *models.Task) error {
	if task.Trigger.Cron == nil {
		log.Printf("Cron expression is nil for task %s", task.ID)
		return nil
	}

	cronExpr := *task.Trigger.Cron
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executor.ExecuteTask(task)
	})

	if err != nil {
		return err
	}

	s.jobs[task.ID] = entryID

	// Calculate next run time
	entry := s.cron.Entry(entryID)
	nextRun := entry.Next
	task.NextRun = &nextRun

	// Update task in database
	if err := s.repo.UpdateTask(task); err != nil {
		log.Printf("Failed to update next_run for task %s: %v", task.ID, err)
	}

	log.Printf("Scheduled cron task: %s (%s), next run: %s", task.Name, cronExpr, nextRun)
	return nil
}

func (s *Scheduler) scheduleOneOffTask(task *models.Task) error {
	if task.Trigger.DateTime == nil {
		log.Printf("DateTime is nil for one-off task %s", task.ID)
		return nil
	}

	scheduledTime := *task.Trigger.DateTime
	task.NextRun = &scheduledTime

	// Update task in database
	if err := s.repo.UpdateTask(task); err != nil {
		log.Printf("Failed to update next_run for task %s: %v", task.ID, err)
	}

	log.Printf("Scheduled one-off task: %s at %s", task.Name, scheduledTime)
	return nil
}

func (s *Scheduler) runOneOffWorker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkOneOffTasks()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Scheduler) checkOneOffTasks() {
	tasks, err := s.repo.GetScheduledTasks()
	if err != nil {
		log.Printf("Failed to get scheduled tasks: %v", err)
		return
	}

	now := time.Now()

	for _, task := range tasks {
		if task.Trigger.Type == models.TriggerOneOff &&
			task.NextRun != nil &&
			task.NextRun.Before(now) {

			// Execute task
			go s.executor.ExecuteTask(&task)

			// Mark task as completed
			task.Status = models.StatusCompleted
			task.UpdatedAt = time.Now()
			task.NextRun = nil

			if err := s.repo.UpdateTask(&task); err != nil {
				log.Printf("Failed to update task status: %v", err)
			}

			// Remove from jobs map
			s.mu.Lock()
			delete(s.jobs, task.ID)
			s.mu.Unlock()
		}
	}
}

func (s *Scheduler) RemoveTask(taskID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.jobs[taskID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, taskID)
		log.Printf("Removed task from scheduler: %s", taskID)
	}
}