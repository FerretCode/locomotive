package graphql

import (
	"encoding/json"

	"github.com/hasura/go-graphql-client"
)

type SubscriptionLogResponse struct {
	EnvironmentLogs []EnvironmentLog `json:"environmentLogs"`
}

type EnvironmentLog struct {
	Message    string
	MessageRaw json.RawMessage `json:"message"`

	Severity    string
	SeverityRaw json.RawMessage `json:"severity"`

	Tags Tags `json:"tags"`

	TimestampRaw json.RawMessage `json:"timestamp"`

	Attributes []Attribute `json:"attributes"`

	Metadata *Metadata
}

type Tags struct {
	ServiceId string `json:"serviceId"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Metadata struct {
	ServiceId       string `json:"serviceId"`
	ServiceName     string `json:"serviceName"`
	EnvironmentId   string `json:"environmentId"`
	EnvironmentName string `json:"environmentName"`
}

type GraphQLClient struct {
	AuthToken           string
	BaseSubscriptionURL string
	BaseURL             string
	client              *graphql.Client
	subscriptionClient  *graphql.SubscriptionClient
}

type Environment struct {
	Environment struct {
		ProjectID string `json:"projectId"`
	} `json:"environment"`
}

type Project struct {
	Project struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Description  string `json:"description"`
		Environments struct {
			Edges []struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"environments"`
		Services struct {
			Edges []struct {
				Node struct {
					ID               string `json:"id"`
					Name             string `json:"name"`
					ServiceInstances struct {
						Edges []struct {
							Node struct {
								EnvironmentID string `json:"environmentId"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"serviceInstances"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"services"`
	} `json:"project"`
}
