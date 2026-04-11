package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"taskflow/backend/internal/response"
	"taskflow/backend/internal/tasks"
)

func (app *application) tasksByProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	if !isUUID(projectID) {
		response.BadRequest(w, "invalid project id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.listTasksHandler(w, r, projectID)
	case http.MethodPost:
		app.createTaskHandler(w, r, projectID)
	default:
		response.MethodNotAllowed(w)
	}
}

func (app *application) taskByIDHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	if !isUUID(taskID) {
		response.BadRequest(w, "invalid task id")
		return
	}

	switch r.Method {
	case http.MethodPatch:
		app.updateTaskHandler(w, r, taskID)
	case http.MethodDelete:
		app.deleteTaskHandler(w, r, taskID)
	default:
		response.MethodNotAllowed(w)
	}
}

func (app *application) listTasksHandler(w http.ResponseWriter, r *http.Request, projectID string) {
	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	pagination, paginationFields := parsePaginationParams(r)
	if paginationFields.HasAny() {
		response.Validation(w, paginationFields)
		return
	}

	filters, validationFields := buildTaskListFilters(r)
	if validationFields.HasAny() {
		response.Validation(w, validationFields)
		return
	}

	projectTasks, err := app.tasksService.ListByProject(r.Context(), projectID, currentUser.ID, filters)
	if err != nil {
		app.handleTaskServiceError(w, r, err, "list tasks failed")
		return
	}

	total := len(projectTasks)
	start, end := paginateBounds(total, pagination)
	pagedTasks := projectTasks[start:end]

	response.JSON(w, http.StatusOK, tasks.ListResponse{
		Tasks: tasks.ToResponses(pagedTasks),
		Meta: &tasks.PaginationMeta{
			Page:  pagination.Page,
			Limit: pagination.Limit,
			Total: total,
		},
	})
}

func (app *application) createTaskHandler(w http.ResponseWriter, r *http.Request, projectID string) {
	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	var req tasks.CreateRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	input, validationFields := buildTaskCreateInput(req)
	if validationFields.HasAny() {
		response.Validation(w, validationFields)
		return
	}

	task, err := app.tasksService.Create(r.Context(), projectID, currentUser.ID, input)
	if err != nil {
		app.handleTaskServiceError(w, r, err, "create task failed")
		return
	}

	response.JSON(w, http.StatusCreated, tasks.ToResponse(task))
}

func (app *application) updateTaskHandler(w http.ResponseWriter, r *http.Request, taskID string) {
	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	var req tasks.UpdateRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	input, validationFields := buildTaskUpdateInput(req)
	if validationFields.HasAny() {
		response.Validation(w, validationFields)
		return
	}

	task, err := app.tasksService.Update(r.Context(), taskID, currentUser.ID, input)
	if err != nil {
		app.handleTaskServiceError(w, r, err, "update task failed")
		return
	}

	response.JSON(w, http.StatusOK, tasks.ToResponse(task))
}

