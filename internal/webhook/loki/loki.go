package loki

import (
	"fmt"
	"net/http"

	"github.com/ferretcode/locomotive/internal/config"
	"github.com/ferretcode/locomotive/internal/logline/reconstructor/reconstruct_loki"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/http_logs"
	"github.com/ferretcode/locomotive/internal/webhook/generic"
)

func SendWebhookForDeployLogs(logs []environment_logs.EnvironmentLogWithMetadata, cfg *config.Config, client *http.Client) error {
	jsonLogs, err := reconstruct_loki.EnvironmentLogLines(logs)
	if err != nil {
		return fmt.Errorf("failed to reconstruct deploy log lines: %w", err)
	}

	return generic.SendRawWebhook(jsonLogs, cfg.LokiIngestUrl, cfg.AdditionalHeaders, client)
}

func SendWebhookForHttpLogs(logs []http_logs.DeploymentHttpLogWithMetadata, cfg *config.Config, client *http.Client) error {
	jsonLogs, err := reconstruct_loki.HttpLogLines(logs)
	if err != nil {
		return fmt.Errorf("failed to reconstruct http log lines: %w", err)
	}

	return generic.SendRawWebhook(jsonLogs, cfg.LokiIngestUrl, cfg.AdditionalHeaders, client)
}
