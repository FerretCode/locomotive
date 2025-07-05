package slack

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/ferretcode/locomotive/internal/config"
	"github.com/ferretcode/locomotive/internal/logline/reconstructor/reconstruct_slack"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
)

func SendWebhookForDeployLogs(logs []environment_logs.EnvironmentLogWithMetadata, cfg *config.Config, client *http.Client) error {
	payload, err := reconstruct_slack.EnvironmentLogLines(logs, cfg)
	if err != nil {
		return fmt.Errorf("failed to reconstruct deploy log lines: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, cfg.SlackWebhookUrl, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("non success status code: %d", res.StatusCode)
		}
		return fmt.Errorf("non success status code: %d; with body: %s", res.StatusCode, body)
	}

	return nil
}
