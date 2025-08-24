package commands

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"

	"stormlightlabs.org/weather_api/internal/secrets"
)

func generateKey(_ context.Context, cmd *cli.Command, logger *log.Logger) error {
	length := cmd.Int("length")
	outputFile := cmd.String("output")
	quiet := cmd.Bool("quiet")

	key, err := secrets.GenerateSecureKey(length)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	if outputFile != "" {
		if err := secrets.WriteKeyToFile(key, outputFile); err != nil {
			return fmt.Errorf("failed to write key to file: %w", err)
		}

		if !quiet {
			logger.Info("Key generated successfully", "file", outputFile, "length", len(key))
			fmt.Printf("Encryption key written to: %s\n", outputFile)
			fmt.Printf("File permissions set to 0600 (owner read/write only)\n")
			fmt.Printf("Added to .gitignore to prevent accidental commits\n")
			fmt.Printf("\nTo use this key:\n")
			fmt.Printf("  export WEATHER_API_ENCRYPTION_KEY=\"$(cat %s)\"\n", outputFile)
			fmt.Printf("  # or\n")
			fmt.Printf("  weather-cli encrypt --key \"$(cat %s)\"\n", outputFile)
		}
	} else {
		// Output to stdout
		if quiet {
			fmt.Print(key)
		} else {
			logger.Info("Key generated successfully", "length", len(key))
			fmt.Printf("Generated encryption key (%d characters):\n", len(key))
			fmt.Printf("%s\n", key)
			fmt.Printf("\nTo use this key:\n")
			fmt.Printf("  export WEATHER_API_ENCRYPTION_KEY=\"%s\"\n", key)
			fmt.Printf("  # or\n")
			fmt.Printf("  weather-cli encrypt --key \"%s\"\n", key)
			fmt.Printf("\nTo save to file:\n")
			fmt.Printf("  weather-cli generate-key --output .env.key\n")
		}
	}

	return nil
}
