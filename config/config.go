package config

import "github.com/caarlos0/env/v10"

type Config struct {
	RailwayApiKey     string   `env:"RAILWAY_API_KEY,required"`
	EnvironmentId     string   `env:"ENVIRONMENT_ID,required"`
	Train             string   `env:"TRAIN,required"`
	LogsFilter        []string `env:"LOGS_FILTER" envSeparator:","`
	DiscordWebhookUrl string   `env:"DISCORD_WEBHOOK_URL"`
	IngestUrl         string   `env:"INGEST_URL"`
	AdditionalHeaders []string `env:"ADDITIONAL_HEADERS" envSeparator:";"`
}

func GetConfig() (Config, error) {
	config := Config{}

	if err := env.Parse(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}
