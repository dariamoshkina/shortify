package model

type URL struct {
	ID        string `json:"-"`
	Original  string `json:"url,omitempty"`
	Shortened string `json:"result,omitempty"`
}
