package main

import (
	"errors"
	"net/http"
	"strings"

	"taskflow/backend/internal/auth"
	"taskflow/backend/internal/response"
	"taskflow/backend/internal/users"
)

const minPasswordLength = 8

func (app *application) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	var req auth.RegisterRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	fields := response.NewFieldErrors()
	name := strings.TrimSpace(req.Name)
	email := auth.NormalizeEmail(req.Email)

	if name == "" {
		fields.Add("name", "is required")
	}

	if email == "" {
		fields.Add("email", "is required")
	} else if !auth.IsValidEmail(email) {
		fields.Add("email", "must be a valid email")
	}

	if req.Password == "" {
		fields.Add("password", "is required")
	} else if len(req.Password) < minPasswordLength {
		fields.Add("password", "must be at least 8 characters")
	}

	if fields.HasAny() {
		response.Validation(w, fields)
		return
	}

	user, err := app.authService.Register(r.Context(), name, email, req.Password)
	switch {
	case errors.Is(err, auth.ErrEmailTaken):
		response.Validation(w, response.FieldErrors{
			"email": "is already in use",
		})
		return
	case err != nil:
		app.logger.Error(
			"register user failed",
			"request_id", requestIDFromContext(r.Context()),
			"error", err.Error(),
		)
		response.Internal(w)
		return
	}

	response.JSON(w, http.StatusCreated, auth.RegisterResponse{
		User: users.ToResponse(user),
	})
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	var req auth.LoginRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	fields := response.NewFieldErrors()
	email := auth.NormalizeEmail(req.Email)

	if email == "" {
		fields.Add("email", "is required")
	} else if !auth.IsValidEmail(email) {
		fields.Add("email", "must be a valid email")
	}

	if req.Password == "" {
		fields.Add("password", "is required")
	}

	if fields.HasAny() {
		response.Validation(w, fields)
		return
	}

	accessToken, err := app.authService.Login(r.Context(), email, req.Password)
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		response.Unauthenticated(w)
		return
	case err != nil:
		app.logger.Error(
			"login failed",
			"request_id", requestIDFromContext(r.Context()),
			"error", err.Error(),
		)
		response.Internal(w)
		return
	}

	response.JSON(w, http.StatusOK, auth.LoginResponse{
		AccessToken: accessToken,
	})
}
