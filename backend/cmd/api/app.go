package main

import (
	"log/slog"

	"taskflow/backend/internal/auth"
	"taskflow/backend/internal/config"
	"taskflow/backend/internal/db"
	"taskflow/backend/internal/projects"
	"taskflow/backend/internal/tasks"
	"taskflow/backend/internal/users"
)

type application struct {
	logger          *slog.Logger
	authService     *auth.Service
	projectsService *projects.Service
	tasksService    *tasks.Service
	usersService    *users.Service
}

func newApplication(logger *slog.Logger, cfg config.Config, dbConn *db.Database) *application {
	usersRepo := users.NewRepository(dbConn.Querier())
	projectsRepo := projects.NewRepository(dbConn.Querier())
	tasksRepo := tasks.NewRepository(dbConn.Querier())

	return &application{
		logger:          logger,
		authService:     auth.NewService(usersRepo, cfg.JWTSecret, cfg.JWTExpiryHours, cfg.BcryptCost),
		projectsService: projects.NewService(projectsRepo, tasksRepo),
		tasksService:    tasks.NewService(tasksRepo, projectsRepo, usersRepo),
		usersService:    users.NewService(usersRepo),
	}
}
