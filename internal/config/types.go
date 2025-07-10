package config

import (
	"time"
)

type AdditionalHeaders map[string]string

type Trains []string

type Config struct {
	RailwayApiKey string `env:"RAILWAY_API_KEY,required"`
	EnvironmentId string `env:"ENVIRONMENT_ID,required"`
	Train         Trains `env:"TRAIN,required"`

	DiscordWebhookUrl string `env:"DISCORD_WEBHOOK_URL"`
	DiscordPrettyJson bool   `env:"DISCORD_PRETTY_JSON" envDefault:"false"`

	SlackWebhookUrl string   `env:"SLACK_WEBHOOK_URL"`
	SlackPrettyJson bool     `env:"SLACK_PRETTY_JSON" envDefault:"false"`
	SlackTags       []string `env:"SLACK_TAGS" envSeparator:","`

	LokiIngestUrl string `env:"LOKI_INGEST_URL"`

	IngestUrl         string            `env:"INGEST_URL"`
	AdditionalHeaders AdditionalHeaders `env:"ADDITIONAL_HEADERS"`

	ReportStatusEvery time.Duration `env:"REPORT_STATUS_EVERY" envDefault:"10s"`

	EnableHttpLogs   bool `env:"ENABLE_HTTP_LOGS" envDefault:"false"`
	EnableDeployLogs bool `env:"ENABLE_DEPLOY_LOGS" envDefault:"true"`

	LogsFilterGlobal  []string `env:"LOGS_FILTER" envSeparator:","`
	LogsFilterDiscord []string `env:"LOGS_FILTER_DISCORD" envSeparator:","`
	LogsFilterSlack   []string `env:"LOGS_FILTER_SLACK" envSeparator:","`
	LogsFilterLoki    []string `env:"LOGS_FILTER_LOKI" envSeparator:","`
	LogsFilterWebhook []string `env:"LOGS_FILTER_WEBHOOK" envSeparator:","`

	// New content filter fields
	LogsContentFilterGlobal  string `env:"LOGS_CONTENT_FILTER"`
	LogsContentFilterDiscord string `env:"LOGS_CONTENT_FILTER_DISCORD"`
	LogsContentFilterSlack   string `env:"LOGS_CONTENT_FILTER_SLACK"`
	LogsContentFilterLoki    string `env:"LOGS_CONTENT_FILTER_LOKI"`
	LogsContentFilterWebhook string `env:"LOGS_CONTENT_FILTER_WEBHOOK"`
}
