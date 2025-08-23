package commands

import (
	"context"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

func startServer(_ context.Context, cmd *cli.Command, logger *log.Logger) error {
	host := cmd.String("host")
	port := cmd.String("port")
	addr := fmt.Sprintf("%s:%s", host, port)

	logger.Info("Starting weather API server", "address", addr)

	// TODO: Replace with actual server implementation
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"weather-api"}`)
	})

	logger.Info("Server listening", "address", addr)
	return http.ListenAndServe(addr, nil)
}
