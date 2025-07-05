package http_logs

import (
	"time"

	"github.com/ferretcode/locomotive/internal/railway/gql/subscriptions"
)

type DeploymentHttpLogWithMetadata struct {
	Log       subscriptions.HttpLog
	Path      string
	Timestamp time.Time
	Metadata  DeploymentHttpLogMetadata
}

type DeploymentHttpLogMetadata struct {
	ProjectID   string `json:"projectId"`
	ProjectName string `json:"projectName"`

	EnvironmentID   string `json:"environmentId"`
	EnvironmentName string `json:"environmentName"`

	ServiceID   string `json:"serviceId"`
	ServiceName string `json:"serviceName"`

	DeploymentID string `json:"deploymentId"`
}

var flushInterval = time.Second
