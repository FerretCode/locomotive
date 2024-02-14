package railway

import (
	"context"
	"fmt"
	"os"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/graphql"
)

type DeploymentsResponse struct {
	Data struct {
		Deployments struct {
			Edges []struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"deployments"`
	} `json:"data"`
}

func GetDeployments(ctx context.Context, client graphql.GraphQLClient, cfg config.Config) (DeploymentsResponse, error) {
	query := fmt.Sprintf(`
		query GetDeployments {
			deployments(
				input: {projectId: "%s", serviceId: "%s"}
			) {
				edges {
					node {
						id
						meta
					}
				}
			}
		}
		`,
		os.Getenv("RAILWAY_PROJECT_ID"), // deprecated
		cfg.Train,
	)

	deploymentsResponse := DeploymentsResponse{}

	err := client.DoQuery(query, nil, &deploymentsResponse, cfg)

	if err != nil {
		return deploymentsResponse, err
	}

	return deploymentsResponse, nil
}
