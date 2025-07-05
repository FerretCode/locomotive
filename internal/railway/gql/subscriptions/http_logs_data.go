package subscriptions

type HttpLogsSubscriptionPayload struct {
	Query     string                         `json:"query"`
	Variables *HttpLogsSubscriptionVariables `json:"variables"`
}

type HttpLogsSubscriptionVariables struct {
	AfterDate    *string `json:"afterDate"`
	AnchorDate   *string `json:"anchorDate"`
	BeforeDate   string  `json:"beforeDate"`
	BeforeLimit  int64   `json:"beforeLimit"`
	DeploymentId string  `json:"deploymentId"`
	Filter       string  `json:"filter"`
}

type HttpLog any

type HttpLogsData struct {
	ID      string           `json:"id"`
	Type    SubscriptionType `json:"type"`
	Payload struct {
		Data struct {
			// we are keeping this as any because Railway may add or remove fields
			HTTPLogs []HttpLog `json:"httpLogs"`
		} `json:"data"`
	} `json:"payload"`
}
