package railway

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

	if gqlConfig.BaseURL != "" {
		gqlConfig.client = graphql.NewClient(gqlConfig.BaseURL, httpClient)
	}

	return gqlConfig, nil
}
