package environment_logs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/ferretcode/locomotive/internal/logger"
	"github.com/ferretcode/locomotive/internal/railway"
	"github.com/ferretcode/locomotive/internal/railway/gql/subscriptions"
	"github.com/ferretcode/locomotive/internal/railway/subscribe"
)

func createEnvironmentLogSubscription(ctx context.Context, client *railway.GraphQLClient, environmentId string, serviceIds []string) (*websocket.Conn, error) {
	payload := &subscriptions.EnvironmentLogsSubscriptionPayload{
		Query: subscriptions.EnvironmentLogsSubscription,
		Variables: &subscriptions.EnvironmentLogsSubscriptionVariables{
			EnvironmentId: environmentId,
			Filter:        buildServiceFilter(serviceIds),

			// needed for seamless subscription resuming
			BeforeDate:  time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339Nano),
			BeforeLimit: 500,
		},
	}

	return client.CreateWebSocketSubscription(ctx, payload)
}

// resubscribeWithRetry handles reconnection logic with retries and proper context cancellation
func resubscribeServiceLogsWithRetry(ctx context.Context, client *railway.GraphQLClient, environmentId string, serviceIds []string, conn *websocket.Conn) (*websocket.Conn, error) {
	subscribe.SafeConnCloseNow(conn)

	// Track total retry time with a maximum of 3600 seconds (1 hour)
	maxRetryDuration := 3600 * time.Second
	retryStart := time.Now()

	// Try to resubscribe with retry loop
	for {
		// Check if we've exceeded the maximum retry duration
		if time.Since(retryStart) > maxRetryDuration {
			return nil, fmt.Errorf("failed to resubscribe after %v: maximum retry duration exceeded", maxRetryDuration)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			newConn, err := createEnvironmentLogSubscription(ctx, client, environmentId, serviceIds)
			if err != nil {
				logger.Stdout.Debug("error resubscribing, will retry in 1 second", logger.ErrAttr(err))

				// Sleep with context cancellation awareness
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(1 * time.Second):
					continue
				}
			}

			// Successfully resubscribed
			return newConn, nil
		}
	}
}

func SubscribeToServiceLogs(ctx context.Context, g *railway.GraphQLClient, logTrack chan<- []EnvironmentLogWithMetadata, environmentId string, serviceIds []string) error {
	metadataMap, err := getMetadataMapForEnvironment(ctx, g.Client, environmentId)
	if err != nil {
		return fmt.Errorf("error getting metadata map: %w", err)
	}

	conn, err := createEnvironmentLogSubscription(ctx, g, environmentId, serviceIds)
	if err != nil {
		return err
	}

	defer conn.CloseNow()

	LogTime := time.Now().UTC()

	for {
		_, logPayload, err := subscribe.SafeConnRead(conn, ctx)
		if err != nil {
			logger.Stdout.Debug("resubscribing",
				slog.String("from", "SubscribeToEnvironmentLogs"),
				logger.ErrAttr(err),
			)

			conn, err = resubscribeServiceLogsWithRetry(ctx, g, environmentId, serviceIds, conn)
			if err != nil {
				return err
			}

			continue
		}

		logs := &subscriptions.EnvironmentLogsData{}

		if err := json.Unmarshal(logPayload, &logs); err != nil {
			return fmt.Errorf("error unmarshalling service logs: %w", err)
		}

		if logs.Type != subscriptions.SubscriptionTypeNext {
			logger.Stdout.Debug("resubscribing", slog.String("reason", fmt.Sprintf("log type not next: %s", logs.Type)))

			conn, err = resubscribeServiceLogsWithRetry(ctx, g, environmentId, serviceIds, conn)
			if err != nil {
				return err
			}

			continue
		}

		filteredLogs := []EnvironmentLogWithMetadata{}

		for i := range logs.Payload.Data.EnvironmentLogs {
			// skip logs with empty messages and no attributes
			// we check for 1 attribute because empty logs will always have at least one attribute, the level
			if logs.Payload.Data.EnvironmentLogs[i].Message == "" && len(logs.Payload.Data.EnvironmentLogs[i].Attributes) == 1 {
				continue
			}

			// skip container logs, container logs have trailing zeros in the timestamp
			if strings.HasSuffix(logs.Payload.Data.EnvironmentLogs[i].Timestamp.Format(time.StampNano), "000000000") {
				logger.Stdout.Debug("skipping container log message")
				continue
			}

			// on first subscription skip logs if they where logged before the first subscription, on resubscription skip logs if they where already processed
			if logs.Payload.Data.EnvironmentLogs[i].Timestamp.Before(LogTime) || LogTime == logs.Payload.Data.EnvironmentLogs[i].Timestamp {
				// logger.Stdout.Debug("skipping stale log message")
				continue
			}

			LogTime = logs.Payload.Data.EnvironmentLogs[i].Timestamp

			serviceName, ok := metadataMap[logs.Payload.Data.EnvironmentLogs[i].Tags.ServiceID]
			if !ok {
				logger.Stdout.Warn("service name could not be found")
				serviceName = "undefined"
			}

			environmentName, ok := metadataMap[logs.Payload.Data.EnvironmentLogs[i].Tags.EnvironmentID]
			if !ok {
				logger.Stdout.Warn("environment name could not be found")
				environmentName = "undefined"
			}

			projectName, ok := metadataMap[logs.Payload.Data.EnvironmentLogs[i].Tags.ProjectID]
			if !ok {
				logger.Stdout.Warn("project name could not be found")
				projectName = "undefined"
			}

			filteredLogs = append(filteredLogs, EnvironmentLogWithMetadata{
				Log: logs.Payload.Data.EnvironmentLogs[i],
				Metadata: EnvironmentLogMetadata{
					ProjectName:     projectName,
					EnvironmentName: environmentName,
					ServiceName:     serviceName,
				},
			})
		}

		if len(filteredLogs) == 0 {
			continue
		}

		logTrack <- filteredLogs
	}
}
