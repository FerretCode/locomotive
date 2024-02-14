package railway

import (
	"context"
	"fmt"

	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/graphql"
)

const (
	ALL   = "all"
	ERROR = "err"
	WARN  = "warn"
	INFO  = "info"
)

const (
	SEVERITY_INFO  = "info"
	SEVERITY_ERROR = "err"
	SEVERITY_WARN  = "warn"
)

type LogsResponse struct {
	Data struct {
		DeploymentLogs []RailwayLog `json:"deploymentLogs"`
	} `json:"data"`
}

type RailwayLog struct {
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
}

func GetLogs(ctx context.Context, client graphql.GraphQLClient, deploymentId string, cfg config.Config) (LogsResponse, error) {
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

	err := client.DoQuery(query, nil, &logsResponse, cfg)

	if err != nil {
		return logsResponse, err
	}

	return logsResponse, nil
}
