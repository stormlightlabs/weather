package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

func makeHTTPRequest(_ context.Context, cmd *cli.Command, logger *log.Logger) error {
	method := strings.ToUpper(cmd.String("method"))
	url := cmd.String("url")
	data := cmd.String("data")
	headers := cmd.String("headers")

	if url == "" {
		return fmt.Errorf("URL is required")
	}

	logger.Info("Making HTTP request", "method", method, "url", url)

	var body io.Reader
	if data != "" {
		body = strings.NewReader(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "weather-api-cli/1.0.0")

	// Parse and set additional headers
	if headers != "" {
		headerPairs := strings.Split(headers, ",")
		for _, pair := range headerPairs {
			if kv := strings.SplitN(strings.TrimSpace(pair), ":", 2); len(kv) == 2 {
				req.Header.Set(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]))
			}
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	logger.Info("Response received", "status", resp.Status, "content-type", resp.Header.Get("Content-Type"))

	// Pretty print JSON responses
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
			fmt.Printf("Status: %s\n", resp.Status)
			fmt.Printf("Headers:\n")
			for key, values := range resp.Header {
				for _, value := range values {
					fmt.Printf("  %s: %s\n", key, value)
				}
			}
			fmt.Printf("\nBody:\n%s\n", prettyJSON.String())
		} else {
			fmt.Printf("Status: %s\nBody:\n%s\n", resp.Status, string(respBody))
		}
	} else {
		fmt.Printf("Status: %s\nBody:\n%s\n", resp.Status, string(respBody))
	}

	return nil
}
