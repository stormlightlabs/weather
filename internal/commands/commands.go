package commands

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

// StartCommand creates the server start command
func StartCommand(logger *log.Logger) *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "Start the weather API server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "port",
				Value: "8080",
				Usage: "Server port",
			},
			&cli.StringFlag{
				Name:  "host",
				Value: "localhost",
				Usage: "Server host",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return startServer(ctx, cmd, logger)
		},
	}
}

// MigrateCommand creates the database migration command
func MigrateCommand(logger *log.Logger) *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Run database migrations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "direction",
				Value: "up",
				Usage: "Migration direction (up/down)",
			},
			&cli.IntFlag{
				Name:  "steps",
				Value: 0,
				Usage: "Number of migration steps (0 = all)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runMigrations(ctx, cmd, logger)
		},
	}
}

// EncryptCommand creates the env encryption command
func EncryptCommand(logger *log.Logger) *cli.Command {
	return &cli.Command{
		Name:  "encrypt",
		Usage: "Encrypt env.local file values",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Value: "env.local",
				Usage: "Environment file to encrypt",
			},
			&cli.StringFlag{
				Name:  "key",
				Usage: "Encryption key (optional, will prompt if not provided)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return encryptEnvFile(ctx, cmd, logger)
		},
	}
}

// DecryptCommand creates the env decryption command
func DecryptCommand(logger *log.Logger) *cli.Command {
	return &cli.Command{
		Name:  "decrypt",
		Usage: "Decrypt env.local file values",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Value: "env.local",
				Usage: "Environment file to decrypt",
			},
			&cli.StringFlag{
				Name:  "key",
				Usage: "Decryption key (optional, will prompt if not provided)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return decryptEnvFile(ctx, cmd, logger)
		},
	}
}

// HTTPCommand creates the HTTP request command
func HTTPCommand(logger *log.Logger) *cli.Command {
	return &cli.Command{
		Name:  "http",
		Usage: "Make HTTP requests to the API",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "method",
				Value: "GET",
				Usage: "HTTP method",
			},
			&cli.StringFlag{
				Name:  "url",
				Usage: "API endpoint URL",
			},
			&cli.StringFlag{
				Name:  "data",
				Usage: "Request body data (JSON)",
			},
			&cli.StringFlag{
				Name:  "headers",
				Usage: "Additional headers (comma-separated key:value pairs)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return makeHTTPRequest(ctx, cmd, logger)
		},
	}
}

// DocCommand creates the swagger documentation generation command
func DocCommand(logger *log.Logger) *cli.Command {
	return &cli.Command{
		Name:  "doc",
		Usage: "Generate swagger documentation",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "output",
				Value: "docs",
				Usage: "Output directory for generated docs",
			},
			&cli.BoolFlag{
				Name:  "serve",
				Usage: "Serve documentation after generation",
			},
			&cli.StringFlag{
				Name:  "port",
				Value: "8081",
				Usage: "Documentation server port",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return generateDocs(ctx, cmd, logger)
		},
	}
}
