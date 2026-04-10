package auth

import "taskflow/backend/internal/users"

// RegisterRequest is the request body for POST /auth/register.
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse is the response body for a successful registration.
type RegisterResponse struct {
	User users.Response `json:"user"`
}

// LoginRequest is the request body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the response body for a successful login.
type LoginResponse struct {
	AccessToken string `json:"access_token"`
}
