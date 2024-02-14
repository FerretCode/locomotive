package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/graphql"
)

func SendGenericWebhook(log graphql.Log, cfg config.Config) error {
	var logs []string

	if log.Message != "" && len(log.Attributes) > 0 {
		message := make(map[string]interface{})

		for _, attr := range log.Attributes {
			message[attr.Key] = attr.Value
		}

		message["message"] = log.Message

		bytes, err := json.Marshal(message)

		if err != nil {
			return err
		}

		log.Message = string(bytes)
		log.Attributes = nil

		logs = append(logs, log.Message)
	} else {
		data, err := json.Marshal(log)

		if err != nil {
			return err
		}

		logs = append(logs, string(data))
	}

	body, err := json.Marshal(logs)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		cfg.IngestUrl,
		bytes.NewBuffer(body),
	)

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	headers := cfg.AdditionalHeaders

	if len(headers) > 0 {
		for _, field := range headers {
			key := field[:strings.Index(field, "=")]
			value := field[len(key)+1:]

			req.Header.Add(key, value)
		}
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode > 300 {
		body, err := io.ReadAll(res.Body)

		if err != nil {
			return err
		}

		return errors.New(string(body))
	}

	return nil
}
