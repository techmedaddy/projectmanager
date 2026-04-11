package projects

import (
	"time"

	"taskflow/backend/internal/tasks"
)

// CreateRequest is the request body for POST /projects.
type CreateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// UpdateRequest is the request body for PATCH /projects/:id.
type UpdateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// Response is the public JSON representation of a project.
type Response struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListResponse is the response body for GET /projects.
type ListResponse struct {
	Projects []Response `json:"projects"`
}

// DetailResponse is the response body for GET /projects/:id.
type DetailResponse struct {
	Project Response         `json:"project"`
	Tasks   []tasks.Response `json:"tasks"`
}

// StatsResponse is the response body for GET /projects/:id/stats.
type StatsResponse struct {
	ByStatus   map[string]int `json:"by_status"`
	ByAssignee map[string]int `json:"by_assignee"`
}

// ToResponse converts a repository project model into its API DTO.
func ToResponse(project Project) Response {
	return Response{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt,
	}
}

// ToResponses converts repository project models into API DTOs.
func ToResponses(projects []Project) []Response {
	items := make([]Response, 0, len(projects))
	for _, project := range projects {
		items = append(items, ToResponse(project))
	}

	return items
}
