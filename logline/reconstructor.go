package logline

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/buger/jsonparser"
	"github.com/ferretcode/locomotive/graphql"
)

func ReconstructLogLine(log *graphql.EnvironmentLog) (*[]byte, error) {
	jsonObject := []byte("{}")

	jsonObject, err := jsonparser.Set(jsonObject, log.MessageRaw, "message")
	if err != nil {
		return nil, fmt.Errorf("failed to append message attribute to object: %w", err)
	}

	metadata, err := json.Marshal(&log.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata object: %w", err)
	}

	jsonObject, err = jsonparser.Set(jsonObject, metadata, "_metadata")
	if err != nil {
		return nil, fmt.Errorf("failed to append metadata attribute to object: %w", err)
	}

	if len(log.Attributes) > 0 {
		for _, attr := range log.Attributes {
			jsonObject, err = jsonparser.Set(jsonObject, unsafe.Slice(unsafe.StringData(attr.Value), len(attr.Value)), attr.Key)
			if err != nil {
				return nil, fmt.Errorf("failed to append json attribute to object: %w", err)
			}
		}
	} else {
		jsonObject, err = jsonparser.Set(jsonObject, log.TimestampRaw, "time")
		if err != nil {
			return nil, fmt.Errorf("failed to append time attribute to object: %w", err)
		}

		jsonObject, err = jsonparser.Set(jsonObject, log.SeverityRaw, "severity")
		if err != nil {
			return nil, fmt.Errorf("failed to append severity attribute to object: %w", err)
		}
	}

	return &jsonObject, nil
}
