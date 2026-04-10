package tasks

import "time"

// CreateRequest is the request body for POST /projects/:id/tasks.
type CreateRequest struct {
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Status      *Status   `json:"status"`
	Priority    *Priority `json:"priority"`
	AssigneeID  *string   `json:"assignee_id"`
	DueDate     *string   `json:"due_date"`
}

// UpdateRequest is the request body for PATCH /tasks/:id.
type UpdateRequest struct {
	Title       NullableStringPatch   `json:"title"`
	Description NullableStringPatch   `json:"description"`
	Status      NullableStatusPatch   `json:"status"`
	Priority    NullablePriorityPatch `json:"priority"`
	AssigneeID  NullableStringPatch   `json:"assignee_id"`
	DueDate     NullableStringPatch   `json:"due_date"`
}

// Response is the public JSON representation of a task.
type Response struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	ProjectID   string    `json:"project_id"`
	AssigneeID  *string   `json:"assignee_id,omitempty"`
	CreatorID   string    `json:"creator_id"`
	DueDate     *string   `json:"due_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListResponse is the response body for task list endpoints.
type ListResponse struct {
	Tasks []Response `json:"tasks"`
}

// ToResponse converts a repository task model into its API DTO.
func ToResponse(task Task) Response {
	return Response{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		ProjectID:   task.ProjectID,
		AssigneeID:  task.AssigneeID,
		CreatorID:   task.CreatorID,
		DueDate:     formatDate(task.DueDate),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

// ToResponses converts repository task models into API DTOs.
func ToResponses(tasks []Task) []Response {
	items := make([]Response, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, ToResponse(task))
	}

	return items
}

// HasAnyField reports whether the PATCH request included any updatable field.
func (r UpdateRequest) HasAnyField() bool {
	return r.Title.Set ||
		r.Description.Set ||
		r.Status.Set ||
		r.Priority.Set ||
		r.AssigneeID.Set ||
		r.DueDate.Set
}

func formatDate(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.Format("2006-01-02")
	return &formatted
}
