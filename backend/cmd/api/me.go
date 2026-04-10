package main

import (
	"net/http"

	"taskflow/backend/internal/auth"
	"taskflow/backend/internal/response"
)

func (app *application) meHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	currentUser, ok := currentUserOr401(w, r)
	if !ok {
		return
	}

	response.JSON(w, http.StatusOK, auth.CurrentUserResponse{
		User: currentUser,
	})
}
