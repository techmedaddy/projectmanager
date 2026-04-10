package main

import (
	"log/slog"

	"taskflow/backend/internal/auth"
)

type application struct {
	logger      *slog.Logger
	authService *auth.Service
}
