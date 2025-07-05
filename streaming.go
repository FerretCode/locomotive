package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ferretcode/locomotive/internal/logger"
	"github.com/ferretcode/locomotive/internal/railway"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/http_logs"
	"github.com/sethvargo/go-retry"
)

func startStreamingDeployLogs(ctx context.Context, gqlClient *railway.GraphQLClient, serviceLogTrack chan []environment_logs.EnvironmentLogWithMetadata, environmentId string, serviceIds []string) error {
	b := retry.NewFibonacci(100 * time.Millisecond)
	b = retry.WithCappedDuration((5 * time.Second), b)
	b = retry.WithMaxRetries(10, b)

	if err := retry.Do(ctx, b, func(ctx context.Context) error {
		if err := environment_logs.SubscribeToServiceLogs(ctx, gqlClient, serviceLogTrack, environmentId, serviceIds); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return retry.RetryableError(err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error subscribing to deploy logs: %w", err)
	}

	logger.Stdout.Debug("deploy logs subscription ended")

	return nil
}

func startStreamingHttpLogs(ctx context.Context, gqlClient *railway.GraphQLClient, httpLogTrack chan []http_logs.DeploymentHttpLogWithMetadata, environmentId string, serviceIds []string) error {
	b := retry.NewFibonacci(100 * time.Millisecond)
	b = retry.WithCappedDuration((5 * time.Second), b)
	b = retry.WithMaxRetries(10, b)

	if err := retry.Do(ctx, b, func(ctx context.Context) error {
		if err := http_logs.SubscribeToHttpLogs(ctx, gqlClient, httpLogTrack, environmentId, serviceIds); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return retry.RetryableError(err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error subscribing to HTTP logs: %w", err)
	}

	logger.Stdout.Debug("HTTP logs subscription ended")

	return nil
}
