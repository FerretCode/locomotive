package reconstruct_loki

import (
	"fmt"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
	"github.com/ferretcode/locomotive/internal/util"
)

func EnvironmentLogLines(logs []environment_logs.EnvironmentLogWithMetadata) ([]byte, error) {
	var jsonObject []byte

	jsonObject = append(jsonObject, []byte(`{"streams": [`)...)

	for i := range logs {
		logObject, err := EnvironmentLogLine(logs[i])
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
func EnvironmentLogLine(log environment_logs.EnvironmentLogWithMetadata) (jsonObject []byte, err error) {
	jsonObject = []byte("{\"stream\":{},\"values\":[[0,1,{}]]}")

	labels := map[string]string{
		"project_id":             log.Log.Tags.ProjectID,
		"project_name":           log.Metadata.ProjectName,
		"environment_id":         log.Log.Tags.EnvironmentID,
		"environment_name":       log.Metadata.EnvironmentName,
		"service_id":             log.Log.Tags.ServiceID,
		"service_name":           log.Metadata.ServiceName,
		"deployment_id":          log.Log.Tags.DeploymentID,
		"deployment_instance_id": log.Log.Tags.DeploymentInstanceID,
		// railway already normalizes the level attribute into the severity field, or vice versa
		"severity": log.Log.Severity,
		"level":    log.Log.Severity,
	}

	for label, value := range labels {
		jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(value)), "stream", label)
		if err != nil {
			return nil, fmt.Errorf("failed to append label to stream object: %w", err)
		}
	}

	cleanMessage := util.StripAnsi(log.Log.Message)

	for i := range log.Log.Attributes {
		if log.Log.Attributes[i].Key == "time" || log.Log.Attributes[i].Key == "level" {
			continue
		}

		jsonObject, err = jsonparser.Set(jsonObject, []byte(util.QuoteIfNeeded(log.Log.Attributes[i].Value)), "values", "[0]", "[2]", log.Log.Attributes[i].Key)
		if err != nil {
			return nil, fmt.Errorf("failed to append json attribute to object: %w", err)
		}
	}

	// only use Railway timestamp
	timeStamp := strconv.FormatInt(log.Log.Timestamp.UnixNano(), 10)

	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(timeStamp)), "values", "[0]", "[0]")
	if err != nil {
		return nil, fmt.Errorf("failed to set timestamp in values slice: %w", err)
	}

	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(cleanMessage)), "values", "[0]", "[1]")
	if err != nil {
		return nil, fmt.Errorf("failed to set message in values slice: %w", err)
	}

	return jsonObject, nil
}
