package graphql

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"unsafe"

	"github.com/ferretcode/locomotive/config"
	"github.com/hasura/go-graphql-client"
)

func (g *GraphQLClient) SubscribeToLogs(logFunc func(log *EnvironmentLog, err error), cfg *config.Config) error {
	client := graphql.NewSubscriptionClient(g.BaseSubscriptionURL).
		WithWebSocketOptions(graphql.WebsocketOptions{
			HTTPClient: &http.Client{
				Transport: &authedTransport{
					token:   cfg.RailwayApiKey,
					wrapped: http.DefaultTransport,
				},
			},
		})

	client.WithProtocol(graphql.GraphQLWS)

	defer client.Close()

	// yucky
	query := "subscription streamEnvironmentLogs($environmentId: String!, $filter: String, $beforeLimit: Int!, $beforeDate: String, $anchorDate: String, $afterDate: String, $afterLimit: Int) {\n  environmentLogs(\n    environmentId: $environmentId\n    filter: $filter\n    beforeDate: $beforeDate\n    anchorDate: $anchorDate\n    afterDate: $afterDate\n    beforeLimit: $beforeLimit\n    afterLimit: $afterLimit\n  ) {\n    ...LogFields\n  }\n}\n\nfragment LogFields on Log {\n  timestamp\n  message\n  severity\n  tags {\n    projectId\n    environmentId\n    pluginId\n    serviceId\n    deploymentId\n    deploymentInstanceId\n    snapshotId\n  }\n  attributes {\n    key\n    value\n  }\n}"

	variables := map[string]interface{}{
		"environmentId": cfg.EnvironmentId,
		"beforeDate":    time.Now().Format(time.RFC3339Nano),
		"beforeLimit":   0,
		"filter":        "@service:" + cfg.Train,
	}

	if _, err := client.Exec(query, variables, func(message []byte, err error) error {
		if err != nil {
			logFunc(nil, err)
		}

		data := SubscriptionLogResponse{}

		if err := json.Unmarshal(message, &data); err != nil {
			logFunc(nil, err)
		}

		if len(data.EnvironmentLogs) == 0 {
			return nil
		}

		for i := range data.EnvironmentLogs {
			data.EnvironmentLogs[i].Message, err = strconv.Unquote(
				unsafe.String(unsafe.SliceData(data.EnvironmentLogs[i].MessageRaw), len(data.EnvironmentLogs[i].MessageRaw)),
			)
			if err != nil {
				logFunc(nil, err)
				return nil
			}

			data.EnvironmentLogs[i].Severity, err = strconv.Unquote(
				unsafe.String(unsafe.SliceData(data.EnvironmentLogs[i].SeverityRaw), len(data.EnvironmentLogs[i].SeverityRaw)),
			)
			if err != nil {
				logFunc(nil, err)
				return nil
			}

			logFunc(&data.EnvironmentLogs[i], nil)
		}

		return nil
	}); err != nil {
		return err
	}

	if err := client.Run(); err != nil {
		return err
	}

	return nil
}
