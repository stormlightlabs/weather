package commands

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"

	"stormlightlabs.org/weather_api/internal/secrets"
)

type commandInterface interface {
	Int(name string) int
	String(name string) string
	Bool(name string) bool
}

type mockCommand struct {
	lengthVal int
	outputVal string
	quietVal  bool
}

func (m *mockCommand) Int(name string) int {
	if name == "length" {
		return m.lengthVal
	}
	return 0
}

func (m *mockCommand) String(name string) string {
	if name == "output" {
		return m.outputVal
	}
	return ""
}

func (m *mockCommand) Bool(name string) bool {
	if name == "quiet" {
		return m.quietVal
	}
	return false
}

func testGenerateKey(_ context.Context, cmd commandInterface, logger *log.Logger) error {
	length := cmd.Int("length")
	outputFile := cmd.String("output")
	quiet := cmd.Bool("quiet")

	key, err := secrets.GenerateSecureKey(length)
	if err != nil {
		return err
	}

	validator := secrets.NewKeyValidator()
	if err := validator.ValidateKey(key); err != nil {
		return err
	}

	if outputFile != "" && !quiet {
		logger.Info("Key generated successfully", "file", outputFile, "length", len(key))
	} else if quiet {
		os.Stdout.WriteString(key)
	} else {
		logger.Info("Key generated successfully", "length", len(key))
	}

	return nil
}

func TestGenerateKey(t *testing.T) {
	t.Run("Command", func(t *testing.T) {
		logger := log.NewWithOptions(os.Stderr, log.Options{
			ReportCaller:    false,
			ReportTimestamp: false,
		})

		tests := []struct {
			name     string
			length   int
			output   string
			quiet    bool
			wantFile bool
			wantErr  bool
		}{
			{
				name:     "default parameters",
				length:   16,
				output:   ".env.key",
				quiet:    false,
				wantFile: true,
				wantErr:  false,
			},
			{
				name:     "custom length",
				length:   24,
				output:   ".env.key",
				quiet:    false,
				wantFile: true,
				wantErr:  false,
			},
			{
				name:     "stdout output",
				length:   16,
				output:   "",
				quiet:    false,
				wantFile: false,
				wantErr:  false,
			},
			{
				name:     "quiet mode with file",
				length:   16,
				output:   "test.key",
				quiet:    true,
				wantFile: true,
				wantErr:  false,
			},
			{
				name:     "minimum length",
				length:   12,
				output:   "min.key",
				quiet:    false,
				wantFile: true,
				wantErr:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mockCmd := &mockCommand{
					lengthVal: tt.length,
					outputVal: tt.output,
					quietVal:  tt.quiet,
				}

				err := testGenerateKey(context.Background(), mockCmd, logger)

				if (err != nil) != tt.wantErr {
					t.Errorf("generateKey() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			})
		}
	})

	t.Run("Uniqueness", func(t *testing.T) {
		keys := make(map[string]bool)
		iterations := 10

		for range iterations {
			key, err := secrets.GenerateSecureKey(16)
			if err != nil {
				t.Fatalf("GenerateSecureKey() failed: %v", err)
			}

			if keys[key] {
				t.Errorf("Duplicate key generated: %s", key)
			}
			keys[key] = true

			validator := secrets.NewKeyValidator()
			if err := validator.ValidateKey(key); err != nil {
				t.Errorf("Generated key failed validation: %v", err)
			}
		}

		if len(keys) != iterations {
			t.Errorf("Expected %d unique keys, got %d", iterations, len(keys))
		}
	})

	t.Run("Invalid Length", func(t *testing.T) {
		key, err := secrets.GenerateSecureKey(5)
		if err != nil {
			t.Fatalf("GenerateSecureKey() failed: %v", err)
		}

		if len(key) < 12 {
			t.Errorf("Generated key length %d is below minimum of 12", len(key))
		}

		validator := secrets.NewKeyValidator()
		if err := validator.ValidateKey(key); err != nil {
			t.Errorf("Generated key failed validation: %v", err)
		}
	})

	t.Run("Gitignore", func(t *testing.T) {
		t.Run("Entry Generation", func(t *testing.T) {
			keyFile := "test.key"
			existingContent := ""

			var builder strings.Builder
			written, err := secrets.WriteGitignoreEntry(&builder, existingContent, keyFile)
			if err != nil {
				t.Fatalf("WriteGitignoreEntry failed: %v", err)
			}

			if !written {
				t.Error("Expected entry to be written for new file")
			}

			result := builder.String()

			if !strings.Contains(result, keyFile) {
				t.Errorf("Key file %s not found in generated gitignore content", keyFile)
			}
			if !strings.Contains(result, "Weather API encryption key") {
				t.Errorf("Expected comment not found in generated gitignore content")
			}

			expectedLines := strings.Split(result, "\n")
			if len(expectedLines) < 3 {
				t.Errorf("Generated gitignore content should have at least 3 lines, got %d", len(expectedLines))
			}
		})

		t.Run("Content Preservation", func(t *testing.T) {
			keyFile := "another.key"
			existingContent := "# Existing content\n*.log\n"

			var builder strings.Builder
			builder.WriteString(existingContent)

			written, err := secrets.WriteGitignoreEntry(&builder, existingContent, keyFile)
			if err != nil {
				t.Fatalf("WriteGitignoreEntry failed: %v", err)
			}

			if !written {
				t.Error("Expected entry to be written for new file")
			}

			result := builder.String()

			if !strings.Contains(result, "*.log") {
				t.Errorf("Existing .gitignore content was not preserved")
			}
			if !strings.Contains(result, "# Existing content") {
				t.Errorf("Existing .gitignore comment was not preserved")
			}

			if !strings.Contains(result, keyFile) {
				t.Errorf("Key file %s not found in updated gitignore content", keyFile)
			}
			if !strings.Contains(result, "Weather API encryption key") {
				t.Errorf("Expected comment not found in updated gitignore content")
			}
		})

		t.Run("Duplicate Preservation", func(t *testing.T) {
			keyFile := "duplicate.key"
			existingContent := "# Existing content\n" + keyFile + "\n*.log\n"

			var builder strings.Builder
			builder.WriteString(existingContent)

			written, err := secrets.WriteGitignoreEntry(&builder, existingContent, keyFile)
			if err != nil {
				t.Fatalf("WriteGitignoreEntry failed: %v", err)
			}

			if written {
				t.Error("Expected entry NOT to be written for existing file")
			}

			result := builder.String()

			count := strings.Count(result, keyFile)
			if count != 1 {
				t.Errorf("Key file %s appears %d times in final content, expected 1", keyFile, count)
			}

			if result != existingContent {
				t.Errorf("Content was modified when it shouldn't have been.\nExpected: %q\nGot: %q", existingContent, result)
			}
		})
	})
}
