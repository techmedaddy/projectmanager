package main

import (
	"log/slog"

	"taskflow/backend/internal/auth"
	"taskflow/backend/internal/projects"
	"taskflow/backend/internal/tasks"
)

type application struct {
	logger          *slog.Logger
	authService     *auth.Service
	projectsService *projects.Service
	tasksService    *tasks.Service
}
