package reconstruct_loki

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/http_logs"
)

func HttpLogLines(logs []http_logs.DeploymentHttpLogWithMetadata) ([]byte, error) {
	var jsonObject []byte

	jsonObject = append(jsonObject, []byte(`{"streams": [`)...)

	for i := range logs {
		logObject, err := HttpLogLine(logs[i])
		if err != nil {
			return nil, err
		}

		jsonObject = append(jsonObject, logObject...)

		if (i + 1) < len(logs) {
			jsonObject = append(jsonObject, []byte(`,`)...)
		}
	}

	jsonObject = append(jsonObject, []byte(`]}`)...)

	return jsonObject, nil
}

// reconstruct a single log into a format acceptable by loki
func HttpLogLine(log http_logs.DeploymentHttpLogWithMetadata) (jsonObject []byte, err error) {
	jsonObject = []byte("{\"stream\":{},\"values\":[[0,1,{}]]}")

	labels := map[string]string{
		"project_id":       log.Metadata.ProjectID,
		"project_name":     log.Metadata.ProjectName,
		"environment_id":   log.Metadata.EnvironmentID,
		"environment_name": log.Metadata.EnvironmentName,
		"service_id":       log.Metadata.ServiceID,
		"service_name":     log.Metadata.ServiceName,
		"deployment_id":    log.Metadata.DeploymentID,
	}

	for label, value := range labels {
		jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(value)), "stream", label)
		if err != nil {
			return nil, fmt.Errorf("failed to append label to stream object: %w", err)
		}
	}

	// only use Railway timestamp
	timeStamp := strconv.FormatInt(log.Timestamp.UnixNano(), 10)

	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(timeStamp)), "values", "[0]", "[0]")
	if err != nil {
		return nil, fmt.Errorf("failed to set timestamp in values slice: %w", err)
	}

	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Path)), "values", "[0]", "[1]")
	if err != nil {
		return nil, fmt.Errorf("failed to set path in values slice: %w", err)
	}

	logAttributes, err := json.Marshal(log.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal log attributes: %w", err)
	}

	logAttributes = jsonparser.Delete(logAttributes, "timestamp")
	logAttributes = jsonparser.Delete(logAttributes, "path")

	jsonObject, err = jsonparser.Set(jsonObject, logAttributes, "values", "[0]", "[2]")
	if err != nil {
		return nil, fmt.Errorf("failed to set log attributes in values slice: %w", err)
	}

	return jsonObject, nil
}
