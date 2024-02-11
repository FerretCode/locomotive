package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/ferretcode/locomotive/graphql"
	"github.com/ferretcode/locomotive/railway"
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

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, os.Kill)

	pollingRate := int64(0)
	pollingRateSeconds := os.Getenv("POLLING_RATE_SECONDS")

	pollingRate, err = strconv.ParseInt(pollingRateSeconds, 10, 64)

	if err != nil {
		log.Fatal(errors.New("POLLING_RATE_SECONDS must be an integer"))
	}

	if pollingRate == 0 {
		pollingRate = 10
	}

	graphQlClient := graphql.GraphQLClient{
		BaseURL: "https://backboard.railway.app/graphql/v2",
	}

	ctx := context.Background()

	go func() {
		ticker := time.NewTicker(time.Duration(pollingRate) * time.Second)
		lastTimestamp := ""

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				deployments, err := railway.GetDeployments(ctx, graphQlClient)

				if err != nil {
					fmt.Println(err)
				}

				logs, err := railway.GetLogs(ctx, graphQlClient, deployments.Data.Deployments.Edges[0].Node.ID)

				if err != nil {
					fmt.Println(err)

					continue
				}

				lastLog := len(logs.Data.DeploymentLogs) - 1

				if logs.Data.DeploymentLogs[lastLog].Timestamp == lastTimestamp {
					continue
				}

				switch os.Getenv("LOGS_FILTER") {
				case "ALL":
					break
				case "ERROR":
					if logs.Data.DeploymentLogs[lastLog].Severity != railway.SEVERITY_ERROR {
						continue
					}
				case "INFO":
					if logs.Data.DeploymentLogs[lastLog].Severity != railway.SEVERITY_INFO {
						continue
					}
				}

				err = webhook.SendDiscordWebhook(webhook.Log{
					Message:  logs.Data.DeploymentLogs[lastLog].Message,
					Severity: logs.Data.DeploymentLogs[lastLog].Severity,
					Embed:    true,
				})

				if err != nil {
					fmt.Println(err)

					continue
				}

				lastTimestamp = logs.Data.DeploymentLogs[lastLog].Timestamp
			}
		}
	}()

	fmt.Println("The locomotive is chugging along...")

	<-done
}
