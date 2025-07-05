package reconstruct_slack

type slackMessage struct {
	Blocks []block `json:"blocks"`
}

type block struct {
	Type     string    `json:"type"`
	Text     *text     `json:"text,omitempty"`
	Elements []element `json:"elements,omitempty"`
}

type text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type element struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
