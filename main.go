package main

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/ferretcode/locomotive/internal/config"
	"github.com/ferretcode/locomotive/internal/errgroup"
	"github.com/ferretcode/locomotive/internal/logger"
	"github.com/ferretcode/locomotive/internal/railway"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/http_logs"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Stderr.Error("error parsing config", logger.ErrAttr(err))
		os.Exit(1)
	}

	if !cfg.EnableHttpLogs {
		logger.Stdout.Info("HTTP logs shipping is disabled. To enable it, set \"ENABLE_HTTP_LOGS=true\" as a service variable")
	}

	if !cfg.EnableDeployLogs {
		logger.Stdout.Info("Deploy logs shipping is disabled. To enable it, set \"ENABLE_DEPLOY_LOGS=true\" as a service variable")
	}

	logger.Stdout.Info("Starting the locomotive...",
		slog.Any("service_ids", cfg.Train),
		slog.String("environment_id", cfg.EnvironmentId),
	)

	gqlClient, err := railway.NewClient(&railway.GraphQLClient{
		AuthToken:           cfg.RailwayApiKey,
		BaseURL:             "https://backboard.railway.app/graphql/v2",
		BaseSubscriptionURL: "wss://backboard.railway.app/graphql/internal",
	})
	if err != nil {
		logger.Stderr.Error("error creating graphql client", logger.ErrAttr(err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serviceLogTrack := make(chan []environment_logs.EnvironmentLogWithMetadata)
	httpLogTrack := make(chan []http_logs.DeploymentHttpLogWithMetadata)

	deployLogsProcessed := atomic.Int64{}
	httpLogsProcessed := atomic.Int64{}

	reportStatusAsync(ctx, cfg, &deployLogsProcessed, &httpLogsProcessed)

	handleDeployLogsAsync(ctx, cfg, &deployLogsProcessed, serviceLogTrack)
	handleHttpLogsAsync(ctx, cfg, &httpLogsProcessed, httpLogTrack)

	errGroup := errgroup.NewErrGroup()

	errGroup.Go(func() error {
		if cfg.EnableDeployLogs {
			return startStreamingDeployLogs(ctx, gqlClient, serviceLogTrack, cfg.EnvironmentId, cfg.Train)
		}

		return nil
	})

	errGroup.Go(func() error {
		if cfg.EnableHttpLogs {
			return startStreamingHttpLogs(ctx, gqlClient, httpLogTrack, cfg.EnvironmentId, cfg.Train)
		}

		return nil
	})

	logger.Stdout.Info("The locomotive is waiting for cargo...")

	if err := errGroup.Wait(); err != nil {
		logger.Stderr.Error("error returned from subscription(s)", logger.ErrAttr(err))
		os.Exit(1)
	}
}
