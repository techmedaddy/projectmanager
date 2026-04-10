package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"

	migrationfiles "taskflow/backend/migrations"
)

// RunMigrationsWithRetry applies all pending embedded SQL migrations, retrying
// while the database is still starting up.
func RunMigrationsWithRetry(ctx context.Context, databaseURL string, attempts int, delay time.Duration) error {
	return retry(ctx, attempts, delay, func() error {
		return runMigrations(ctx, databaseURL)
	})
}

func runMigrations(ctx context.Context, databaseURL string) error {
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open sql database: %w", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping sql database: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationfiles.Files, ".")
	if err != nil {
		return fmt.Errorf("create embedded migration source: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		_, _ = migrator.Close()
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

// NewWithRetry opens a pgx pool with retries so container startup can wait for
// PostgreSQL to become reachable.
func NewWithRetry(ctx context.Context, databaseURL string, attempts int, delay time.Duration) (*Database, error) {
	var conn *Database

	err := retry(ctx, attempts, delay, func() error {
		var err error
		conn, err = New(ctx, databaseURL)
		return err
	})
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func retry(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt == attempts {
			break
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("startup retry context canceled: %w", ctx.Err())
		case <-timer.C:
		}
	}

	return fmt.Errorf("startup operation failed after %d attempts: %w", attempts, lastErr)
}
