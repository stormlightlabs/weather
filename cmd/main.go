package main

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"

	"stormlightlabs.org/weather_api/internal/commands"
	"stormlightlabs.org/weather_api/internal/secrets"
)

func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
	})

	_, err := secrets.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}

	app := &cli.Command{
		Name:    "weather-api",
		Usage:   "Weather API CLI tool",
		Version: "1.0.0",
		Commands: []*cli.Command{
			commands.StartCommand(logger),
			commands.MigrateCommand(logger),
			commands.EncryptCommand(logger),
			commands.DecryptCommand(logger),
			commands.GenerateKeyCommand(logger),
			commands.HTTPCommand(logger),
			commands.DocCommand(logger),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		logger.Fatal("CLI execution failed", "error", err)
	}
}
