package config

import (
	"errors"
	"os"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

func GetConfig() (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}

	config := Config{}

	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	if config.DiscordWebhookUrl != "" && !strings.HasPrefix(config.DiscordWebhookUrl, "https://discord.com/api/webhooks/") {
		return nil, errors.New("invalid Discord webhook URL")
	}

	if config.SlackWebhookUrl != "" && !strings.HasPrefix(config.SlackWebhookUrl, "https://hooks.slack.com/services/") {
		return nil, errors.New("invalid Slack webhook URL")
	}

	if config.DiscordWebhookUrl == "" && config.IngestUrl == "" && config.SlackWebhookUrl == "" && config.LokiIngestUrl == "" {
		return nil, errors.New("specify either a discord webhook url or an ingest url or a slack webhook url or a loki url")
	}

	if !config.EnableDeployLogs && !config.EnableHttpLogs {
		return nil, errors.New("at least one of ENABLE_DEPLOY_LOGS or ENABLE_HTTP_LOGS must be true")
	}

	return &config, nil
}
