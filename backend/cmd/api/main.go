package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskflow/backend/internal/auth"
	"taskflow/backend/internal/config"
	"taskflow/backend/internal/db"
	"taskflow/backend/internal/users"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	dbConn, err := db.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer dbConn.Close()

	usersRepo := users.NewRepository(dbConn.Querier())
	authService := auth.NewService(usersRepo, cfg.JWTSecret, cfg.JWTExpiryHours, cfg.BcryptCost)
	app := &application{
		logger:      logger,
		authService: authService,
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.AppPort),
		Handler:           newRouter(app),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info(
		"starting http server",
		slog.Int("port", cfg.AppPort),
		slog.Int("jwt_expiry_hours", cfg.JWTExpiryHours),
		slog.Int("bcrypt_cost", cfg.BcryptCost),
		slog.String("database", "connected"),
	)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-sigCtx.Done()

	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("http server stopped")
}
