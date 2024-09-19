package slack

type SlackMessage struct {
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type     string    `json:"type"`
	Text     *Text     `json:"text,omitempty"`
	Elements []Element `json:"elements,omitempty"`
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Element struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
