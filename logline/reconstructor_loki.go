package logline

import (
	"fmt"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/ferretcode/locomotive/railway"
	"github.com/ferretcode/locomotive/util"
)

func ReconstructLogLinesLoki(logs []railway.EnvironmentLog) ([]byte, error) {
	var jsonObject []byte

	jsonObject = append(jsonObject, []byte(`{"streams": [`)...)

	for i := range logs {
		logObject, err := ReconstructLogLineLoki(logs[i])
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
func ReconstructLogLineLoki(log railway.EnvironmentLog) ([]byte, error) {
	var err error

	jsonObject := []byte("{\"stream\":{},\"values\":[[0,1,{}]]}")

	labels := map[string]string{
		"project_id":             log.Tags.ProjectID,
		"project_name":           log.Tags.ProjectName,
		"environment_id":         log.Tags.EnvironmentID,
		"environment_name":       log.Tags.EnvironmentName,
		"service_id":             log.Tags.ServiceID,
		"service_name":           log.Tags.ServiceName,
		"deployment_id":          log.Tags.DeploymentID,
		"deployment_instance_id": log.Tags.DeploymentInstanceID,
	}

	for label, value := range labels {
		jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(value)), "stream", label)
		if err != nil {
			return nil, fmt.Errorf("failed to append label to stream object: %w", err)
		}
	}

	slogAttributes := []byte("{}")

	cleanMessage := AnsiEscapeRe.ReplaceAllString(log.Message, "")

	for i := range log.Attributes {
		if log.Attributes[i].Key == "time" || log.Attributes[i].Key == "level" {
			continue
		}

		slogAttributes, err = jsonparser.Set(slogAttributes, []byte(util.QuoteIfNeeded(log.Attributes[i].Value)), log.Attributes[i].Key)
		if err != nil {
			return nil, fmt.Errorf("failed to append json attribute to object: %w", err)
		}
	}

	// only use Railway timestamp
	timeStamp := strconv.FormatInt(log.Timestamp.UnixNano(), 10)

	// set severity in all situations for backwards compatibility
	// railway already normilizes the level attribute into the severity field, or vice versa
	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Severity)), "stream", "severity")
	if err != nil {
		return nil, fmt.Errorf("failed to append severity attribute to object: %w", err)
	}

	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Severity)), "stream", "level")
	if err != nil {
		return nil, fmt.Errorf("failed to append severity attribute to object: %w", err)
	}

	jsonObject, err = jsonparser.Set(jsonObject, []byte(util.QuoteIfNeeded(timeStamp)), "values", "[0]", "[0]")
	if err != nil {
		return nil, fmt.Errorf("failed to set timestamp in values slice: %w", err)
	}

	jsonObject, err = jsonparser.Set(jsonObject, []byte(util.QuoteIfNeeded(cleanMessage)), "values", "[0]", "[1]")
	if err != nil {
		return nil, fmt.Errorf("failed to set message in values slice: %w", err)
	}

	jsonObject, err = jsonparser.Set(jsonObject, slogAttributes, "values", "[0]", "[2]")
	if err != nil {
		return nil, fmt.Errorf("failed to set slog attributes in values slice: %w", err)
	}

	return jsonObject, nil
}
