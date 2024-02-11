package railway

import (
	"context"
	"fmt"

	"github.com/ferretcode/locomotive/graphql"
)

const (
	ALL   = "all"
	ERROR = "error"
	INFO  = "info"
)

const (
	SEVERITY_INFO  = "info"
	SEVERITY_ERROR = "error"
)

type LogsResponse struct {
	Data struct {
		DeploymentLogs []struct {
			Message   string `json:"message"`
			Severity  string `json:"severity"`
			Timestamp string `json:"timestamp"`
		} `json:"deploymentLogs"`
	} `json:"data"`
}

func GetLogs(ctx context.Context, client graphql.GraphQLClient, deploymentId string) (LogsResponse, error) {
	query := fmt.Sprintf(
		`
		query MyQuery {
			deploymentLogs(deploymentId: "%s") {
				message
				severity
				timestamp
			}
		}
		`,
		deploymentId,
	)

	logsResponse := LogsResponse{}

	err := client.DoQuery(query, nil, &logsResponse)

	if err != nil {
		return logsResponse, err
	}

	return logsResponse, nil
}
