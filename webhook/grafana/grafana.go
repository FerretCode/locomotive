package grafana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"

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
	firstLog := logs[0]

	streams := streams{
		Streams: []stream{
			{
				Stream: map[string]string{
					"project_id":             firstLog.Tags.ProjectID,
					"project_name":           firstLog.Tags.ProjectName,
					"environment_id":         firstLog.Tags.EnvironmentID,
					"environment_name":       firstLog.Tags.EnvironmentName,
					"service_id":             firstLog.Tags.ServiceID,
					"service_name":           firstLog.Tags.ServiceName,
					"deployment_id":          firstLog.Tags.DeploymentID,
					"deployment_instance_id": firstLog.Tags.DeploymentInstanceID,
				},
				Values: [][]string{},
			},
		},
	}

	for i := range logs {
		rawLog, err := logline.ReconstructLogLine(logs[i])
		if err != nil {
			return fmt.Errorf("failed to reconstruct log object: %w", err)
		}

		time := strconv.FormatInt(logs[i].Timestamp.UnixNano(), 10)

		streams.Streams[0].Values = append(streams.Streams[0].Values, []string{time, string(rawLog)})
	}

	encodedStreams, err := json.Marshal(streams)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.GrafanaIngestUrl, bytes.NewReader(encodedStreams))
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
