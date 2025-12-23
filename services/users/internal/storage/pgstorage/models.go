package pgstorage

import "time"

type User struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
}

type UserURL struct {
	ID            string
	UserID        string
	URL           string
	NormalizedURL string
	PollingIntervalSeconds int
	CreatedAt     time.Time
}
