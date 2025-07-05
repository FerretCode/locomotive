package generic

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/ferretcode/locomotive/internal/config"
	"github.com/ferretcode/locomotive/internal/logline/reconstructor/reconstruct_json"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/http_logs"
)

var acceptedStatusCodes = []int{
	http.StatusOK,
	http.StatusNoContent,
	http.StatusAccepted,
	http.StatusCreated,
}

func SendWebhookForDeployLogs(logs []environment_logs.EnvironmentLogWithMetadata, cfg *config.Config, client *http.Client) error {
	jsonLogs, err := reconstruct_json.EnvironmentLogLinesJson(logs)
	if err != nil {
		return fmt.Errorf("failed to reconstruct deploy log lines: %w", err)
	}

	return SendRawWebhook(jsonLogs, cfg.IngestUrl, cfg.AdditionalHeaders, client)
}

func SendWebhookForHttpLogs(logs []http_logs.DeploymentHttpLogWithMetadata, cfg *config.Config, client *http.Client) error {
	jsonLogs, err := reconstruct_json.HttpLogLinesJson(logs)
	if err != nil {
		return fmt.Errorf("failed to reconstruct http log lines: %w", err)
	}

	return SendRawWebhook(jsonLogs, cfg.IngestUrl, cfg.AdditionalHeaders, client)
}

func SendRawWebhook(logs []byte, url string, additionalHeaders map[string]string, client *http.Client) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(logs))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Keep-Alive", "timeout=5, max=1000")

	for key, value := range additionalHeaders {
		req.Header.Set(key, value)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}

	defer res.Body.Close()

	if !slices.Contains(acceptedStatusCodes, res.StatusCode) {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("non success status code: %d", res.StatusCode)
		}

		return fmt.Errorf("non success status code: %d; with body: %s", res.StatusCode, body)
	}

	return nil
}
