package queries

type ProjectData struct {
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
