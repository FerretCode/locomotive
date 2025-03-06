package logline

import (
	"fmt"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/ferretcode/locomotive/railway"
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

	fmt.Println(string(jsonObject))

	return jsonObject, nil
}

// reconstruct a single log into a format acceptable by loki
func ReconstructLogLineLoki(log railway.EnvironmentLog) ([]byte, error) {
	var err error
	jsonObject := []byte("{}")

	jsonObject, err = jsonparser.Set(jsonObject, []byte("{}"), "stream")
	if err != nil {
		return jsonObject, err
	}

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
		jsonObject, err = jsonparser.Set(jsonObject, []byte(fmt.Sprintf("\"%s\"", []byte(value))), "stream", label)
		if err != nil {
			return jsonObject, err
		}
	}

	slogAttributes := []byte("{}")

	cleanMessage := AnsiEscapeRe.ReplaceAllString(log.Message, "")

	for i := range log.Attributes {
		slogAttributes, err = jsonparser.Set(slogAttributes, []byte(log.Attributes[i].Value), log.Attributes[i].Key)
		if err != nil {
			return nil, fmt.Errorf("failed to append json attribute to object: %w", err)
		}
	}

	fmt.Println(slogAttributes)

	// only use Railway timestamp
	timeStamp := []byte(fmt.Sprintf("%d", log.Timestamp.UnixNano()))

	// set severity in all situations for backwards compatibility
	// railway already normilizes the level attribute into the severity field, or vice versa
	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Severity)), "stream", "severity")
	if err != nil {
		return nil, fmt.Errorf("failed to append severity attribute to object: %w", err)
	}

	values := []byte(
		fmt.Sprintf("[[\"%s\", \"%s\", %s]]", string(timeStamp), cleanMessage, string(slogAttributes)),
	)

	jsonObject, err = jsonparser.Set(jsonObject, values, "values")
	if err != nil {
		return jsonObject, err
	}

	return jsonObject, nil
}
