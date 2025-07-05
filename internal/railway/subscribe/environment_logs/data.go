package environment_logs

import "github.com/ferretcode/locomotive/internal/railway/gql/subscriptions"

type EnvironmentLogWithMetadata struct {
	Log      subscriptions.EnvironmentLog
	Metadata EnvironmentLogMetadata
}

type EnvironmentLogMetadata struct {
	ProjectName     string `json:"projectName"`
	EnvironmentName string `json:"environmentName"`
	ServiceName     string `json:"serviceName"`
}
