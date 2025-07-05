package main

import (
	"context"
	"sync/atomic"

	"github.com/ferretcode/locomotive/internal/config"
	"github.com/ferretcode/locomotive/internal/logger"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/http_logs"
	"github.com/ferretcode/locomotive/internal/webhook"
)

func handleDeployLogsAsync(ctx context.Context, cfg *config.Config, deployLogsProcessed *atomic.Int64, serviceLogTrack chan []environment_logs.EnvironmentLogWithMetadata) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case logs := <-serviceLogTrack:
				if errors := webhook.SendDeployLogsWebhook(logs, cfg); len(errors) > 0 {
					logger.Stderr.Error("error sending deploy logs webhook(s)", logger.ErrorsAttr(errors...))

					continue
				}

				deployLogsProcessed.Add(int64(len(logs)))
			}
		}
	}()
}

func handleHttpLogsAsync(ctx context.Context, cfg *config.Config, httpLogsProcessed *atomic.Int64, httpLogTrack chan []http_logs.DeploymentHttpLogWithMetadata) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case logs := <-httpLogTrack:
				if errors := webhook.SendHttpLogsWebhook(logs, cfg); len(errors) > 0 {
					logger.Stderr.Error("error sending http logs webhook(s)", logger.ErrorsAttr(errors...))

					continue
				}

				httpLogsProcessed.Add(int64(len(logs)))
			}
		}
	}()
}
