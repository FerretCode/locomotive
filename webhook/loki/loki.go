package loki

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/logline"
	"github.com/ferretcode/locomotive/railway"
)

var acceptedStatusCodes = []int{
	http.StatusOK,
	http.StatusNoContent,
	http.StatusAccepted,
	http.StatusCreated,
}

func SendWebhook(logs []railway.EnvironmentLog, cfg *config.Config, client *http.Client) error {
	jsonLogs, err := logline.ReconstructLogLinesLoki(logs)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.LokiIngestUrl, bytes.NewReader(jsonLogs))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Keep-Alive", "timeout=5, max=1000")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}

	defer res.Body.Close()

	if !slices.Contains(acceptedStatusCodes, res.StatusCode) {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("non success status code: %d", res.StatusCode)
		}

		return fmt.Errorf("non success status code: %d; with body: %s", res.StatusCode, body)
	}

	return nil
}

func findServiceStream(serviceId string, streams *streams) int {
	for i, stream := range streams.Streams {
		if stream.Stream["service_id"] == serviceId {
			return i
		}
	}

	return -1
}
