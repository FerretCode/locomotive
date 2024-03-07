package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/logline"
	"github.com/ferretcode/locomotive/railway"
)

func SendWebhook(logs []railway.EnvironmentLog, cfg *config.Config, client *http.Client) error {
	em := []embed{}

	for i := range logs {
		rawLog, err := logline.ReconstructLogLine(logs[i])
		if err != nil {
			return err
		}

		em = append(em, embed{
			Title:       strings.ToUpper(logs[i].Severity),
			Color:       getColor(logs[i].Severity),
			Description: fmt.Sprintf("```%s```", logs[i].Message),
			Fields: []field{{
				Name:   "⠀",
				Value:  fmt.Sprintf("```%s```", rawLog),
				Inline: false,
			}}},
		)

	}

	hook := hook{
		Username: "locomotive",
		Embeds:   em,
	}

	payload, err := json.Marshal(&hook)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.DiscordWebhookUrl, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("non success status code: %d", resp.StatusCode)
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
