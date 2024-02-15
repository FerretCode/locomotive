package main

import (
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/graphql"
	"github.com/ferretcode/locomotive/logger"
	"github.com/ferretcode/locomotive/webhook"
	"github.com/joho/godotenv"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if godotenv.Load() != nil {
			logger.Stderr.Error("error loading .env file", logger.ErrAttr(err))
			os.Exit(1)
		}
	}

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Stderr.Error("error parsing config", logger.ErrAttr(err))
		os.Exit(1)
	}

	firstTrain := make(chan struct{}, 1)
	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	graphQlClient := graphql.GraphQLClient{
		BaseURL:             "https://backboard.railway.app/graphql/v2",
		BaseSubscriptionURL: "wss://backboard.railway.app/graphql/internal",
	}

	go func() {
		if err := graphQlClient.SubscribeToLogs(func(log *graphql.EnvironmentLog, err error) {
			if err != nil {
				logger.Stderr.Error("error during log subscription", logger.ErrAttr(err))
				return
			}

			if len(cfg.LogsFilter) > 0 && !slices.Contains(cfg.LogsFilter, "all") && !slices.Contains(cfg.LogsFilter, log.Severity) {
				return
			}

			if cfg.DiscordWebhookUrl != "" {
				if err := webhook.SendDiscordWebhook(log, true, cfg); err != nil {
					logger.Stderr.Error("error sending Discord webhook", logger.ErrAttr(err))
					return
				}
			}

			if cfg.IngestUrl != "" {
				if err := webhook.SendGenericWebhook(log, cfg); err != nil {
					logger.Stderr.Error("error sending generic webhook", logger.ErrAttr(err))
					return
				}
			}

			firstTrain <- struct{}{}
		}, cfg); err != nil {
			logger.Stderr.Error("error subscribing to logs", logger.ErrAttr(err))
			os.Exit(1)
		}
	}()

	logger.Stdout.Info("The locomotive is waiting for cargo...")

	<-firstTrain

	logger.Stdout.Info("The locomotive is chugging along...")

	<-done
}
