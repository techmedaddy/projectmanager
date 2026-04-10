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
