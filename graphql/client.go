package graphql

import (
	"errors"
	"net/http"

	"github.com/hasura/go-graphql-client"
)

func NewClient(gqlConfig *GraphQLClient) (*GraphQLClient, error) {
	if gqlConfig == nil {
		return nil, errors.New("gqlConfig must not be nil")
	}

	if gqlConfig.AuthToken == "" {
		return nil, errors.New("auth token must not be empty")
	}

	httpClient := &http.Client{
		Transport: &authedTransport{
			token:   gqlConfig.AuthToken,
			wrapped: http.DefaultTransport,
		},
	}

	config := &GraphQLClient{}

	if gqlConfig.BaseURL != "" {
		config.client = graphql.NewClient(gqlConfig.BaseURL, httpClient)
	}

	if gqlConfig.BaseSubscriptionURL != "" {
		config.subscriptionClient = graphql.NewSubscriptionClient(gqlConfig.BaseSubscriptionURL).
			WithWebSocketOptions(graphql.WebsocketOptions{
				HTTPClient: httpClient,
			})

		config.subscriptionClient.WithProtocol(graphql.GraphQLWS)
	}

	return config, nil
}
