package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ayushsarode/task-scheduler/internal/db"
	"github.com/ayushsarode/task-scheduler/internal/models"
)

type Executor struct {
	repo   *db.Repository
	client *http.Client
}

func NewExecutor(repo *db.Repository) *Executor {
	return &Executor{
		repo: repo,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (e *Executor) ExecuteTask(task *models.Task) {
	log.Printf("Executing task: %s (ID: %s)", task.Name, task.ID)

	startTime := time.Now()
	result := &models.TaskResult{
		ID:        uuid.New(),
		TaskID:    task.ID,
		RunAt:     startTime,
		CreatedAt: time.Now(),
	}

	// Prepare HTTP request
	var reqBody io.Reader
	if task.Action.Payload != nil && len(task.Action.Payload) > 0 {
		reqBody = bytes.NewReader(task.Action.Payload)
	}

	req, err := http.NewRequest(task.Action.Method, task.Action.URL, reqBody)
	if err != nil {
		result.Success = false
		errorMsg := fmt.Sprintf("Failed to create request: %v", err)
		result.ErrorMessage = &errorMsg
		result.DurationMs = time.Since(startTime).Milliseconds()
		e.saveResult(result)
		return
	}

	// Set headers
	if task.Action.Headers != nil {
		for key, value := range task.Action.Headers {
			req.Header.Set(key, value)
		}
	}

	// Set default Content-Type if payload exists
	if task.Action.Payload != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := e.client.Do(req)
	if err != nil {
		result.Success = false
		errorMsg := fmt.Sprintf("Request failed: %v", err)
		result.ErrorMessage = &errorMsg
		result.DurationMs = time.Since(startTime).Milliseconds()
		e.saveResult(result)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		body = []byte{}
	}

	// Calculate duration
	result.DurationMs = time.Since(startTime).Milliseconds()
	result.StatusCode = resp.StatusCode
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	result.ResponseBody = string(body)

	// Store response headers as JSON
	if resp.Header != nil {
		headersJSON, err := json.Marshal(resp.Header)
		if err != nil {
			log.Printf("Failed to marshal response headers: %v", err)
			result.ResponseHeaders = json.RawMessage("null")
		} else {
			result.ResponseHeaders = json.RawMessage(headersJSON)
		}
	} else {
		result.ResponseHeaders = json.RawMessage("null")
	}

	// Save result
	e.saveResult(result)

	log.Printf("Task executed: %s, Status: %d, Success: %v, Duration: %dms",
		task.Name, result.StatusCode, result.Success, result.DurationMs)
}

func (e *Executor) saveResult(result *models.TaskResult) {
	if err := e.repo.CreateTaskResult(result); err != nil {
		log.Printf("Failed to save task result: %v", err)
	}
}