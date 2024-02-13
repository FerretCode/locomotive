package graphql

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-module/carbon"
	"github.com/hasura/go-graphql-client"
)

type authedTransport struct {
	token   string
	wrapped http.RoundTripper
}

type SubscriptionLogResponse struct {
	Data struct {
		EnvironmentLogs []struct {
			Message   string            `json:"message"`
			Severity  string            `json:"severity"`
			Tags      map[string]string `json:"tags"`
			Timestamp string            `json:"timestamp"`
		} `json:"environmentLogs"`
	} `json:"data"`
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return t.wrapped.RoundTrip(req)
}

func (g *GraphQLClient) SubscribeToLogs(newLog chan SubscriptionLogResponse) error {
	client := graphql.NewSubscriptionClient(g.BaseSubscriptionURL).
		WithWebSocketOptions(graphql.WebsocketOptions{
			HTTPClient: &http.Client{
				Transport: &authedTransport{
					token:   os.Getenv("RAILWAY_API_KEY"),
					wrapped: http.DefaultTransport,
				},
			},
		})

	defer client.Close()

	// yucky
	query := "subscription streamEnvironmentLogs($environmentId: String!, $filter: String, $beforeLimit: Int!, $beforeDate: String, $anchorDate: String, $afterDate: String, $afterLimit: Int) {\n  environmentLogs(\n    environmentId: $environmentId\n    filter: $filter\n    beforeDate: $beforeDate\n    anchorDate: $anchorDate\n    afterDate: $afterDate\n    beforeLimit: $beforeLimit\n    afterLimit: $afterLimit\n  ) {\n    ...LogFields\n  }\n}\n\nfragment LogFields on Log {\n  timestamp\n  message\n  severity\n  tags {\n    projectId\n    environmentId\n    pluginId\n    serviceId\n    deploymentId\n    deploymentInstanceId\n    snapshotId\n  }\n  attributes {\n    key\n    value\n  }\n}"

	variables := map[string]interface{}{
		"environmentId": os.Getenv("ENVIRONMENT_ID"),
		"beforeDate": carbon.
			NewCarbon().
			CreateFromTime(time.Now().UTC().Clock()).
			ToRfc3339NanoString("UTC"),
		"beforeLimit": 1000,
		"filter":      "@service:" + os.Getenv("TRAIN"),
	}

	fmt.Println(variables)

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

		fmt.Println(string(message))

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
