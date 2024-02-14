package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/graphql"
	"github.com/ferretcode/locomotive/webhook"
	"github.com/joho/godotenv"
)

func main() {
	var cfg config.Config

	if _, err := os.Stat(".env"); err == nil {
		if godotenv.Load() != nil {
			log.Fatal(err)
		}
	}

	cfg, err := config.GetConfig()

	if err != nil {
		log.Fatal(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	graphQlClient := graphql.GraphQLClient{
		BaseURL:             "https://backboard.railway.app/graphql/v2",
		BaseSubscriptionURL: "wss://backboard.railway.app/graphql/internal",
	}

	newLog := make(chan graphql.SubscriptionLogResponse)

	go func() {
		stderr := log.New(os.Stderr, "", 0)

		err := graphQlClient.SubscribeToLogs(newLog, cfg)

		if err != nil {
			stderr.Println(err)
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case logs := <-newLog:
				if len(logs.EnvironmentLogs) == 0 {
					continue
				}

				for _, log := range logs.EnvironmentLogs {
					if len(cfg.LogsFilter) > 0 &&
						!slices.Contains(cfg.LogsFilter, "all") &&
						!slices.Contains(cfg.LogsFilter, strings.ToLower(log.Severity)) {
						continue
					}

					graphQlLog := graphql.Log{
						Message:    log.Message,
						Severity:   log.Severity,
						Attributes: log.Attributes,
						Embed:      true,
					}

					if cfg.DiscordWebhookUrl != "" {
						err = webhook.SendDiscordWebhook(graphQlLog, cfg)

						if err != nil {
							fmt.Println(err)

							continue
						}
					}

					if cfg.IngestUrl != "" {
						err = webhook.SendGenericWebhook(graphQlLog, cfg)

						if err != nil {
							fmt.Println(err)

							continue
						}
					}
				}
			}
		}
	}()

	fmt.Println("The locomotive is chugging along...")

	<-done
}
