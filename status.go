package main

import (
	"context"
	"log/slog"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/ferretcode/locomotive/internal/config"
	"github.com/ferretcode/locomotive/internal/logger"
	"github.com/ferretcode/locomotive/internal/util"
)

func reportStatusAsync(ctx context.Context, cfg *config.Config, deployLogsProcessed *atomic.Int64, httpLogsProcessed *atomic.Int64) {
	go func() {
		t := time.NewTicker(cfg.ReportStatusEvery)
		defer t.Stop()

		for range t.C {
			deployLogsProcessed := deployLogsProcessed.Load()
			httpLogsProcessed := httpLogsProcessed.Load()

			if deployLogsProcessed == 0 && httpLogsProcessed == 0 {
				continue
			}

			statusLog := logger.Stdout.With(
				slog.Int64("deploy_logs_processed", deployLogsProcessed),
				slog.Int64("http_logs_processed", httpLogsProcessed),
			)

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
}
