package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UserURL struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	URL           string    `json:"url"`
	NormalizedURL string    `json:"normalized_url"`
	PollingIntervalSeconds int `json:"polling_interval_seconds"`
	CreatedAt     time.Time `json:"created_at"`
}
