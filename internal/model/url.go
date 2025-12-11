package model

import "github.com/google/uuid"

type URL struct {
	ID        string     `json:"id,omitempty"`
	Original  string     `json:"url,omitempty"`
	Shortened string     `json:"result,omitempty"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Deleted   *bool      `json:"-"`
}

type BatchURL struct {
	ID        string `json:"id,omitempty"`
	CorrID    string `json:"correlation_id,omitempty"`
	Original  string `json:"original_url,omitempty"`
	Shortened string `json:"short_url,omitempty"`
}
