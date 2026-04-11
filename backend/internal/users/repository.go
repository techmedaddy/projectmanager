package users

import (
	"context"
	"fmt"
	"time"

	"taskflow/backend/internal/db"
)

// User represents an application user record.
type User struct {
	ID        string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

// CreateParams contains the fields required to insert a user.
type CreateParams struct {
	Name     string
	Email    string
	Password string
}

// Repository provides explicit user queries backed by PostgreSQL.
type Repository struct {
	q db.Querier
}

// NewRepository builds a user repository from a pgx-backed query interface.
func NewRepository(q db.Querier) *Repository {
	return &Repository{q: q}
}

// Create inserts a user and returns the stored row.
func (r *Repository) Create(ctx context.Context, params CreateParams) (User, error) {
	const query = `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id::text, name, email, password, created_at
	`

	var user User
	err := r.q.QueryRow(ctx, query, params.Name, params.Email, params.Password).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// GetByEmail fetches a single user by email address.
func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT id::text, name, email, password, created_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.q.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

// GetByID fetches a single user by primary key.
func (r *Repository) GetByID(ctx context.Context, id string) (User, error) {
	const query = `
		SELECT id::text, name, email, password, created_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.q.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

// ListProjectAssignees returns unique users involved in project tasks (creator
// or assignee) plus the project owner.
func (r *Repository) ListProjectAssignees(ctx context.Context, projectID string) ([]User, error) {
	const query = `
		SELECT DISTINCT u.id::text, u.name, u.email, u.password, u.created_at
		FROM users u
		JOIN (
			SELECT p.owner_id AS user_id
			FROM projects p
			WHERE p.id = $1
			UNION
			SELECT t.assignee_id AS user_id
			FROM tasks t
			WHERE t.project_id = $1
			  AND t.assignee_id IS NOT NULL
			UNION
			SELECT t.creator_id AS user_id
			FROM tasks t
			WHERE t.project_id = $1
		) involved ON involved.user_id = u.id
		ORDER BY u.name ASC, u.created_at ASC
	`

	rows, err := r.q.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project assignees: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		if scanErr := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt); scanErr != nil {
			return nil, fmt.Errorf("scan project assignee: %w", scanErr)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project assignees: %w", err)
	}

	return users, nil
}
