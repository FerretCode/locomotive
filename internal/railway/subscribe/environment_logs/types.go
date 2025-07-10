package environment_logs

import "github.com/ferretcode/locomotive/internal/railway/gql/subscriptions"

type EnvironmentLogWithMetadata struct {
	Log      subscriptions.EnvironmentLog
	Metadata EnvironmentLogMetadata
}

type EnvironmentLogMetadata struct {
	ProjectName string `json:"projectName"`
	ProjectID   string `json:"projectId"`

	EnvironmentName string `json:"environmentName"`
	EnvironmentID   string `json:"environmentId"`

	ServiceName string `json:"serviceName"`
	ServiceID   string `json:"serviceId"`

	DeploymentID         string `json:"deploymentId"`
	DeploymentInstanceId string `json:"deploymentInstanceId"`
}
