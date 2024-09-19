package webhook

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/logger"
	"github.com/ferretcode/locomotive/railway"
	"github.com/ferretcode/locomotive/webhook/discord"
	"github.com/ferretcode/locomotive/webhook/generic"
	"github.com/ferretcode/locomotive/webhook/slack"
)

func SendGenericWebhook(logs []railway.EnvironmentLog, cfg *config.Config) error {
	return generic.SendWebhook(logs, cfg, client)
}

func SendDiscordWebhook(logs []railway.EnvironmentLog, cfg *config.Config) error {
	return discord.SendWebhook(logs, cfg, client)
}

func SendSlackWebhook(logs []railway.EnvironmentLog, cfg *config.Config) error {
	return slack.SendWebhook(logs, cfg, client)
}

func SendWebhooks(logs []railway.EnvironmentLog, cfg *config.Config) (int64, []error) {
	logsTransported := atomic.Int64{}

	errChan := make(chan error)
	defer close(errChan)

	errors := []error{}

	wg := sync.WaitGroup{}

	go func() {
		for err := range errChan {
			errors = append(errors, err)
			wg.Done()
		}
	}()

	if cfg.DiscordWebhookUrl != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			filteredLogs := railway.FilterLogs(logs, cfg.LogsFilterDiscord, cfg.LogsContentFilterDiscord)

			logsLen, filteredLogsLen := len(logs), len(filteredLogs)

			if logsLen > filteredLogsLen {
				logger.Stdout.Debug("discord logs filtered",
					slog.Int("amount filtered", logsLen-filteredLogsLen),
					slog.Int("logs pre filter", logsLen),
					slog.Int("logs post filter", filteredLogsLen),
				)
			}

			err := SendDiscordWebhook(filteredLogs, cfg)
			if err != nil {
				errChan <- fmt.Errorf("discord error: %w", err)
				wg.Add(1)
			}

			if err == nil {
				logsTransported.Add(int64(filteredLogsLen))
			}
		}()
	}

	if cfg.SlackWebhookUrl != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			filteredLogs := railway.FilterLogs(logs, cfg.LogsFilterSlack, cfg.LogsContentFilterSlack)

			logsLen, filteredLogsLen := len(logs), len(filteredLogs)

			if logsLen > filteredLogsLen {
				logger.Stdout.Debug("slack logs filtered",
					slog.Int("amount filtered", logsLen-filteredLogsLen),
					slog.Int("logs pre filter", logsLen),
					slog.Int("logs post filter", filteredLogsLen),
				)
			}

			err := SendSlackWebhook(filteredLogs, cfg)
			if err != nil {
				errChan <- fmt.Errorf("slack error: %w", err)
				wg.Add(1)
			}

			if err == nil {
				logsTransported.Add(int64(filteredLogsLen))
			}
		}()
	}

	if cfg.IngestUrl != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			filteredLogs := railway.FilterLogs(logs, cfg.LogsFilterWebhook, cfg.LogsContentFilterWebhook)

			logsLen, filteredLogsLen := len(logs), len(filteredLogs)

			if logsLen > filteredLogsLen {
				logger.Stdout.Debug("webhook logs filtered",
					slog.Int("amount filtered", logsLen-filteredLogsLen),
					slog.Int("logs pre filter", logsLen),
					slog.Int("logs post filter", filteredLogsLen),
				)
			}

			err := SendGenericWebhook(filteredLogs, cfg)
			if err != nil {
				errChan <- fmt.Errorf("ingest error: %w", err)
				wg.Add(1)
			}

			if err == nil {
				logsTransported.Add(int64(filteredLogsLen))
			}
		}()
	}

	wg.Wait()

	return logsTransported.Load(), errors
}
