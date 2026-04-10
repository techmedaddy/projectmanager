package projects

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"taskflow/backend/internal/db"
)

// Project represents a project record.
type Project struct {
	ID          string
	Name        string
	Description *string
	OwnerID     string
	CreatedAt   time.Time
}

// CreateParams contains the fields required to insert a project.
type CreateParams struct {
	Name        string
	Description *string
	OwnerID     string
}

// UpdateParams contains the fields required to persist a project update.
type UpdateParams struct {
	ID          string
	Name        string
	Description *string
}

// Repository provides explicit project queries backed by PostgreSQL.
type Repository struct {
	q db.Querier
}

// NewRepository builds a project repository from a pgx-backed query interface.
func NewRepository(q db.Querier) *Repository {
	return &Repository{q: q}
}

// Create inserts a project and returns the stored row.
func (r *Repository) Create(ctx context.Context, params CreateParams) (Project, error) {
	const query = `
		INSERT INTO projects (name, description, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id::text, name, description, owner_id::text, created_at
	`

	project, err := scanProject(
		r.q.QueryRow(ctx, query, params.Name, params.Description, params.OwnerID),
	)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}

	return project, nil
}

// GetByID fetches a project by primary key.
func (r *Repository) GetByID(ctx context.Context, id string) (Project, error) {
	const query = `
		SELECT id::text, name, description, owner_id::text, created_at
		FROM projects
		WHERE id = $1
	`

	project, err := scanProject(r.q.QueryRow(ctx, query, id))
	if err != nil {
		return Project{}, fmt.Errorf("get project by id: %w", err)
	}

	return project, nil
}

// ListAccessibleByUser returns projects the user owns or participates in via
// assigned or created tasks.
func (r *Repository) ListAccessibleByUser(ctx context.Context, userID string) ([]Project, error) {
	const query = `
		SELECT DISTINCT p.id::text, p.name, p.description, p.owner_id::text, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1
		   OR t.assignee_id = $1
		   OR t.creator_id = $1
		ORDER BY p.created_at DESC
	`

	rows, err := r.q.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list accessible projects: %w", err)
	}
	defer rows.Close()

	projects := make([]Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("scan accessible project: %w", err)
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessible projects: %w", err)
	}

	return projects, nil
}

// Update persists a full project update and returns the stored row.
func (r *Repository) Update(ctx context.Context, params UpdateParams) (Project, error) {
	const query = `
		UPDATE projects
		SET name = $2,
		    description = $3
		WHERE id = $1
		RETURNING id::text, name, description, owner_id::text, created_at
	`

	project, err := scanProject(r.q.QueryRow(ctx, query, params.ID, params.Name, params.Description))
	if err != nil {
		return Project{}, fmt.Errorf("update project: %w", err)
	}

	return project, nil
}

// Delete removes a project by primary key.
func (r *Repository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM projects WHERE id = $1`

	if _, err := r.q.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	return nil
}

type projectScanner interface {
	Scan(dest ...any) error
}

func scanProject(scanner projectScanner) (Project, error) {
	var (
		project     Project
		description pgtype.Text
	)

	err := scanner.Scan(
		&project.ID,
		&project.Name,
		&description,
		&project.OwnerID,
		&project.CreatedAt,
	)
	if err != nil {
		return Project{}, err
	}

	if description.Valid {
		project.Description = &description.String
	}

	return project, nil
}
