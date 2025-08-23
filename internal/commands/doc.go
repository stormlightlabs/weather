package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

func generateDocs(ctx context.Context, cmd *cli.Command, logger *log.Logger) error {
	outputDir := cmd.String("output")
	serve := cmd.Bool("serve")
	port := cmd.String("port")

	// Check if swag is installed
	if _, err := exec.LookPath("swag"); err != nil {
		logger.Warn("swag command not found, please install it: go install github.com/swaggo/swag/cmd/swag@latest")
		return fmt.Errorf("swag command not found, install with: go install github.com/swaggo/swag/cmd/swag@latest")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate swagger docs
	logger.Info("Generating swagger documentation", "output", outputDir)

	cmdE := exec.Command("swag", "init", "-g", "cmd/main.go", "--output", outputDir)
	cmdE.Stdout = os.Stdout
	cmdE.Stderr = os.Stderr

	args := cmd.Args().Slice()
	if err := cmd.Run(ctx, args); err != nil {
		return fmt.Errorf("failed to generate docs: %w", err)
	}

	logger.Info("Documentation generated successfully", "location", outputDir)

	if serve {
		return serveDocs(outputDir, port, logger)
	}

	return nil
}

func serveDocs(docsDir, port string, logger *log.Logger) error {
	swaggerFile := filepath.Join(docsDir, "swagger.json")
	if _, err := os.Stat(swaggerFile); os.IsNotExist(err) {
		return fmt.Errorf("swagger.json not found in %s", docsDir)
	}

	logger.Info("Serving documentation", "port", port, "docs", docsDir)

	// Serve static files from docs directory
	fs := http.FileServer(http.Dir(docsDir))
	http.Handle("/", fs)

	// Custom handler for swagger UI
	http.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		// Simple swagger UI HTML
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Weather API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/swagger.json',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	logger.Info("Documentation server started", "url", fmt.Sprintf("http://localhost:%s/swagger/", port))
	return http.ListenAndServe(":"+port, nil)
}
