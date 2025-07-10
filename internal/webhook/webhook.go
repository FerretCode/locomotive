package webhook

import (
	"fmt"
	"sync"

	"github.com/ferretcode/locomotive/internal/config"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/http_logs"
	"github.com/ferretcode/locomotive/internal/webhook/discord"
	"github.com/ferretcode/locomotive/internal/webhook/generic"
	"github.com/ferretcode/locomotive/internal/webhook/loki"
	"github.com/ferretcode/locomotive/internal/webhook/slack"
)

func SendDeployLogsWebhook(logs []environment_logs.EnvironmentLogWithMetadata, cfg *config.Config) []error {
	errChan := make(chan error)
	defer close(errChan)

	errors := []error{}

	wg := sync.WaitGroup{}

	go func() {
		for err := range errChan {
			errors = append(errors, err)
		}
	}()

	globalFilteredLogs := environment_logs.FilterLogs(logs, cfg.LogsFilterGlobal, cfg.LogsContentFilterGlobal)

	if cfg.IngestUrl != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			webhookFilteredLogs := environment_logs.FilterLogs(globalFilteredLogs, cfg.LogsFilterWebhook, cfg.LogsContentFilterWebhook)

			if err := generic.SendWebhookForDeployLogs(webhookFilteredLogs, cfg, client); err != nil {
				errChan <- fmt.Errorf("failed to send generic webhook for deploy logs: %w", err)
			}
		}()
	}

	if cfg.DiscordWebhookUrl != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			discordFilteredLogs := environment_logs.FilterLogs(globalFilteredLogs, cfg.LogsFilterDiscord, cfg.LogsContentFilterDiscord)

			if err := discord.SendWebhookForDeployLogs(discordFilteredLogs, cfg, client); err != nil {
				errChan <- fmt.Errorf("failed to send discord webhook for deploy logs: %w", err)
			}
		}()
	}

	if cfg.SlackWebhookUrl != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			slackFilteredLogs := environment_logs.FilterLogs(globalFilteredLogs, cfg.LogsFilterSlack, cfg.LogsContentFilterSlack)

			if err := slack.SendWebhookForDeployLogs(slackFilteredLogs, cfg, client); err != nil {
				errChan <- fmt.Errorf("failed to send slack webhook for deploy logs: %w", err)
			}
		}()
	}

	if cfg.LokiIngestUrl != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			lokiFilteredLogs := environment_logs.FilterLogs(globalFilteredLogs, cfg.LogsFilterLoki, cfg.LogsContentFilterLoki)

			if err := loki.SendWebhookForDeployLogs(lokiFilteredLogs, cfg, client); err != nil {
				errChan <- fmt.Errorf("failed to send loki webhook for deploy logs: %w", err)
			}
		}()
	}

	wg.Wait()

	return errors
}

func SendHttpLogsWebhook(logs []http_logs.DeploymentHttpLogWithMetadata, cfg *config.Config) []error {
	errChan := make(chan error)
	defer close(errChan)

	errors := []error{}

	wg := sync.WaitGroup{}

	go func() {
		for err := range errChan {
			errors = append(errors, err)
		}
	}()

	if cfg.IngestUrl != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := generic.SendWebhookForHttpLogs(logs, cfg, client); err != nil {
				errChan <- fmt.Errorf("failed to send generic webhook for http logs: %w", err)
			}
		}()
	}

	if cfg.LokiIngestUrl != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := loki.SendWebhookForHttpLogs(logs, cfg, client); err != nil {
				errChan <- fmt.Errorf("failed to send loki webhook for http logs: %w", err)
			}
		}()
	}

	wg.Wait()

	return errors
}
