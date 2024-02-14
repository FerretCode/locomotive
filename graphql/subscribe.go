package graphql

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ferretcode/locomotive/config"
	"github.com/hasura/go-graphql-client"
)

type authedTransport struct {
	token   string
	wrapped http.RoundTripper
}

type SubscriptionLogResponse struct {
	EnvironmentLogs []struct {
		Message    string            `json:"message"`
		Severity   string            `json:"severity"`
		Tags       map[string]string `json:"tags"`
		Timestamp  string            `json:"timestamp"`
		Attributes []Attribute       `json:"attributes"`
	} `json:"environmentLogs"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return t.wrapped.RoundTrip(req)
}

func (g *GraphQLClient) SubscribeToLogs(newLog chan SubscriptionLogResponse, cfg config.Config) error {
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

	stderr := log.New(os.Stderr, "", 0)

	_, err := client.Exec(query, variables, func(message []byte, err error) error {
		if err != nil {
			stderr.Println(err)

			return nil
		}

		data := SubscriptionLogResponse{}

		err = json.Unmarshal(message, &data)

		if err != nil {
			stderr.Println(err)

			return nil
		}

		newLog <- data

		return nil
	})

	if err != nil {
		return err
	}

	err = client.Run()

	if err != nil {
		return err
	}

	return nil
}
