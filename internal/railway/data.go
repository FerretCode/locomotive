package railway

import (
	"github.com/hasura/go-graphql-client"
)

type GraphQLClient struct {
	AuthToken           string
	BaseSubscriptionURL string
	BaseURL             string
	Client              *graphql.Client
}
