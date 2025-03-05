package loki

import (
	"bytes"
	"encoding/json"
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
	streams := streams{
		Streams: []stream{},
	}

	for i := range logs {
		var serviceStream stream

		streamIndex := findServiceStream(logs[i].Tags.ServiceID, &streams)
		if streamIndex < 0 {
			log := logs[i]

			serviceStream = stream{
				Stream: map[string]string{
					"project_id":             log.Tags.ProjectID,
					"project_name":           log.Tags.ProjectName,
					"environment_id":         log.Tags.EnvironmentID,
					"environment_name":       log.Tags.EnvironmentName,
					"service_id":             log.Tags.ServiceID,
					"service_name":           log.Tags.ServiceName,
					"deployment_id":          log.Tags.DeploymentID,
					"deployment_instance_id": log.Tags.DeploymentInstanceID,
				},
				Values: [][]interface{}{},
			}

			streams.Streams = append(streams.Streams, serviceStream)
			streamIndex = len(streams.Streams) - 1
		}

		rawLog, err := logline.ReconstructLogLineLoki(logs[i])
		if err != nil {
			return fmt.Errorf("failed to reconstruct log object: %w", err)
		}

		streams.Streams[streamIndex].Values = append(streams.Streams[streamIndex].Values, rawLog)
	}

	encodedStreams, err := json.Marshal(streams)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.LokiIngestUrl, bytes.NewReader(encodedStreams))
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
