package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/logline"
	"github.com/ferretcode/locomotive/railway"
)

const descriptionFormat = "```%s```\n```%s```"

func SendWebhook(logs []railway.EnvironmentLog, cfg *config.Config, client *http.Client) error {
	em := []embed{}

	for i := range logs {
		rawLog, err := logline.ReconstructLogLine(logs[i])
		if err != nil {
			return fmt.Errorf("failed to reconstruct log object: %w", err)
		}

		var description string

		if cfg.DiscordPrettyJson {
			buf := &bytes.Buffer{}

			if err := json.Indent(buf, rawLog, "", "  "); err != nil {
				return fmt.Errorf("failed to indent json log object: %w", err)
			}

			description = fmt.Sprintf(descriptionFormat, logs[i].Message, buf)
		} else {
			description = fmt.Sprintf(descriptionFormat, logs[i].Message, rawLog)
		}

		em = append(em, embed{
			Title:       strings.ToUpper(logs[i].Severity),
			Color:       getColor(logs[i].Severity),
			Description: description,
			Timestamp:   logs[i].Timestamp,
		})
	}

	hook := hook{
		Username: "locomotive",
		Embeds:   em,
	}

	payload, err := json.Marshal(&hook)
	if err != nil {
		return fmt.Errorf("failed to marshal hook: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, cfg.DiscordWebhookUrl, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("non success status code: %d", res.StatusCode)
		}

		return fmt.Errorf("non success status code: %d; with body: %s", res.StatusCode, body)
	}

	return nil
}

// ref: https://gist.github.com/thomasbnt/b6f455e2c7d743b796917fa3c205f812
func getColor(severity string) int {
	severity = strings.ToLower(severity)

	color := 2303786 // black

	switch severity {
	case "info":
		color = 16777215 // white
	case "err", "error":
		color = 15548997 // red
	case "warn":
		color = 16776960 // yellow
	case "debug":
		color = 9807270 // grey
	}

	return color
}
