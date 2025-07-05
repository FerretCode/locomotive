package config

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (h *AdditionalHeaders) UnmarshalText(envByte []byte) error {
	if h == nil {
		return fmt.Errorf("AdditionalHeaders is nil")
	}

	envString := string(envByte)
	headers := make(map[string]string)

	headerPairs := strings.SplitN(envString, ";", 2)

	for _, header := range headerPairs {
		keyValue := strings.SplitN(header, "=", 2)

		if len(keyValue) != 2 {
			return fmt.Errorf("header key value pair must be in format k=v")
		}

		headers[strings.TrimSpace(keyValue[0])] = strings.TrimSpace(keyValue[1])
	}

	*h = headers

	return nil
}

func (t *Trains) UnmarshalText(envByte []byte) error {
	if t == nil {
		return fmt.Errorf("Train is nil")
	}

	envString := string(envByte)

	trains := strings.Split(envString, ",")

	for _, train := range trains {
		train = strings.TrimSpace(train)

		if train == "" {
			continue
		}

		if _, err := uuid.Parse(train); err != nil {
			return fmt.Errorf("invalid train: \"%s\"; must be a valid uuid", train)
		}

		*t = append(*t, train)
	}

	return nil
}
