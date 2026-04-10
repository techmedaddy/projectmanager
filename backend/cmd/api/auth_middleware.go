package main

import (
	"errors"
	"net/http"
	"strings"

	"taskflow/backend/internal/auth"
	"taskflow/backend/internal/response"
)

func (app *application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerTokenFromHeader(r.Header.Get("Authorization"))
		if !ok {
			response.Unauthenticated(w)
			return
		}

		currentUser, err := app.authService.Authenticate(r.Context(), token)
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			response.Unauthenticated(w)
			return
		case err != nil:
			app.logger.Error(
				"authenticate request failed",
				"request_id", requestIDFromContext(r.Context()),
				"error", err.Error(),
			)
			response.Internal(w)
			return
		}

		ctx := auth.WithCurrentUser(r.Context(), currentUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerTokenFromHeader(headerValue string) (string, bool) {
	parts := strings.Fields(strings.TrimSpace(headerValue))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}
