package reconstruct_slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ferretcode/locomotive/internal/config"
	"github.com/ferretcode/locomotive/internal/logline/reconstructor/reconstruct_json"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
	"github.com/ferretcode/locomotive/internal/util"
)

func EnvironmentLogLines(logs []environment_logs.EnvironmentLogWithMetadata, cfg *config.Config) ([]byte, error) {
	blocks := []block{}

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
		rawLog, err := reconstruct_json.EnvironmentLogLineJson(log)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct log object: %w", err)
		}

		var logJson string
		if cfg.SlackPrettyJson {
			buf := &bytes.Buffer{}
			if err := json.Indent(buf, rawLog, "", "  "); err != nil {
				return nil, fmt.Errorf("failed to indent json log object: %w", err)
			}
			logJson = buf.String()
		} else {
			logJson = string(rawLog)
		}

		blocks = append(blocks, block{
			Type: "section",
			Text: &text{
				Type: "mrkdwn",
				Text: fmt.Sprintf("%s*%s*\n```%s```\n```%s```", userTags, strings.ToUpper(log.Log.Severity), util.StripAnsi(log.Log.Message), logJson),
			},
		})

		blocks = append(blocks, block{
			Type: "context",
			Elements: []element{
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Timestamp:* %s", log.Log.Timestamp),
				},
			},
		})

		blocks = append(blocks, block{
			Type: "divider",
		})
	}

	message := slackMessage{
		Blocks: blocks,
	}

	payload, err := json.Marshal(&message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return payload, nil
}
