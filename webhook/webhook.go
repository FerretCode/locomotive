package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"unsafe"

	"github.com/buger/jsonparser"
	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/graphql"
)

var acceptedStatusCodes = []int{
	http.StatusOK,
	http.StatusNoContent,
	http.StatusAccepted,
	http.StatusCreated,
}

func SendGenericWebhook(log *graphql.EnvironmentLog, cfg *config.Config) (err error) {
	if len(log.MessageRaw) == 0 {
		return nil
	}

	jsonObject := json.RawMessage("{}")

	jsonObject, err = jsonparser.Set(jsonObject, log.MessageRaw, "message")
	if err != nil {
		return fmt.Errorf("failed to append message attribute to object: %w", err)
	}

	if len(log.Attributes) > 0 {
		for _, attr := range log.Attributes {
			jsonObject, err = jsonparser.Set(jsonObject, unsafe.Slice(unsafe.StringData(attr.Value), len(attr.Value)), attr.Key)
			if err != nil {
				return fmt.Errorf("failed to append json attribute to object: %w", err)
			}
		}
	} else {
		jsonObject, err = jsonparser.Set(jsonObject, log.TimestampRaw, "time")
		if err != nil {
			return fmt.Errorf("failed to append time attribute to object: %w", err)
		}

		jsonObject, err = jsonparser.Set(jsonObject, log.SeverityRaw, "severity")
		if err != nil {
			return fmt.Errorf("failed to append severity attribute to object: %w", err)
		}
	}

	req, err := http.NewRequest(http.MethodPost, cfg.IngestUrl, bytes.NewBuffer(jsonObject))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-ndjson")

	for key, value := range cfg.AdditionalHeaders {
		req.Header.Set(key, value)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}

	defer res.Body.Close()

	if !slices.Contains(acceptedStatusCodes, res.StatusCode) {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("non success status code: %d", res.StatusCode)
		}

		return errors.New(string(body))
	}

	return nil
}
