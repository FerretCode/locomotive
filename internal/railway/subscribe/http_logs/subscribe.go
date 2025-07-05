package http_logs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/coder/websocket"
	"github.com/ferretcode/locomotive/internal/logger"
	"github.com/ferretcode/locomotive/internal/railway"
	"github.com/ferretcode/locomotive/internal/railway/gql/subscriptions"
	"github.com/ferretcode/locomotive/internal/railway/subscribe"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/deployment_changes"
	"github.com/ferretcode/locomotive/internal/slice"
)

func createHttpLogSubscription(ctx context.Context, g *railway.GraphQLClient, deploymentId string) (*websocket.Conn, error) {
	payload := &subscriptions.HttpLogsSubscriptionPayload{
		Query: subscriptions.HttpLogsSubscription,
		Variables: &subscriptions.HttpLogsSubscriptionVariables{
			BeforeDate:   time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339Nano),
			BeforeLimit:  500,
			DeploymentId: deploymentId,
			Filter:       "",
		},
	}

	return g.CreateWebSocketSubscription(ctx, payload)
}

// resubscribeHttpLogsWithRetry handles reconnection logic with retries and proper context cancellation
func resubscribeHttpLogsWithRetry(ctx context.Context, g *railway.GraphQLClient, deploymentId string, conn *websocket.Conn) (*websocket.Conn, error) {
	subscribe.SafeConnCloseNow(conn)

	// Track total retry time with a maximum of 1200 seconds (20 minutes)
	maxRetryDuration := 1200 * time.Second
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
			newConn, err := createHttpLogSubscription(ctx, g, deploymentId)
			if err != nil {
				logger.Stdout.Debug("error resubscribing, will retry in 1 second",
					slog.String("deployment_id", deploymentId),
					logger.ErrAttr(err),
				)

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

func SubscribeToHttpLogs(ctx context.Context, g *railway.GraphQLClient, logTrack chan<- []DeploymentHttpLogWithMetadata, environmentId string, serviceIds []string) error {
	deploymentIdSlice := slice.NewSync[string]()
	changeDetected := make(chan struct{})
	errorChan := make(chan error, 1)

	go func() {
		logger.Stdout.Debug("starting deployment ID changes subscription", slog.String("environment_id", environmentId), slog.Any("service_ids", serviceIds))

		if err := deployment_changes.SubscribeToDeploymentIdChanges(ctx, g, deploymentIdSlice, changeDetected, environmentId, serviceIds); err != nil {
			if errors.Is(err, context.Canceled) {
				errorChan <- ctx.Err()
				return
			}

			errorChan <- fmt.Errorf("error subscribing to deployment id changes: %w", err)

			return
		}
	}()

	bufferedLogTrack := make(chan []DeploymentHttpLogWithMetadata)
	httpLogBuffer := slice.NewSync[DeploymentHttpLogWithMetadata]()

	go func() {
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if httpLogBuffer.Length() == 0 {
					continue
				}

				logTrack <- httpLogBuffer.Get()

				httpLogBuffer.Clear()
			case logs := <-bufferedLogTrack:
				httpLogBuffer.AppendMany(logs)
			}
		}
	}()

	// Track which deployment IDs have active goroutines
	activeDeploymentIds := slice.NewSync[string]()

	// Wait for initial deployment IDs
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errorChan:
		return err
	case <-changeDetected:
		// Initial deployment IDs received
		logger.Stdout.Debug("initial deployment IDs received", slog.Any("deployment_ids", deploymentIdSlice.Get()))

		// Start goroutines for initial deployment IDs
		for _, deploymentId := range deploymentIdSlice.Get() {
			logger.Stdout.Debug("starting initial HTTP log goroutine for deployment", slog.String("deployment_id", deploymentId))

			activeDeploymentIds.Append(deploymentId)

			go func() {
				defer activeDeploymentIds.Delete(deploymentId)

				if err := getHttpLogs(ctx, g, deploymentId, bufferedLogTrack, deploymentIdSlice); err != nil {
					errorChan <- err
				}
			}()
		}
	}

	// Main loop to handle deployment ID changes
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errorChan:
			return err
		case <-changeDetected:
			// Handle deployment ID changes
			currentDeploymentIds := deploymentIdSlice.Get()

			// Start new goroutines for deployment IDs that don't have active goroutines
			for _, deploymentId := range currentDeploymentIds {
				if !activeDeploymentIds.Contains(deploymentId) {
					logger.Stdout.Debug("starting new goroutine for new deployment", slog.String("deployment_id", deploymentId))

					activeDeploymentIds.Append(deploymentId)

					go func() {
						defer activeDeploymentIds.Delete(deploymentId)

						if err := getHttpLogs(ctx, g, deploymentId, bufferedLogTrack, deploymentIdSlice); err != nil {
							errorChan <- err
						}
					}()
				}
			}
		}
	}
}

