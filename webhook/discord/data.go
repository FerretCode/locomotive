package discord

import "time"

type field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type thumbnail struct {
	Url string `json:"url"`
}

type footer struct {
	Text     string `json:"text"`
	Icon_url string `json:"icon_url"`
}

type embed struct {
	Title       string    `json:"title"`
	Url         string    `json:"url"`
	Description string    `json:"description"`
	Color       int       `json:"color"`
	Thumbnail   thumbnail `json:"thumbnail"`
	Footer      footer    `json:"footer"`
	Fields      []field   `json:"fields"`
	Timestamp   time.Time `json:"timestamp"`
	Author      author    `json:"author"`
}

type author struct {
	Name     string `json:"name"`
	Icon_URL string `json:"icon_url"`
	Url      string `json:"url"`
}

type attachment struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	Filename    string `json:"filename"`
}

type hook struct {
	Username    string       `json:"username"`
	Avatar_url  string       `json:"avatar_url"`
	Content     string       `json:"content"`
	Embeds      []embed      `json:"embeds"`
	Attachments []attachment `json:"attachments"`
}
