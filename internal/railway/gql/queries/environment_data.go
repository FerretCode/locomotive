package queries

import "time"

type EnvironmentData struct {
	Environment struct {
		Deployments struct {
			Edges []struct {
				Node struct {
					ServiceID string    `json:"serviceId"`
					ProjectID string    `json:"projectId"`
					Status    string    `json:"status"`
					CreatedAt time.Time `json:"createdAt"`
					ID        string    `json:"id"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"deployments"`
		ProjectID string `json:"projectId"`
	} `json:"environment"`
}
