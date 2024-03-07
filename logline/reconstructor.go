package logline

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/ferretcode/locomotive/railway"
)

var commonTimeStampAttributes = []string{"time", "_time", "timestamp", "ts", "datetime", "dt"}

// reconstruct multiple logs into a raw json array containing json log lines
func ReconstructLogLines(logs []railway.EnvironmentLog) ([]byte, error) {
	var jsonObject []byte

	jsonObject = append(jsonObject, []byte(`[`)...)

	for i := range logs {
		logObject, err := ReconstructLogLine(logs[i])
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
func ReconstructLogLine(log railway.EnvironmentLog) ([]byte, error) {
	jsonObject := []byte("{}")

	jsonObject, err := jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Message)), "message")
	if err != nil {
		return nil, fmt.Errorf("failed to append message attribute to object: %w", err)
	}

	metadata, err := json.Marshal(log.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata object: %w", err)
	}

	jsonObject, err = jsonparser.Set(jsonObject, metadata, "_metadata")
	if err != nil {
		return nil, fmt.Errorf("failed to append metadata attribute to object: %w", err)
	}

	if len(log.Attributes) > 0 {
		for i := range log.Attributes {
			jsonObject, err = jsonparser.Set(jsonObject, []byte(log.Attributes[i].Value), log.Attributes[i].Key)
			if err != nil {
				return nil, fmt.Errorf("failed to append json attribute to object: %w", err)
			}
		}

		// add the timestamps that common logging services like betterstack and axiom expect
		// ref: https://betterstack.com/docs/logs/http-rest-api/#sending-timestamps
		// ref: https://axiom.co/docs/send-data/ingest#timestamp-field
		// use the first found timestamp from structured logging to set all other common timestamp attributes
		if value, hasKey := railway.AttributesHasKeys(log.Attributes, commonTimeStampAttributes); hasKey {
			timeStamp := []byte(value)

			for i := range commonTimeStampAttributes {
				jsonObject, err = jsonparser.Set(jsonObject, timeStamp, commonTimeStampAttributes[i])
				if err != nil {
					return nil, fmt.Errorf("failed to append %s attribute to object: %w", commonTimeStampAttributes[i], err)
				}
			}
		}

		return jsonObject, nil
	}

	// add the timestamps that common logging services like betterstack and axiom expect
	// ref: https://betterstack.com/docs/logs/http-rest-api/#sending-timestamps
	// ref: https://axiom.co/docs/send-data/ingest#timestamp-field
	// use the timestamp set by Railway to set all other common timestamp attributes

	timeStamp := []byte(strconv.Quote(log.Timestamp.Format(time.RFC3339Nano)))

	for i := range commonTimeStampAttributes {
		jsonObject, err = jsonparser.Set(jsonObject, timeStamp, commonTimeStampAttributes[i])
		if err != nil {
			return nil, fmt.Errorf("failed to append %s attribute to object: %w", commonTimeStampAttributes[i], err)
		}
	}

	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Severity)), "severity")
	if err != nil {
		return nil, fmt.Errorf("failed to append severity attribute to object: %w", err)
	}

	return jsonObject, nil
}
