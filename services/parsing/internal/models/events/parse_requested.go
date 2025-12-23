package events

import "time"

type ParseRequested struct {
	EventID        string    `json:"event_id"`
	OccurredAt     time.Time `json:"occurred_at"`
	CorrelationID  string    `json:"correlation_id"`
	ProductID      string    `json:"product_id,omitempty"`
	URL            string    `json:"url"`
	ScheduledAt    time.Time `json:"scheduled_at,omitempty"`
	Priority       int       `json:"priority,omitempty"`
}

