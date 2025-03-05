package loki

type streams struct {
	Streams []stream `json:"streams"`
}

type stream struct {
	Stream map[string]string `json:"stream"` // encoded json labels
	Values [][]interface{}   `json:"values"`
}
