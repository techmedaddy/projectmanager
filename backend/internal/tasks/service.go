package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"taskflow/backend/internal/users"
)

var (
	// ErrTaskNotFound is returned when the requested task does not exist.
	ErrTaskNotFound = errors.New("task not found")
	// ErrTaskForbidden is returned when the current user is not allowed to modify
	// the requested task.
	ErrTaskForbidden = errors.New("task forbidden")
	// ErrTaskProjectNotFound is returned when the requested project does not exist.
	ErrTaskProjectNotFound = errors.New("project not found")
	// ErrTaskProjectForbidden is returned when the current user cannot access the
	// requested project.
	ErrTaskProjectForbidden = errors.New("project forbidden")
	// ErrAssigneeNotFound is returned when an assignee id does not match a user.
	ErrAssigneeNotFound = errors.New("assignee not found")
)

type taskRepository interface {
	Create(ctx context.Context, params CreateParams) (Task, error)
	GetByID(ctx context.Context, id string) (Task, error)
	ListByProject(ctx context.Context, projectID string, filters ListFilters) ([]Task, error)
	Update(ctx context.Context, params UpdateParams) (Task, error)
	Delete(ctx context.Context, id string) error
}

type projectRepository interface {
	GetOwnerID(ctx context.Context, id string) (string, error)
	HasAssignedTask(ctx context.Context, projectID, userID string) (bool, error)
}

type userRepository interface {
	GetByID(ctx context.Context, id string) (users.User, error)
}

// CreateInput contains normalized task creation inputs.
type CreateInput struct {
	Title       string
	Description *string
	Status      Status
	Priority    Priority
	AssigneeID  *string
	DueDate     *time.Time
}

// UpdateInput contains normalized task patch inputs.
type UpdateInput struct {
	Title       NullableStringPatch
	Description NullableStringPatch
	Status      NullableStatusPatch
	Priority    NullablePriorityPatch
	AssigneeID  NullableStringPatch
	DueDate     NullableStringPatch
}

// Service contains task business logic and task-specific authorization rules.
type Service struct {
	tasksRepo    taskRepository
	projectsRepo projectRepository
	usersRepo    userRepository
}

// NewService constructs a task service from explicit repository dependencies.
func NewService(tasksRepo taskRepository, projectsRepo projectRepository, usersRepo userRepository) *Service {
	return &Service{
		tasksRepo:    tasksRepo,
		projectsRepo: projectsRepo,
		usersRepo:    usersRepo,
	}
}

// ListByProject returns tasks in a project when the current user can access the
// project.
func (s *Service) ListByProject(ctx context.Context, projectID, userID string, filters ListFilters) ([]Task, error) {
	if _, err := s.authorizeProjectAccess(ctx, projectID, userID); err != nil {
		return nil, err
	}

	projectTasks, err := s.tasksRepo.ListByProject(ctx, projectID, filters)
	if err != nil {
		return nil, fmt.Errorf("list tasks by project: %w", err)
	}

	return projectTasks, nil
}

