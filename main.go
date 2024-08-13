package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/logger"
	"github.com/ferretcode/locomotive/railway"
	"github.com/ferretcode/locomotive/util"
	"github.com/ferretcode/locomotive/webhook"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-retry"
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

	gqlClient, err := railway.NewClient(&railway.GraphQLClient{
		AuthToken:           cfg.RailwayApiKey,
		BaseURL:             "https://backboard.railway.app/graphql/v2",
		BaseSubscriptionURL: "wss://backboard.railway.app/graphql/internal",
	})
	if err != nil {
		logger.Stderr.Error("error creating graphql client", logger.ErrAttr(err))
		os.Exit(1)
	}

	logTrack := make(chan []railway.EnvironmentLog)

	ctx := context.Background()

	go func() {
		b := retry.NewFibonacci(100 * time.Millisecond)

		b = retry.WithCappedDuration((5 * time.Second), b)

		if err := retry.Do(ctx, b, func(ctx context.Context) error {
			if err := gqlClient.SubscribeToLogs(ctx, logTrack, cfg); err != nil {
				logger.Stderr.Error("error subscribing to logs", logger.ErrAttr(err))

				return retry.RetryableError(err)
			}

			return nil
		}); err != nil {
			logger.Stderr.Error("fatal error subscribing to logs", logger.ErrAttr(err))
		}

		logger.Stdout.Debug("log subscription ended")
	}()

	var logsTransported atomic.Int64

	go func() {
		t := time.NewTicker(cfg.ReportStatusEvery)
		defer t.Stop()

		for range t.C {
			logsSent := logsTransported.Load()

			if logsSent == 0 {
				continue
			}

			statusLog := logger.Stdout.With(slog.Int64("logs_transported", logsSent))

			if logger.StdoutLvl.Level() == slog.LevelDebug {
				memStats := &runtime.MemStats{}
				runtime.ReadMemStats(memStats)

				statusLog = statusLog.With(
					slog.String("total_alloc", util.ByteCountIEC(memStats.TotalAlloc)),
					slog.String("heap_alloc", util.ByteCountIEC(memStats.HeapAlloc)),
					slog.String("heap_in_use", util.ByteCountIEC(memStats.HeapInuse)),
					slog.String("stack_in_use", util.ByteCountIEC(memStats.StackInuse)),
					slog.String("other_sys", util.ByteCountIEC(memStats.OtherSys)),
					slog.String("sys", util.ByteCountIEC(memStats.Sys)),
				)
			}

			statusLog.Info("The locomotive is chugging along...")
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				os.Exit(0)
			case logs := <-logTrack:
				logsSent, errors := webhook.SendWebhooks(logs, cfg)
				if errorsLen := len(errors); errorsLen > 0 {
					logger.Stderr.Error("error sending webhook(s)", logger.ErrorsAttr(errors...))

					continue
				}

				logsTransported.Add(logsSent)
			}
		}
	}()

	logger.Stdout.Info("The locomotive is waiting for cargo...")

	<-done
}
