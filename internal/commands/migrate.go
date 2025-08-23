package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/urfave/cli/v3"
)

func runMigrations(_ context.Context, cmd *cli.Command, logger *log.Logger) error {
	direction := cmd.String("direction")
	steps := cmd.Int("steps")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	migrationsPath := "file://migrations"
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found")
	}

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	logger.Info("Running migrations", "direction", direction, "steps", steps)

	switch direction {
	case "up":
		if steps == 0 {
			err = m.Up()
		} else {
			err = m.Steps(steps)
		}
	case "down":
		if steps == 0 {
			err = m.Down()
		} else {
			err = m.Steps(-steps)
		}
	default:
		return fmt.Errorf("invalid direction: %s (use 'up' or 'down')", direction)
	}

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err == migrate.ErrNoChange {
		logger.Info("No migrations to run")
	} else {
		logger.Info("Migrations completed successfully")
	}

	return nil
}
