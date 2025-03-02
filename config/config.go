package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
)

type AdditionalHeaders map[string]string

func (h *AdditionalHeaders) UnmarshalText(envByte []byte) error {
	envString := string(envByte)
	headers := make(map[string]string)

	headerPairs := strings.SplitN(envString, ";", 2)

	for _, header := range headerPairs {
		keyValue := strings.SplitN(header, "=", 2)

		if len(keyValue) != 2 {
			return fmt.Errorf("header key value pair must be in format k=v")
		}

		headers[strings.TrimSpace(keyValue[0])] = strings.TrimSpace(keyValue[1])
	}

	*h = headers

	return nil
}

type Config struct {
	RailwayApiKey string   `env:"RAILWAY_API_KEY,required"`
	EnvironmentId string   `env:"ENVIRONMENT_ID,required"`
	Train         []string `env:"TRAIN,required" envSeparator:","`

	DiscordWebhookUrl string `env:"DISCORD_WEBHOOK_URL"`
	DiscordPrettyJson bool   `env:"DISCORD_PRETTY_JSON" envDefault:"false"`

	SlackWebhookUrl string   `env:"SLACK_WEBHOOK_URL"`
	SlackPrettyJson bool     `env:"SLACK_PRETTY_JSON" envDefault:"false"`
	SlackTags       []string `env:"SLACK_TAGS" envSeparator:","`

	GrafanaIngestUrl string `env:"GRAFANA_INGEST_URL"`

	IngestUrl         string            `env:"INGEST_URL"`
	AdditionalHeaders AdditionalHeaders `env:"ADDITIONAL_HEADERS"`

	ReportStatusEvery time.Duration `env:"REPORT_STATUS_EVERY" envDefault:"10s"`

	LogsFilterGlobal  []string `env:"LOGS_FILTER" envSeparator:","`
	LogsFilterDiscord []string `env:"LOGS_FILTER_DISCORD" envSeparator:","`
	LogsFilterSlack   []string `env:"LOGS_FILTER_SLACK" envSeparator:","`
	LogsFilterGrafana []string `env:"LOGS_FILTER_GRAFANA" envSeparator:","`
	LogsFilterWebhook []string `env:"LOGS_FILTER_WEBHOOK" envSeparator:","`

	// New content filter fields
	LogsContentFilterGlobal  string `env:"LOGS_CONTENT_FILTER"`
	LogsContentFilterDiscord string `env:"LOGS_CONTENT_FILTER_DISCORD"`
	LogsContentFilterSlack   string `env:"LOGS_CONTENT_FILTER_SLACK"`
	LogsContentFilterGrafana string `env:"LOGS_CONTENT_FILTER_GRAFANA"`
	LogsContentFilterWebhook string `env:"LOGS_CONTENT_FILTER_WEBHOOK"`
}

func GetConfig() (*Config, error) {
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

	if config.DiscordWebhookUrl == "" && config.IngestUrl == "" && config.SlackWebhookUrl == "" && config.GrafanaIngestUrl == "" {
		return nil, errors.New("specify either a discord webhook url or an ingest url or a slack webhook url or a grafana url")
	}

	return &config, nil
}
