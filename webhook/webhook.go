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
	"github.com/ferretcode/locomotive/webhook/grafana"
	"github.com/ferretcode/locomotive/webhook/slack"
)

type webhookCtx struct {
	provider      string
	logsFilter    []string
	contentFilter string

	webhooksConfig *webhooksConfig

	sendWebhook func(logs []railway.EnvironmentLog, cfg *config.Config) error
}

type webhooksConfig struct {
	wg              *sync.WaitGroup
	cfg             *config.Config
	logsTransported *atomic.Int64
	logs            []railway.EnvironmentLog
	errChan         chan error
}

func SendGenericWebhook(logs []railway.EnvironmentLog, cfg *config.Config) error {
	return generic.SendWebhook(logs, cfg, client)
}

func SendDiscordWebhook(logs []railway.EnvironmentLog, cfg *config.Config) error {
	return discord.SendWebhook(logs, cfg, client)
}

func SendGrafanaWebhook(logs []railway.EnvironmentLog, cfg *config.Config) error {
	return grafana.SendWebhook(logs, cfg, client)
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

	webhooksConfig := &webhooksConfig{
		wg:              &wg,
		cfg:             cfg,
		logsTransported: &logsTransported,
		logs:            logs,
		errChan:         errChan,
	}

	if cfg.DiscordWebhookUrl != "" {
		sendWebhookWithProvider(webhookCtx{
			provider:       "discord",
			logsFilter:     cfg.LogsFilterDiscord,
			contentFilter:  cfg.LogsContentFilterDiscord,
			webhooksConfig: webhooksConfig,
			sendWebhook:    SendDiscordWebhook,
		})
	}

	if cfg.SlackWebhookUrl != "" {
		sendWebhookWithProvider(webhookCtx{
			provider:       "slack",
			logsFilter:     cfg.LogsFilterSlack,
			contentFilter:  cfg.LogsContentFilterSlack,
			webhooksConfig: webhooksConfig,
			sendWebhook:    SendSlackWebhook,
		})
	}

	if cfg.GrafanaIngestUrl != "" {
		sendWebhookWithProvider(webhookCtx{
			provider:       "grafana",
			logsFilter:     cfg.LogsFilterGrafana,
			contentFilter:  cfg.LogsContentFilterGrafana,
			webhooksConfig: webhooksConfig,
			sendWebhook:    SendGrafanaWebhook,
		})
	}

	if cfg.IngestUrl != "" {
		sendWebhookWithProvider(webhookCtx{
			provider:       "webhook",
			logsFilter:     cfg.LogsFilterWebhook,
			contentFilter:  cfg.LogsContentFilterWebhook,
			webhooksConfig: webhooksConfig,
			sendWebhook:    SendGenericWebhook,
		})
	}

	wg.Wait()

	return logsTransported.Load(), errors
}

func sendWebhookWithProvider(ctx webhookCtx) {
	ctx.webhooksConfig.wg.Add(1)

	go func() {
		defer ctx.webhooksConfig.wg.Done()

		filteredLogs := railway.FilterLogs(ctx.webhooksConfig.logs, ctx.logsFilter, ctx.contentFilter)

		logsLen, filteredLogsLen := len(ctx.webhooksConfig.logs), len(filteredLogs)

		if logsLen > filteredLogsLen {
			logger.Stdout.Debug(ctx.provider+"logs filtered",
				slog.Int("amount filtered", logsLen-filteredLogsLen),
				slog.Int("logs pre filter", logsLen),
				slog.Int("logs post filter", filteredLogsLen),
			)
		}

		err := ctx.sendWebhook(filteredLogs, ctx.webhooksConfig.cfg)
		if err != nil {
			ctx.webhooksConfig.errChan <- fmt.Errorf("ingest error: %w", err)
			ctx.webhooksConfig.wg.Add(1)
		}

		if err == nil {
			ctx.webhooksConfig.logsTransported.Add(int64(filteredLogsLen))
		}
	}()
}
