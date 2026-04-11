package projects

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"taskflow/backend/internal/tasks"
)

var (
	// ErrProjectNotFound is returned when the requested project does not exist.
	ErrProjectNotFound = errors.New("project not found")
	// ErrProjectForbidden is returned when the current user is not allowed to
	// access or mutate the requested project.
	ErrProjectForbidden = errors.New("project forbidden")
)

type projectRepository interface {
	Create(ctx context.Context, params CreateParams) (Project, error)
	GetByID(ctx context.Context, id string) (Project, error)
	ListAccessibleByUser(ctx context.Context, userID string) ([]Project, error)
	Update(ctx context.Context, params UpdateParams) (Project, error)
	Delete(ctx context.Context, id string) error
	HasTaskInvolvement(ctx context.Context, projectID, userID string) (bool, error)
}

type taskRepository interface {
	ListByProject(ctx context.Context, projectID string, filters tasks.ListFilters) ([]tasks.Task, error)
}

// Service contains project business logic and centralizes project access rules.
//
// Intended visibility rule:
// - owner OR involved-in-task (creator or assignee).
type Service struct {
	projectsRepo projectRepository
	tasksRepo    taskRepository
}

// NewService constructs a project service from explicit repository
// dependencies.
func NewService(projectsRepo projectRepository, tasksRepo taskRepository) *Service {
	return &Service{
		projectsRepo: projectsRepo,
		tasksRepo:    tasksRepo,
	}
}

// ListAccessible returns projects visible to the current user.
//
// Visibility means owner OR involved-in-task (creator or assignee).
func (s *Service) ListAccessible(ctx context.Context, userID string) ([]Project, error) {
	projects, err := s.projectsRepo.ListAccessibleByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list accessible projects: %w", err)
	}

	return projects, nil
}

// Create creates a new project owned by the current user.
func (s *Service) Create(ctx context.Context, ownerID, name string, description *string) (Project, error) {
	project, err := s.projectsRepo.Create(ctx, CreateParams{
		Name:        strings.TrimSpace(name),
		Description: normalizeOptionalString(description),
		OwnerID:     ownerID,
	})
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}

	return project, nil
}

// GetDetail returns a project and its tasks when the current user is allowed to
// access it.
//
// Access follows the same visibility intent: owner OR involved-in-task
// (creator or assignee).
func (s *Service) GetDetail(ctx context.Context, projectID, userID string) (Project, []tasks.Task, error) {
	project, err := s.authorizeAccess(ctx, projectID, userID)
	if err != nil {
		return Project{}, nil, err
	}

	projectTasks, err := s.tasksRepo.ListByProject(ctx, projectID, tasks.ListFilters{})
	if err != nil {
		return Project{}, nil, fmt.Errorf("list project tasks: %w", err)
	}

	return project, projectTasks, nil
}

// Update modifies a project when the current user owns it.
func (s *Service) Update(ctx context.Context, projectID, userID, name string, description *string) (Project, error) {
	project, err := s.authorizeOwner(ctx, projectID, userID)
	if err != nil {
		return Project{}, err
	}

	updatedProject, err := s.projectsRepo.Update(ctx, UpdateParams{
		ID:          project.ID,
		Name:        strings.TrimSpace(name),
		Description: normalizeOptionalString(description),
	})
	if err != nil {
		return Project{}, fmt.Errorf("update project: %w", err)
	}

	return updatedProject, nil
}

// Delete removes a project when the current user owns it.
func (s *Service) Delete(ctx context.Context, projectID, userID string) error {
	project, err := s.authorizeOwner(ctx, projectID, userID)
	if err != nil {
		return err
	}

	if err := s.projectsRepo.Delete(ctx, project.ID); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	return nil
}

// authorizeAccess enforces project visibility checks.
//
// Intended rule: owner OR involved-in-task (creator or assignee).
func (s *Service) authorizeAccess(ctx context.Context, projectID, userID string) (Project, error) {
	project, err := s.projectsRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Project{}, ErrProjectNotFound
		}

		return Project{}, fmt.Errorf("get project: %w", err)
	}

	if project.OwnerID == userID {
		return project, nil
	}

	hasTaskInvolvement, err := s.projectsRepo.HasTaskInvolvement(ctx, projectID, userID)
	if err != nil {
		return Project{}, fmt.Errorf("check project access: %w", err)
	}

	if !hasTaskInvolvement {
		return Project{}, ErrProjectForbidden
	}

	return project, nil
}

func (s *Service) authorizeOwner(ctx context.Context, projectID, userID string) (Project, error) {
	project, err := s.projectsRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Project{}, ErrProjectNotFound
		}

		return Project{}, fmt.Errorf("get project: %w", err)
	}

	if project.OwnerID != userID {
		return Project{}, ErrProjectForbidden
	}

	return project, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
