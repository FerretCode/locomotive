package graphql

import "encoding/json"

type SubscriptionLogResponse struct {
	EnvironmentLogs []EnvironmentLog `json:"environmentLogs"`
}

type EnvironmentLog struct {
	Message    string
	MessageRaw json.RawMessage `json:"message,string"`

	Severity    string
	SeverityRaw json.RawMessage `json:"severity,string"`

	Tags map[string]string `json:"tags,string"`

	// Timestamp    string `json:"timestamp"`
	TimestampRaw json.RawMessage `json:"timestamp,string"`

	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GraphQLClient struct {
	BaseSubscriptionURL string
	BaseURL             string
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}
