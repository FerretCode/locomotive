package queries

type Deployment struct {
	Deployment struct {
		Service struct {
			Name    string `json:"name"`
			ID      string `json:"id"`
			Project struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"project"`
		} `json:"service"`
		Environment struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"environment"`
	} `json:"deployment"`
}
