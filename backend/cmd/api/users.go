package main

import (
	"errors"
	"net/http"

	"taskflow/backend/internal/projects"
	"taskflow/backend/internal/response"
	"taskflow/backend/internal/users"
)

func (app *application) projectAssigneesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	projectID := r.PathValue("id")
	if !isUUID(projectID) {
		response.BadRequest(w, "invalid project id")
		return
	}

	if _, _, err := app.projectsService.GetDetail(r.Context(), projectID, currentUser.ID); err != nil {
		switch {
		case errors.Is(err, projects.ErrProjectNotFound):
			response.NotFound(w)
		case errors.Is(err, projects.ErrProjectForbidden):
			response.Forbidden(w)
		default:
			app.logRequestError(r, "authorize project assignees failed", err)
			response.Internal(w)
		}
		return
	}

	assignees, err := app.usersService.ListProjectAssignees(r.Context(), projectID)
	if err != nil {
		app.logRequestError(r, "list project assignees failed", err)
		response.Internal(w)
		return
	}

	response.JSON(w, http.StatusOK, users.AssigneesResponse{Users: users.ToResponses(assignees)})
}
