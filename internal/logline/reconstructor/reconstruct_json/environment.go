package reconstruct_json

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/buger/jsonparser"

	"github.com/ferretcode/locomotive/internal/railway/subscribe/environment_logs"
	"github.com/ferretcode/locomotive/internal/util"
)

// reconstruct multiple logs into a raw json array containing json log lines
func EnvironmentLogLinesJson(logs []environment_logs.EnvironmentLogWithMetadata) ([]byte, error) {
	var jsonObject []byte

	jsonObject = append(jsonObject, []byte(`[`)...)

	for i := range logs {
		logObject, err := EnvironmentLogLineJson(logs[i])
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
func EnvironmentLogLineJson(log environment_logs.EnvironmentLogWithMetadata) ([]byte, error) {
	jsonObject := []byte("{}")

	cleanMessage := util.StripAnsi(log.Log.Message)

	jsonObject, err := jsonparser.Set(jsonObject, []byte(strconv.Quote(cleanMessage)), "message")
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

	for i := range log.Log.Attributes {
		jsonObject, err = jsonparser.Set(jsonObject, []byte(log.Log.Attributes[i].Value), log.Log.Attributes[i].Key)
		if err != nil {
			return nil, fmt.Errorf("failed to append json attribute to object: %w", err)
		}
	}

	// check for a timestamp attribute
	timeStampAttr, hasTimeStampAttr := environment_logs.AttributesHasKeys(log.Log.Attributes, commonTimeStampAttributes)

	if !hasTimeStampAttr {
		timeStampAttr = log.Log.Timestamp.Format(time.RFC3339Nano)
	}

	// add the timestamps that common logging services like betterstack and axiom expect
	// ref: https://betterstack.com/docs/logs/http-rest-api/#sending-timestamps
	// ref: https://axiom.co/docs/send-data/ingest#timestamp-field
	for i := range commonTimeStampAttributes {
		jsonObject, err = jsonparser.Set(jsonObject, []byte(util.QuoteIfNeeded(timeStampAttr)), commonTimeStampAttributes[i])
		if err != nil {
			return nil, fmt.Errorf("failed to append %s attribute to object: %w", commonTimeStampAttributes[i], err)
		}
	}

	// set severity in all situations for backwards compatibility
	// railway already normalizes the level attribute into the severity field, or vice versa
	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Log.Severity)), "severity")
	if err != nil {
		return nil, fmt.Errorf("failed to append severity attribute to object: %w", err)
	}

	return jsonObject, nil
}
