package main

import (
	"log/slog"

	"taskflow/backend/internal/auth"
	"taskflow/backend/internal/projects"
)

type application struct {
	logger      *slog.Logger
	authService *auth.Service
	projectsService *projects.Service
}
