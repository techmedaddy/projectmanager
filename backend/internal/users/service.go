package users

import (
	"context"
	"fmt"
)

type assigneeRepository interface {
	ListProjectAssignees(ctx context.Context, projectID string) ([]User, error)
}

// Service contains user-read business logic used by API handlers.
type Service struct {
	repo assigneeRepository
}

// NewService constructs a user service from explicit repository dependencies.
func NewService(repo assigneeRepository) *Service {
	return &Service{repo: repo}
}

// ListProjectAssignees returns unique users involved in project tasks plus the
// project owner, ordered by name.
func (s *Service) ListProjectAssignees(ctx context.Context, projectID string) ([]User, error) {
	users, err := s.repo.ListProjectAssignees(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project assignees: %w", err)
	}

	return users, nil
}