func (app *application) deleteTaskHandler(w http.ResponseWriter, r *http.Request, taskID string) {
	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	if err := app.tasksService.Delete(r.Context(), taskID, currentUser.ID); err != nil {
		app.handleTaskServiceError(w, r, err, "delete task failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) handleTaskServiceError(w http.ResponseWriter, r *http.Request, err error, message string) {
	switch {
	case errors.Is(err, tasks.ErrTaskNotFound), errors.Is(err, tasks.ErrTaskProjectNotFound):
		response.NotFound(w)
	case errors.Is(err, tasks.ErrTaskForbidden), errors.Is(err, tasks.ErrTaskProjectForbidden):
		response.Forbidden(w)
	case errors.Is(err, tasks.ErrAssigneeNotFound):
		response.Validation(w, response.FieldErrors{
			"assignee_id": "must reference an existing user",
		})
	default:
		app.logRequestError(r, message, err)
		response.Internal(w)
	}
}

func buildTaskListFilters(r *http.Request) (tasks.ListFilters, response.FieldErrors) {
	fields := response.NewFieldErrors()
	query := r.URL.Query()
	filters := tasks.ListFilters{}

	if rawStatus := strings.TrimSpace(query.Get("status")); rawStatus != "" {
		status := tasks.Status(rawStatus)
		if !tasks.IsValidStatus(status) {
			fields.Add("status", "must be one of todo, in_progress, done")
		} else {
			filters.Status = &status
		}
	}

	if rawAssignee := strings.TrimSpace(query.Get("assignee")); rawAssignee != "" {
		if !isUUID(rawAssignee) {
			fields.Add("assignee", "must be a valid uuid")
		} else {
			filters.AssigneeID = &rawAssignee
		}
	}

	return filters, fields
}

func buildTaskCreateInput(req tasks.CreateRequest) (tasks.CreateInput, response.FieldErrors) {
	fields := response.NewFieldErrors()
	title := strings.TrimSpace(req.Title)

	if title == "" {
		fields.Add("title", "is required")
	}

	status := tasks.StatusTodo
	if req.Status != nil {
		status = *req.Status
	}
	if !tasks.IsValidStatus(status) {
		fields.Add("status", "must be one of todo, in_progress, done")
	}

	priority := tasks.PriorityMedium
	if req.Priority != nil {
		priority = *req.Priority
	}
	if !tasks.IsValidPriority(priority) {
		fields.Add("priority", "must be one of low, medium, high")
	}

	assigneeID := normalizeNullableUUID(req.AssigneeID, "assignee_id", fields)
	dueDate := parseNullableDate(req.DueDate, "due_date", fields)

	return tasks.CreateInput{
		Title:       title,
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		AssigneeID:  assigneeID,
		DueDate:     dueDate,
	}, fields
}

func buildTaskUpdateInput(req tasks.UpdateRequest) (tasks.UpdateInput, response.FieldErrors) {
	fields := response.NewFieldErrors()

	if !req.HasAnyField() {
		fields.Add("body", "must include at least one updatable field")
		return tasks.UpdateInput{}, fields
	}

	if req.Title.Set {
		if req.Title.Value == nil {
			fields.Add("title", "cannot be null")
		} else if strings.TrimSpace(*req.Title.Value) == "" {
			fields.Add("title", "is required")
		}
	}

	if req.Status.Set {
		if req.Status.Value == nil {
			fields.Add("status", "cannot be null")
		} else if !tasks.IsValidStatus(*req.Status.Value) {
			fields.Add("status", "must be one of todo, in_progress, done")
		}
	}

	if req.Priority.Set {
		if req.Priority.Value == nil {
			fields.Add("priority", "cannot be null")
		} else if !tasks.IsValidPriority(*req.Priority.Value) {
			fields.Add("priority", "must be one of low, medium, high")
		}
	}

	assigneePatch := req.AssigneeID
	if assigneePatch.Set && assigneePatch.Value != nil {
		trimmed := strings.TrimSpace(*assigneePatch.Value)
		if trimmed == "" || !isUUID(trimmed) {
			fields.Add("assignee_id", "must be a valid uuid")
		} else {
			assigneePatch.Value = &trimmed
		}
	}

	dueDatePatch := req.DueDate
	if dueDatePatch.Set && dueDatePatch.Value != nil {
		trimmed := strings.TrimSpace(*dueDatePatch.Value)
		if _, err := time.Parse("2006-01-02", trimmed); err != nil {
			fields.Add("due_date", "must be in YYYY-MM-DD format")
		} else {
			dueDatePatch.Value = &trimmed
		}
	}

	titlePatch := req.Title
	if titlePatch.Set && titlePatch.Value != nil {
		trimmed := strings.TrimSpace(*titlePatch.Value)
		titlePatch.Value = &trimmed
	}

	descriptionPatch := req.Description
	if descriptionPatch.Set && descriptionPatch.Value != nil {
		trimmed := strings.TrimSpace(*descriptionPatch.Value)
		descriptionPatch.Value = &trimmed
	}

	return tasks.UpdateInput{
		Title:       titlePatch,
		Description: descriptionPatch,
		Status:      req.Status,
		Priority:    req.Priority,
		AssigneeID:  assigneePatch,
		DueDate:     dueDatePatch,
	}, fields
}

func normalizeNullableUUID(value *string, field string, errors response.FieldErrors) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" || !isUUID(trimmed) {
		errors.Add(field, "must be a valid uuid")
		return value
	}

	return &trimmed
}

func parseNullableDate(value *string, field string, errors response.FieldErrors) *time.Time {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		errors.Add(field, "must be in YYYY-MM-DD format")
		return nil
	}

	return &parsed
}
