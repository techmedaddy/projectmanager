package main

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"taskflow/backend/internal/projects"
	"taskflow/backend/internal/response"
	"taskflow/backend/internal/tasks"
)

var uuidPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

func (app *application) projectsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.listProjectsHandler(w, r)
	case http.MethodPost:
		app.createProjectHandler(w, r)
	default:
		response.MethodNotAllowed(w)
	}
}

func (app *application) projectByIDHandler(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	if !isUUID(projectID) {
		response.BadRequest(w, "invalid project id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.getProjectHandler(w, r, projectID)
	case http.MethodPatch:
		app.updateProjectHandler(w, r, projectID)
	case http.MethodDelete:
		app.deleteProjectHandler(w, r, projectID)
	default:
		response.MethodNotAllowed(w)
	}
}

func (app *application) projectStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	projectID := r.PathValue("id")
	if !isUUID(projectID) {
		response.BadRequest(w, "invalid project id")
		return
	}

	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	stats, err := app.projectsService.GetStats(r.Context(), projectID, currentUser.ID)
	if err != nil {
		app.handleProjectServiceError(w, r, err, "get project stats failed")
		return
	}

	response.JSON(w, http.StatusOK, projects.StatsResponse{
		ByStatus:   stats.ByStatus,
		ByAssignee: stats.ByAssignee,
	})
}

func (app *application) listProjectsHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	pagination, fields := parsePaginationParams(r)
	if fields.HasAny() {
		response.Validation(w, fields)
		return
	}

	projectList, err := app.projectsService.ListAccessible(r.Context(), currentUser.ID)
	if err != nil {
		app.logRequestError(r, "list projects failed", err)
		response.Internal(w)
		return
	}

	total := len(projectList)
	start, end := paginateBounds(total, pagination)
	pagedProjects := projectList[start:end]

	response.JSON(w, http.StatusOK, projects.ListResponse{
		Projects: projects.ToResponses(pagedProjects),
		Meta: &projects.PaginationMeta{
			Page:  pagination.Page,
			Limit: pagination.Limit,
			Total: total,
		},
	})
}

func (app *application) createProjectHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	var req projects.CreateRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	fields := response.NewFieldErrors()
	name := strings.TrimSpace(req.Name)
	if name == "" {
		fields.Add("name", "is required")
	}

	if fields.HasAny() {
		response.Validation(w, fields)
		return
	}

	project, err := app.projectsService.Create(r.Context(), currentUser.ID, name, req.Description)
	if err != nil {
		app.logRequestError(r, "create project failed", err)
		response.Internal(w)
		return
	}

	response.JSON(w, http.StatusCreated, projects.ToResponse(project))
}

func (app *application) getProjectHandler(w http.ResponseWriter, r *http.Request, projectID string) {
	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	project, projectTasks, err := app.projectsService.GetDetail(r.Context(), projectID, currentUser.ID)
	if err != nil {
		app.handleProjectServiceError(w, r, err, "get project failed")
		return
	}

	response.JSON(w, http.StatusOK, projects.DetailResponse{
		Project: projects.ToResponse(project),
		Tasks:   tasks.ToResponses(projectTasks),
	})
}

func (app *application) updateProjectHandler(w http.ResponseWriter, r *http.Request, projectID string) {
	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	var req projects.UpdateRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	fields := response.NewFieldErrors()
	name := strings.TrimSpace(req.Name)
	if name == "" {
		fields.Add("name", "is required")
	}

	if fields.HasAny() {
		response.Validation(w, fields)
		return
	}

	project, err := app.projectsService.Update(r.Context(), projectID, currentUser.ID, name, req.Description)
	if err != nil {
		app.handleProjectServiceError(w, r, err, "update project failed")
		return
	}

	response.JSON(w, http.StatusOK, projects.ToResponse(project))
}

func (app *application) deleteProjectHandler(w http.ResponseWriter, r *http.Request, projectID string) {
	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	if err := app.projectsService.Delete(r.Context(), projectID, currentUser.ID); err != nil {
		app.handleProjectServiceError(w, r, err, "delete project failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) handleProjectServiceError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, projects.ErrProjectNotFound):
		response.NotFound(w)
	case errors.Is(err, projects.ErrProjectForbidden):
		response.Forbidden(w)
	default:
		app.logRequestError(r, message, err)
		response.Internal(w)
	}
}

func isUUID(value string) bool {
	return uuidPattern.MatchString(value)
}