func getHttpLogs(ctx context.Context, g *railway.GraphQLClient, initialDeploymentId string, logTrack chan<- []DeploymentHttpLogWithMetadata, activeDeploymentIds *slice.Sync[string]) error {
	conn, err := createHttpLogSubscription(ctx, g, initialDeploymentId)
	if err != nil {
		return fmt.Errorf("failed to create subscription for deployment %s: %w", initialDeploymentId, err)
	}

	defer subscribe.SafeConnCloseNow(conn)

	logTimes := time.Now()

	logger.Stdout.Debug("successfully created HTTP log subscription", slog.String("deployment_id", initialDeploymentId))

	metadata, err := getMetadataForDeployment(ctx, g, initialDeploymentId)
	if err != nil {
		return fmt.Errorf("error getting metadata for deployment %s: %w", initialDeploymentId, err)
	}

	// Main loop for reading from this specific connection
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Check if this deployment ID is still wanted
			if !activeDeploymentIds.Contains(initialDeploymentId) {
				logger.Stdout.Debug("deployment id no longer wanted, exiting goroutine",
					slog.String("deployment_id", initialDeploymentId),
					slog.String("from", "getHttpLogs_deploymentIdCheck"),
				)

				return nil
			}

			_, logPayload, err := subscribe.SafeConnRead(conn, ctx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					// No data available, continue
					continue
				}

				if !activeDeploymentIds.Contains(initialDeploymentId) {
					logger.Stdout.Debug("deployment id no longer wanted, exiting goroutine",
						slog.String("deployment_id", initialDeploymentId),
						slog.String("from", "getHttpLogs_connRead"),
					)

					return nil
				}

				logger.Stdout.Debug("resubscribing",
					slog.String("deployment_id", initialDeploymentId),
					slog.String("from", "getHttpLogs_connRead"),
					logger.ErrAttr(err),
				)

				// Close old connection and create new one
				subscribe.SafeConnCloseNow(conn)

				newConn, err := resubscribeHttpLogsWithRetry(ctx, g, initialDeploymentId, conn)
				if err != nil {
					return fmt.Errorf("failed to resubscribe for deployment %s: %w", initialDeploymentId, err)
				}

				conn = newConn

				continue
			}

			logs := &subscriptions.HttpLogsData{}

			if err := json.Unmarshal(logPayload, &logs); err != nil {
				logger.Stdout.Error("failed to unmarshal log payload",
					slog.String("deployment_id", initialDeploymentId),
					slog.String("from", "getHttpLogs_unmarshal"),
					logger.ErrAttr(err),
				)

				continue
			}

			if logs.Type != subscriptions.SubscriptionTypeNext {
				logger.Stdout.Debug("unexpected log type, resubscribing",
					slog.String("deployment_id", initialDeploymentId),
					slog.String("type", string(logs.Type)),
					slog.String("from", "getHttpLogs_typeCheck"),
				)

				// Close old connection and create new one
				subscribe.SafeConnCloseNow(conn)

				newConn, err := resubscribeHttpLogsWithRetry(ctx, g, initialDeploymentId, conn)
				if err != nil {
					logger.Stdout.Error("failed to resubscribe",
						slog.String("deployment_id", initialDeploymentId),
						slog.String("from", "getHttpLogs_typeCheck"),
						logger.ErrAttr(err),
					)

					return err
				}

				conn = newConn

				continue
			}

			if len(logs.Payload.Data.HTTPLogs) == 0 {
				continue
			}

			filteredHttpLogs := []DeploymentHttpLogWithMetadata{}

			for i := range logs.Payload.Data.HTTPLogs {
				logTimestamp, err := getTimeStampAttributeFromHttpLog(logs.Payload.Data.HTTPLogs[i])
				if err != nil {
					logger.Stdout.Error("failed to get timestamp from http log",
						slog.String("deployment_id", initialDeploymentId),
						slog.String("from", "getHttpLogs_payload_range"),
						logger.ErrAttr(err),
					)

					// we return an error here because this isn't something we can recover from
					// returning here will cause the goroutine to exit and the parent SubscribeToHttpLogs function to return the error
					return fmt.Errorf("failed to get timestamp from http log: %w", err)
				}

				if logTimestamp.Before(logTimes) || logTimes == logTimestamp {
					continue
				}

				path, err := getStringAttributeFromHttpLog(logs.Payload.Data.HTTPLogs[i], "path")
				if err != nil {
					logger.Stdout.Error("failed to get path from http log",
						slog.String("deployment_id", initialDeploymentId),
						slog.String("from", "getHttpLogs_payload_range"),
						logger.ErrAttr(err),
					)
				}

				filteredHttpLogs = append(filteredHttpLogs, DeploymentHttpLogWithMetadata{
					Log:       logs.Payload.Data.HTTPLogs[i],
					Path:      path,
					Timestamp: logTimestamp,
					Metadata:  metadata,
				})

				logTimes = logTimestamp
			}

			if len(filteredHttpLogs) == 0 {
				continue
			}

			logTrack <- filteredHttpLogs
		}
	}
}
