package environment_invalidation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/coder/websocket"
	"github.com/ferretcode/locomotive/internal/logger"
	"github.com/ferretcode/locomotive/internal/railway"
	"github.com/ferretcode/locomotive/internal/railway/gql/subscriptions"
	"github.com/ferretcode/locomotive/internal/railway/subscribe"
)

func createInvalidationRequestSubscription(ctx context.Context, g *railway.GraphQLClient, environmentId string) (*websocket.Conn, error) {
	payload := &subscriptions.CanvasInvalidationSubscriptionPayload{
		Query: subscriptions.CanvasInvalidationSubscription,
		Variables: &subscriptions.CanvasInvalidationSubscriptionVariables{
			EnvironmentId: environmentId,
		},
	}

	return g.CreateWebSocketSubscription(ctx, payload)
}

// resubscribeWithRetry handles reconnection logic with retries and proper context cancellation
func resubscribeWithRetry(ctx context.Context, g *railway.GraphQLClient, environmentId string, conn *websocket.Conn) (*websocket.Conn, error) {
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
			newConn, err := createInvalidationRequestSubscription(ctx, g, environmentId)
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

func SubscribeToInvalidationRequests(ctx context.Context, g *railway.GraphQLClient, environmentHashTrack chan<- string, environmentId string) error {
	conn, err := createInvalidationRequestSubscription(ctx, g, environmentId)
	if err != nil {
		return err
	}

	defer conn.CloseNow()

	lastHash := ""

	for {
		_, payload, err := subscribe.SafeConnRead(conn, ctx)
		if err != nil {
			logger.Stdout.Debug("resubscribing",
				slog.String("from", "SubscribeToInvalidationRequests"),
				logger.ErrAttr(err),
			)

			conn, err = resubscribeWithRetry(ctx, g, environmentId, conn)
			if err != nil {
				return err
			}

			continue
		}

		invalidationRequest := &subscriptions.CanvasInvalidationData{}

		if err := json.Unmarshal(payload, &invalidationRequest); err != nil {
			return fmt.Errorf("error unmarshalling invalidation request: %w", err)
		}

		if invalidationRequest.Type != subscriptions.SubscriptionTypeNext || invalidationRequest.Type == subscriptions.SubscriptionTypeComplete {
			// logger.Stdout.Debug("resubscribing", slog.String("err", fmt.Sprintf("log type not next: %s", invalidationRequest.Type)))
			conn, err = resubscribeWithRetry(ctx, g, environmentId, conn)
			if err != nil {
				return err
			}

			continue
		}

		if lastHash == "" {
			// logger.Stdout.Debug("skipping because last hash is empty", slog.String("id", invalidationRequest.Payload.Data.CanvasInvalidation.ID))
			lastHash = invalidationRequest.Payload.Data.CanvasInvalidation.ID
			continue
		}

		if invalidationRequest.Payload.Data.CanvasInvalidation.ID == lastHash {
			// logger.Stdout.Debug("skipping because last hash is the same", slog.String("id", invalidationRequest.Payload.Data.CanvasInvalidation.ID))
			continue
		}

		lastHash = invalidationRequest.Payload.Data.CanvasInvalidation.ID

		environmentHashTrack <- invalidationRequest.Payload.Data.CanvasInvalidation.ID
	}
}
