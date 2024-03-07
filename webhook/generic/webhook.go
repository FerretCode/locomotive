package generic

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
	jsonLogs, err := logline.ReconstructLogLines(logs)
	if err != nil {
		return fmt.Errorf("error reconstructing log(s) to json: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, cfg.IngestUrl, bytes.NewReader(jsonLogs))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Keep-Alive", "timeout=5, max=1000")

	for key, value := range cfg.AdditionalHeaders {
		req.Header.Set(key, value)
	}

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
