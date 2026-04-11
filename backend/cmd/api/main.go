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

	"taskflow/backend/internal/config"
	"taskflow/backend/internal/db"
)

const shutdownTimeout = 10 * time.Second

const (
	startupTimeout   = 60 * time.Second
	startupAttempts  = 15
	startupRetryWait = 2 * time.Second
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	startupCtx, cancelStartup := context.WithTimeout(context.Background(), startupTimeout)
	defer cancelStartup()

	logger.Info("running database migrations")
	if err := db.RunMigrationsWithRetry(startupCtx, cfg.DatabaseURL, startupAttempts, startupRetryWait); err != nil {
		log.Fatalf("run database migrations: %v", err)
	}

	dbConn, err := db.NewWithRetry(startupCtx, cfg.DatabaseURL, startupAttempts, startupRetryWait)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer dbConn.Close()

	seedResult, err := db.RunSeedIfNeeded(startupCtx, dbConn, cfg.AutoSeed)
	if err != nil {
		log.Fatalf("run seed data: %v", err)
	}

	logger.Info(
		"seed runner finished",
		slog.Bool("auto_seed_enabled", seedResult.Enabled),
		slog.Bool("seed_applied", seedResult.Applied),
		slog.String("seed_result", seedResult.Reason),
	)

	app := newApplication(logger, cfg, dbConn)

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
		slog.String("migrations", "applied"),
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
