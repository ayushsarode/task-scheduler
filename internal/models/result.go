package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type TaskResult struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	TaskID          uuid.UUID       `json:"task_id" db:"task_id"`
	RunAt           time.Time       `json:"run_at" db:"run_at"`
	StatusCode      int             `json:"status_code" db:"status_code"`
	Success         bool            `json:"success" db:"success"`
	ResponseHeaders json.RawMessage `json:"response_headers,omitempty" db:"response_headers"`
	ResponseBody    string          `json:"response_body,omitempty" db:"response_body"`
	ErrorMessage    *string         `json:"error_message,omitempty" db:"error_message"`
	DurationMs      int64           `json:"duration_ms" db:"duration_ms"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type ResponseHeaders map[string][]string

func (r *ResponseHeaders) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal ResponseHeaders value")
	}
	return json.Unmarshal(bytes, r)
}

func (r ResponseHeaders) Value() (driver.Value, error) {
	return json.Marshal(r)
}

type ListResultsParams struct {
	Page     int        `form:"page"`
	Limit    int        `form:"limit" binding:"max=100"`
	TaskID   *uuid.UUID `form:"task_id"`
	Success  *bool      `form:"success"`
	DateFrom *time.Time `form:"date_from"`
	DateTo   *time.Time `form:"date_to"`
}
