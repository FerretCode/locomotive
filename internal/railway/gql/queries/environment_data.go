package queries

type EnvironmentData struct {
	Environment struct {
		Deployments struct {
			Edges []struct {
				Node struct {
					ServiceID string `json:"serviceId"`
					ProjectID string `json:"projectId"`
					Status    string `json:"status"`
					ID        string `json:"id"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"deployments"`
		ProjectID string `json:"projectId"`
	} `json:"environment"`
}