// Create creates a task inside a project and records the current user as the
// creator.
func (s *Service) Create(ctx context.Context, projectID, userID string, input CreateInput) (Task, error) {
	if _, err := s.authorizeProjectAccess(ctx, projectID, userID); err != nil {
		return Task{}, err
	}

	if err := s.ensureAssigneeExists(ctx, input.AssigneeID); err != nil {
		return Task{}, err
	}

	task, err := s.tasksRepo.Create(ctx, CreateParams{
		Title:       strings.TrimSpace(input.Title),
		Description: normalizeOptionalString(input.Description),
		Status:      input.Status,
		Priority:    input.Priority,
		ProjectID:   projectID,
		AssigneeID:  normalizeOptionalString(input.AssigneeID),
		CreatorID:   userID,
		DueDate:     input.DueDate,
	})
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

// Update modifies a task when the current user is the project owner, task
// creator, or task assignee.
func (s *Service) Update(ctx context.Context, taskID, userID string, input UpdateInput) (Task, error) {
	existingTask, err := s.authorizeTaskUpdate(ctx, taskID, userID)
	if err != nil {
		return Task{}, err
	}

	params := UpdateParams{
		ID:          existingTask.ID,
		Title:       existingTask.Title,
		Description: existingTask.Description,
		Status:      existingTask.Status,
		Priority:    existingTask.Priority,
		AssigneeID:  existingTask.AssigneeID,
		DueDate:     existingTask.DueDate,
	}

	if input.Title.Set {
		if input.Title.Value == nil {
			params.Title = ""
		} else {
			params.Title = strings.TrimSpace(*input.Title.Value)
		}
	}

	if input.Description.Set {
		params.Description = normalizeOptionalString(input.Description.Value)
	}

	if input.Status.Set {
		if input.Status.Value == nil {
			params.Status = ""
		} else {
			params.Status = *input.Status.Value
		}
	}

	if input.Priority.Set {
		if input.Priority.Value == nil {
			params.Priority = ""
		} else {
			params.Priority = *input.Priority.Value
		}
	}

	if input.AssigneeID.Set {
		params.AssigneeID = normalizeOptionalString(input.AssigneeID.Value)
	}

	if input.DueDate.Set {
		params.DueDate = inputToDate(input.DueDate.Value)
	}

	if err := s.ensureAssigneeExists(ctx, params.AssigneeID); err != nil {
		return Task{}, err
	}

	task, err := s.tasksRepo.Update(ctx, params)
	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

// Delete removes a task when the current user is the project owner or task
// creator.
func (s *Service) Delete(ctx context.Context, taskID, userID string) error {
	task, ownerID, err := s.getTaskWithProject(ctx, taskID)
	if err != nil {
		return err
	}

	if ownerID != userID && task.CreatorID != userID {
		return ErrTaskForbidden
	}

	if err := s.tasksRepo.Delete(ctx, task.ID); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (s *Service) authorizeProjectAccess(ctx context.Context, projectID, userID string) (string, error) {
	ownerID, err := s.projectsRepo.GetOwnerID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrTaskProjectNotFound
		}

		return "", fmt.Errorf("get project owner: %w", err)
	}

	if ownerID == userID {
		return ownerID, nil
	}

	hasAssignedTask, err := s.projectsRepo.HasAssignedTask(ctx, projectID, userID)
	if err != nil {
		return "", fmt.Errorf("check project access: %w", err)
	}

	if !hasAssignedTask {
		return "", ErrTaskProjectForbidden
	}

	return ownerID, nil
}

func (s *Service) authorizeTaskUpdate(ctx context.Context, taskID, userID string) (Task, error) {
	task, ownerID, err := s.getTaskWithProject(ctx, taskID)
	if err != nil {
		return Task{}, err
	}

	if ownerID == userID {
		return task, nil
	}

	hasProjectAccess, err := s.projectsRepo.HasAssignedTask(ctx, task.ProjectID, userID)
	if err != nil {
		return Task{}, fmt.Errorf("check task project access: %w", err)
	}

	if !hasProjectAccess {
		return Task{}, ErrTaskProjectForbidden
	}

	if task.CreatorID == userID {
		return task, nil
	}

	if task.AssigneeID != nil && *task.AssigneeID == userID {
		return task, nil
	}

	return Task{}, ErrTaskForbidden
}

func (s *Service) getTaskWithProject(ctx context.Context, taskID string) (Task, string, error) {
	task, err := s.tasksRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, "", ErrTaskNotFound
		}

		return Task{}, "", fmt.Errorf("get task: %w", err)
	}

	ownerID, err := s.projectsRepo.GetOwnerID(ctx, task.ProjectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, "", ErrTaskProjectNotFound
		}

		return Task{}, "", fmt.Errorf("get task project owner: %w", err)
	}

	return task, ownerID, nil
}

func (s *Service) ensureAssigneeExists(ctx context.Context, assigneeID *string) error {
	if assigneeID == nil {
		return nil
	}

	_, err := s.usersRepo.GetByID(ctx, *assigneeID)
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAssigneeNotFound
	}

	return fmt.Errorf("get assignee: %w", err)
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

func inputToDate(value *string) *time.Time {
	if value == nil {
		return nil
	}

	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*value))
	if err != nil {
		return nil
	}

	return &parsed
}
