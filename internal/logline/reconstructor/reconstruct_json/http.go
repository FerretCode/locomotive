package reconstruct_json

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/ferretcode/locomotive/internal/railway/subscribe/http_logs"
	"github.com/ferretcode/locomotive/internal/util"
)

// reconstruct multiple logs into a raw json array containing json log lines
func HttpLogLinesJson(logs []http_logs.DeploymentHttpLogWithMetadata) ([]byte, error) {
	var jsonObject []byte

	jsonObject = append(jsonObject, []byte(`[`)...)

	for i := range logs {
		logObject, err := HttpLogLineJson(logs[i])
		if err != nil {
			return nil, err
		}

		jsonObject = append(jsonObject, logObject...)

		if (i + 1) < len(logs) {
			jsonObject = append(jsonObject, []byte(`,`)...)
		}
	}

	jsonObject = append(jsonObject, []byte(`]`)...)

	return jsonObject, nil
}

// reconstruct a single log into a raw json object
func HttpLogLineJson(log http_logs.DeploymentHttpLogWithMetadata) ([]byte, error) {
	jsonObject, err := json.Marshal(log.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal log attributes: %w", err)
	}

	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Path)), "message")
	if err != nil {
		return nil, fmt.Errorf("failed to append message attribute to object: %w", err)
	}

	metadata, err := json.Marshal(log.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata object: %w", err)
	}

	jsonObject, err = jsonparser.Set(jsonObject, metadata, "_metadata")
	if err != nil {
		return nil, fmt.Errorf("failed to append metadata attribute to object: %w", err)
	}

	timeStamp := log.Timestamp.Format(time.RFC3339Nano)

	for i := range commonTimeStampAttributes {
		jsonObject, err = jsonparser.Set(jsonObject, []byte(util.QuoteIfNeeded(timeStamp)), commonTimeStampAttributes[i])
		if err != nil {
			return nil, fmt.Errorf("failed to append %s attribute to object: %w", commonTimeStampAttributes[i], err)
		}
	}

	return jsonObject, nil
}
