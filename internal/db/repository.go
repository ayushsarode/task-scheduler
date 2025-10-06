package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ayushsarode/task-scheduler/internal/models"
)

type Repository struct {
	db *DB
}

func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

// Task Repository Methods

func (r *Repository) CreateTask(task *models.Task) error {
	query := `
		INSERT INTO tasks (id, name, trigger, action, status, created_at, updated_at, next_run)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query,
		task.ID,
		task.Name,
		task.Trigger,
		task.Action,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
		task.NextRun,
	)
	return err
}

func (r *Repository) GetTaskByID(id uuid.UUID) (*models.Task, error) {
	task := &models.Task{}
	query := `
		SELECT id, name, trigger, action, status, created_at, updated_at, next_run
		FROM tasks
		WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.Name,
		&task.Trigger,
		&task.Action,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.NextRun,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	return task, err
}

func (r *Repository) ListTasks(params models.ListTasksParams) ([]models.Task, int, error) {
	// Default pagination
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = 10
	}

	offset := (params.Page - 1) * params.Limit

	// Build query with filters
	conditions := []string{}
	args := []interface{}{}
	argCount := 1

	if params.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, params.Status)
		argCount++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, name, trigger, action, status, created_at, updated_at, next_run
		FROM tasks
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Trigger,
			&task.Action,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.NextRun,
		)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

func (r *Repository) UpdateTask(task *models.Task) error {
	query := `
		UPDATE tasks
		SET name = $1, trigger = $2, action = $3, status = $4, updated_at = $5, next_run = $6
		WHERE id = $7
	`
	result, err := r.db.Exec(query,
		task.Name,
		task.Trigger,
		task.Action,
		task.Status,
		task.UpdatedAt,
		task.NextRun,
		task.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func (r *Repository) DeleteTask(id uuid.UUID) error {
	query := "UPDATE tasks SET status = $1, updated_at = $2 WHERE id = $3"
	result, err := r.db.Exec(query, models.StatusCancelled, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func (r *Repository) GetScheduledTasks() ([]models.Task, error) {
	query := `
		SELECT id, name, trigger, action, status, created_at, updated_at, next_run
		FROM tasks
		WHERE status = $1
		ORDER BY next_run ASC NULLS LAST
	`
	rows, err := r.db.Query(query, models.StatusScheduled)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Trigger,
			&task.Action,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.NextRun,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// TaskResult Repository Methods

func (r *Repository) CreateTaskResult(result *models.TaskResult) error {
	query := `
		INSERT INTO task_results (id, task_id, run_at, status_code, success, response_headers, response_body, error_message, duration_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	var responseHeaders interface{}
	if len(result.ResponseHeaders) > 0 && string(result.ResponseHeaders) != "null" {
		responseHeaders = string(result.ResponseHeaders)
	} else {
		responseHeaders = nil
	}

	_, err := r.db.Exec(query,
		result.ID,
		result.TaskID,
		result.RunAt,
		result.StatusCode,
		result.Success,
		responseHeaders,
		result.ResponseBody,
		result.ErrorMessage,
		result.DurationMs,
		result.CreatedAt,
	)
	return err
}

func (r *Repository) GetTaskResults(taskID uuid.UUID, page, limit int) ([]models.TaskResult, int, error) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM task_results WHERE task_id = $1"
	err := r.db.QueryRow(countQuery, taskID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, task_id, run_at, status_code, success, response_headers, response_body, error_message, duration_ms, created_at
		FROM task_results
		WHERE task_id = $1
		ORDER BY run_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, taskID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := []models.TaskResult{}
	for rows.Next() {
		var result models.TaskResult
		var responseHeaders sql.NullString
		err := rows.Scan(
			&result.ID,
			&result.TaskID,
			&result.RunAt,
			&result.StatusCode,
			&result.Success,
			&responseHeaders,
			&result.ResponseBody,
			&result.ErrorMessage,
			&result.DurationMs,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		// Handle nullable response headers
		if responseHeaders.Valid {
			result.ResponseHeaders = json.RawMessage(responseHeaders.String)
		} else {
			result.ResponseHeaders = json.RawMessage("null")
		}
		
		results = append(results, result)
	}

	return results, total, nil
}

func (r *Repository) ListAllResults(params models.ListResultsParams) ([]models.TaskResult, int, error) {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = 10
	}

	offset := (params.Page - 1) * params.Limit

	conditions := []string{}
	args := []interface{}{}
	argCount := 1

	if params.TaskID != nil {
		conditions = append(conditions, fmt.Sprintf("task_id = $%d", argCount))
		args = append(args, *params.TaskID)
		argCount++
	}

	if params.Success != nil {
		conditions = append(conditions, fmt.Sprintf("success = $%d", argCount))
		args = append(args, *params.Success)
		argCount++
	}

	if params.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("run_at >= $%d", argCount))
		args = append(args, *params.DateFrom)
		argCount++
	}

	if params.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("run_at <= $%d", argCount))
		args = append(args, *params.DateTo)
		argCount++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM task_results %s", whereClause)
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, task_id, run_at, status_code, success, response_headers, response_body, error_message, duration_ms, created_at
		FROM task_results
		%s
		ORDER BY run_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := []models.TaskResult{}
	for rows.Next() {
		var result models.TaskResult
		var responseHeaders sql.NullString
		err := rows.Scan(
			&result.ID,
			&result.TaskID,
			&result.RunAt,
			&result.StatusCode,
			&result.Success,
			&responseHeaders,
			&result.ResponseBody,
			&result.ErrorMessage,
			&result.DurationMs,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		// Handle nullable response headers
		if responseHeaders.Valid {
			result.ResponseHeaders = json.RawMessage(responseHeaders.String)
		} else {
			result.ResponseHeaders = json.RawMessage("null")
		}
		
		results = append(results, result)
	}

	return results, total, nil
}