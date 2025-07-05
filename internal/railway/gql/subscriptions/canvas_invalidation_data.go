package subscriptions

type CanvasInvalidationSubscriptionPayload struct {
	Query     string                                   `json:"query"`
	Variables *CanvasInvalidationSubscriptionVariables `json:"variables"`
}

type CanvasInvalidationSubscriptionVariables struct {
	EnvironmentId string `json:"environmentId"`
}

type CanvasInvalidationData struct {
	Payload struct {
		Data struct {
			CanvasInvalidation struct {
				ID string `json:"id"`
			} `json:"canvasInvalidation"`
		} `json:"data"`
	} `json:"payload"`
	Type SubscriptionType `json:"type"`
}
