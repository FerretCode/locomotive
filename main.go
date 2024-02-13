package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/ferretcode/locomotive/graphql"
	"github.com/ferretcode/locomotive/webhook"
	"github.com/joho/godotenv"
)

func main() {
	_, err := os.Stat(".env")

	if err == nil {
		err := godotenv.Load()

		if err != nil {
			log.Fatal(err)
		}
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

		err := graphQlClient.SubscribeToLogs(newLog)

		if err != nil {
			stderr.Println(err)
		}
	}()

	filters := strings.Split(os.Getenv("LOGS_FILTER"), ",")

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
					if os.Getenv("LOGS_FILTER") != "" &&
						os.Getenv("LOGS_FILTER") != "all" &&
						!slices.Contains(filters, log.Severity) {
						continue
					}

					if os.Getenv("DISCORD_WEBHOOK_URL") != "" {
						err = webhook.SendDiscordWebhook(graphql.Log{
							Message:  log.Message,
							Severity: log.Severity,
							Embed:    true,
						})

						if err != nil {
							fmt.Println(err)

							continue
						}
					}

					if os.Getenv("INGEST_URL") != "" {
						err = webhook.SendGenericWebhook(graphql.Log{
							Message:  log.Message,
							Severity: log.Severity,
							Embed:    true,
						})

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
