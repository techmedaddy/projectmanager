package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"taskflow/backend/internal/db"
)

// Status represents the allowed task workflow states.
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// Priority represents the allowed task priorities.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Task represents a task record.
type Task struct {
	ID          string
	Title       string
	Description *string
	Status      Status
	Priority    Priority
	ProjectID   string
	AssigneeID  *string
	CreatorID   string
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateParams contains the fields required to insert a task.
type CreateParams struct {
	Title       string
	Description *string
	Status      Status
	Priority    Priority
	ProjectID   string
	AssigneeID  *string
	CreatorID   string
	DueDate     *time.Time
}

// UpdateParams contains the fields required to persist a task update.
type UpdateParams struct {
	ID          string
	Title       string
	Description *string
	Status      Status
	Priority    Priority
	AssigneeID  *string
	DueDate     *time.Time
}

// ListFilters contains supported task list filters.
type ListFilters struct {
	Status     *Status
	AssigneeID *string
}

// Repository provides explicit task queries backed by PostgreSQL.
type Repository struct {
	q db.Querier
}

// NewRepository builds a task repository from a pgx-backed query interface.
func NewRepository(q db.Querier) *Repository {
	return &Repository{q: q}
}

// Create inserts a task and returns the stored row.
func (r *Repository) Create(ctx context.Context, params CreateParams) (Task, error) {
	const query = `
		INSERT INTO tasks (
			title,
			description,
			status,
			priority,
			project_id,
			assignee_id,
			creator_id,
			due_date
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING
			id::text,
			title,
			description,
			status::text,
			priority::text,
			project_id::text,
			assignee_id::text,
			creator_id::text,
			due_date,
			created_at,
			updated_at
	`

	task, err := scanTask(
		r.q.QueryRow(
			ctx,
			query,
			params.Title,
			params.Description,
			params.Status,
			params.Priority,
			params.ProjectID,
			params.AssigneeID,
			params.CreatorID,
			params.DueDate,
		),
	)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

// GetByID fetches a task by primary key.
func (r *Repository) GetByID(ctx context.Context, id string) (Task, error) {
	const query = `
		SELECT
			id::text,
			title,
			description,
			status::text,
			priority::text,
			project_id::text,
			assignee_id::text,
			creator_id::text,
			due_date,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1
	`

	task, err := scanTask(r.q.QueryRow(ctx, query, id))
	if err != nil {
		return Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return task, nil
}

// ListByProject returns tasks for a project with optional status and assignee
// filters.
func (r *Repository) ListByProject(ctx context.Context, projectID string, filters ListFilters) ([]Task, error) {
	var (
		args  = []any{projectID}
		parts = []string{"project_id = $1"}
	)

	if filters.Status != nil {
		args = append(args, *filters.Status)
		parts = append(parts, fmt.Sprintf("status = $%d", len(args)))
	}

	if filters.AssigneeID != nil {
		args = append(args, *filters.AssigneeID)
		parts = append(parts, fmt.Sprintf("assignee_id = $%d", len(args)))
	}

	query := fmt.Sprintf(`
		SELECT
			id::text,
			title,
			description,
			status::text,
			priority::text,
			project_id::text,
			assignee_id::text,
			creator_id::text,
			due_date,
			created_at,
			updated_at
		FROM tasks
		WHERE %s
		ORDER BY created_at DESC
	`, strings.Join(parts, " AND "))

	rows, err := r.q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks by project: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

// Update persists a full task update and returns the stored row.
func (r *Repository) Update(ctx context.Context, params UpdateParams) (Task, error) {
	const query = `
		UPDATE tasks
		SET title = $2,
		    description = $3,
		    status = $4,
		    priority = $5,
		    assignee_id = $6,
		    due_date = $7
		WHERE id = $1
		RETURNING
			id::text,
			title,
			description,
			status::text,
			priority::text,
			project_id::text,
			assignee_id::text,
			creator_id::text,
			due_date,
			created_at,
			updated_at
	`

	task, err := scanTask(
		r.q.QueryRow(
			ctx,
			query,
			params.ID,
			params.Title,
			params.Description,
			params.Status,
			params.Priority,
			params.AssigneeID,
			params.DueDate,
		),
	)
	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

// Delete removes a task by primary key.
func (r *Repository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	if _, err := r.q.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (Task, error) {
	var (
		task        Task
		description pgtype.Text
		status      string
		priority    string
		assigneeID  pgtype.Text
		dueDate     pgtype.Date
	)

	err := scanner.Scan(
		&task.ID,
		&task.Title,
		&description,
		&status,
		&priority,
		&task.ProjectID,
		&assigneeID,
		&task.CreatorID,
		&dueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return Task{}, err
	}

	task.Status = Status(status)
	task.Priority = Priority(priority)

	if description.Valid {
		task.Description = &description.String
	}

	if assigneeID.Valid {
		task.AssigneeID = &assigneeID.String
	}

	if dueDate.Valid {
		parsedDate := dueDate.Time
		task.DueDate = &parsedDate
	}

	return task, nil
}
