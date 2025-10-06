package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	StatusScheduled TaskStatus = "scheduled"
	StatusCancelled TaskStatus = "cancelled"
	StatusCompleted TaskStatus = "completed"
)

type TriggerType string

const (
	TriggerOneOff TriggerType = "one-off"
	TriggerCron   TriggerType = "cron"
)

type Trigger struct {
	Type     TriggerType `json:"type" binding:"required,oneof=one-off cron"`
	DateTime *time.Time  `json:"datetime,omitempty"`
	Cron     *string     `json:"cron,omitempty"`
}

func (t *Trigger) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal Trigger value")
	}
	return json.Unmarshal(bytes, t)
}

func (t Trigger) Value() (driver.Value, error) {
	return json.Marshal(t)
}

type Action struct {
	Method  string            `json:"method" binding:"required"`
	URL     string            `json:"url" binding:"required,url"`
	Headers map[string]string `json:"headers,omitempty"`
	Payload json.RawMessage   `json:"payload,omitempty"`
}

func (a *Action) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal Action value")
	}
	return json.Unmarshal(bytes, a)
}

func (a Action) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type Task struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" binding:"required" db:"name"`
	Trigger   Trigger    `json:"trigger" binding:"required" db:"trigger"`
	Action    Action     `json:"action" binding:"required" db:"action"`
	Status    TaskStatus `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	NextRun   *time.Time `json:"next_run,omitempty" db:"next_run"`
}

type CreateTaskRequest struct {
	Name    string  `json:"name" binding:"required"`
	Trigger Trigger `json:"trigger" binding:"required"`
	Action  Action  `json:"action" binding:"required"`
}

type UpdateTaskRequest struct {
	Name    *string     `json:"name,omitempty"`
	Trigger *Trigger    `json:"trigger,omitempty"`
	Action  *Action     `json:"action,omitempty"`
	Status  *TaskStatus `json:"status,omitempty"`
}

type ListTasksParams struct {
	Page   int        `form:"page"`
	Limit  int        `form:"limit" binding:"max=100"`
	Status TaskStatus `form:"status"`
}
