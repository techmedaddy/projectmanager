package main

import (
	"net/http"

	"taskflow/backend/internal/auth"
	"taskflow/backend/internal/response"
)

func currentUserOr401(w http.ResponseWriter, r *http.Request) (auth.CurrentUser, bool) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.Unauthenticated(w)
		return auth.CurrentUser{}, false
	}

	return currentUser, true
}
