package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ferretcode/locomotive/graphql"
)

func SendGenericWebhook(log graphql.Log) error {
	fmt.Println(log)

	data, err := json.Marshal(log)

	if err != nil {
		return nil
	}

	fmt.Println(string(data))

	req, err := http.NewRequest(
		"POST",
		os.Getenv("INGEST_URL"),
		bytes.NewBuffer(data),
	)

	if err != nil {
		return err
	}

	headers := os.Getenv("ADDITIONAL_HEADERS")

	if headers != "" {
		fields := strings.Split(headers, ";")

		for _, field := range fields {
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
