package db

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"

	seedfiles "taskflow/backend/seed"
)

const (
	seedUserEmail = "test@example.com"
)

// SeedResult describes what the seed runner did at startup.
type SeedResult struct {
	Enabled bool
	Applied bool
	Reason  string
}

// RunSeedIfNeeded applies embedded seed SQL exactly once per database when
// AUTO_SEED is enabled. It is safe to run on every startup.
func RunSeedIfNeeded(ctx context.Context, database *Database, enabled bool) (SeedResult, error) {
	if !enabled {
		return SeedResult{Enabled: false, Applied: false, Reason: "disabled_by_config"}, nil
	}

	if database == nil || database.Pool() == nil {
		return SeedResult{}, fmt.Errorf("seed runner requires initialized database")
	}

	alreadySeeded, err := hasSeedUser(ctx, database)
	if err != nil {
		return SeedResult{}, fmt.Errorf("check seed state: %w", err)
	}

	if alreadySeeded {
		return SeedResult{Enabled: true, Applied: false, Reason: "already_seeded"}, nil
	}

	seedScripts, err := loadSeedScripts()
	if err != nil {
		return SeedResult{}, fmt.Errorf("load seed scripts: %w", err)
	}

	if len(seedScripts) == 0 {
		return SeedResult{Enabled: true, Applied: false, Reason: "no_seed_scripts"}, nil
	}

	if err := database.WithTransaction(ctx, func(tx pgx.Tx) error {
		for _, script := range seedScripts {
			if strings.TrimSpace(script) == "" {
				continue
			}
			if _, execErr := tx.Exec(ctx, script); execErr != nil {
				return execErr
			}
		}
		return nil
	}); err != nil {
		return SeedResult{}, fmt.Errorf("apply seed scripts: %w", err)
	}

	return SeedResult{Enabled: true, Applied: true, Reason: "seed_applied"}, nil
}

func hasSeedUser(ctx context.Context, database *Database) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	if err := database.Querier().QueryRow(ctx, query, seedUserEmail).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func loadSeedScripts() ([]string, error) {
	entries, err := fs.ReadDir(seedfiles.Files, ".")
	if err != nil {
		return nil, err
	}

	seedFileNames := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			seedFileNames = append(seedFileNames, name)
		}
	}

	sort.Strings(seedFileNames)

	scripts := make([]string, 0, len(seedFileNames))
	for _, name := range seedFileNames {
		raw, readErr := fs.ReadFile(seedfiles.Files, name)
		if readErr != nil {
			return nil, readErr
		}
		scripts = append(scripts, string(raw))
	}

	return scripts, nil
}
