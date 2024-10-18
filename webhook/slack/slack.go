package slack

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

func SendWebhook(logs []railway.EnvironmentLog, cfg *config.Config, client *http.Client) error {
	blocks := []Block{}

	// Prepare user tags
	var userTags string
	if len(cfg.SlackTags) > 0 {
		tags := make([]string, len(cfg.SlackTags))
		for i, tag := range cfg.SlackTags {
			tags[i] = fmt.Sprintf("<@%s>", tag)
		}
		userTags = strings.Join(tags, " ") + "\n"
	}

	for _, log := range logs {
		rawLog, err := logline.ReconstructLogLine(log)
		if err != nil {
			return fmt.Errorf("failed to reconstruct log object: %w", err)
		}

		var logJson string
		if cfg.SlackPrettyJson {
			buf := &bytes.Buffer{}
			if err := json.Indent(buf, rawLog, "", "  "); err != nil {
				return fmt.Errorf("failed to indent json log object: %w", err)
			}
			logJson = buf.String()
		} else {
			logJson = string(rawLog)
		}

		blocks = append(blocks, Block{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: fmt.Sprintf("%s*%s*\n```%s```\n```%s```", userTags, strings.ToUpper(log.Severity), log.Message, logJson),
			},
		})

		blocks = append(blocks, Block{
			Type: "context",
			Elements: []Element{
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Timestamp:* %s", log.Timestamp),
				},
			},
		})

		blocks = append(blocks, Block{
			Type: "divider",
		})
	}

	message := SlackMessage{
		Blocks: blocks,
	}

	payload, err := json.Marshal(&message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, cfg.SlackWebhookUrl, bytes.NewReader(payload))
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

// Note: The getColor function has been removed as it's not needed for Slack Blocks
