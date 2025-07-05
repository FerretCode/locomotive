package reconstruct_discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ferretcode/locomotive/internal/logline/reconstructor/reconstruct_json"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
	"github.com/ferretcode/locomotive/internal/util"
)

func EnvironmentLogLines(logs []environment_logs.EnvironmentLogWithMetadata, prettyJson bool) ([]byte, error) {
	em := []embed{}

	for i := range logs {
		rawLog, err := reconstruct_json.EnvironmentLogLineJson(logs[i])
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct log object: %w", err)
		}

		var description string

		if prettyJson {
			buf := &bytes.Buffer{}

			if err := json.Indent(buf, rawLog, "", "  "); err != nil {
				return nil, fmt.Errorf("failed to indent json log object: %w", err)
			}

			description = fmt.Sprintf(descriptionFormat, util.StripAnsi(logs[i].Log.Message), buf)
		} else {
			description = fmt.Sprintf(descriptionFormat, util.StripAnsi(logs[i].Log.Message), rawLog)
		}

		em = append(em, embed{
			Title:       strings.ToUpper(logs[i].Log.Severity),
			Color:       getColor(logs[i].Log.Severity),
			Description: description,
			Timestamp:   logs[i].Log.Timestamp,
		})
	}

	hook := hook{
		Username: "locomotive",
		Embeds:   em,
	}

	payload, err := json.Marshal(&hook)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hook: %w", err)
	}

	return payload, nil
}
