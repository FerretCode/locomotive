package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/graphql"
	"github.com/ferretcode/locomotive/logger"
	"github.com/ferretcode/locomotive/logline"
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

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	gqlClient, err := graphql.NewClient(&graphql.GraphQLClient{
		AuthToken:           cfg.RailwayApiKey,
		BaseURL:             "https://backboard.railway.app/graphql/v2",
		BaseSubscriptionURL: "wss://backboard.railway.app/graphql/internal",
	})
	if err != nil {
		logger.Stderr.Error("error creating graphql client", logger.ErrAttr(err))
		os.Exit(1)
	}

	logTrack := make(chan *graphql.EnvironmentLog)
	trackError := make(chan error)

	firstLog := true

	go func() {
		if err := gqlClient.SubscribeToLogs(logTrack, trackError, cfg); err != nil {
			logger.Stderr.Error("error subscribing to logs", logger.ErrAttr(err))
			os.Exit(1)
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				os.Exit(0)
			case log := <-logTrack:
				jsonLog, err := logline.ReconstructLogLine(log)
				if err != nil {
					return
				}

				if cfg.DiscordWebhookUrl != "" {
					if err := webhook.SendDiscordWebhook(jsonLog, log, true, cfg); err != nil {
						logger.Stderr.Error("error sending Discord webhook", logger.ErrAttr(err))
						continue
					}
				}

				if cfg.IngestUrl != "" {
					if err := webhook.SendGenericWebhook(jsonLog, cfg); err != nil {
						logger.Stderr.Error("error sending generic webhook", logger.ErrAttr(err))
						return
					}
				}

				if firstLog {
					logger.Stdout.Info("The locomotive is chugging along...")
					firstLog = false
				}
			case err := <-trackError:
				logger.Stderr.Error("error during log subscription", logger.ErrAttr(err))
				continue
			}
		}
	}()

	logger.Stdout.Info("The locomotive is waiting for cargo...")

	<-done
}
