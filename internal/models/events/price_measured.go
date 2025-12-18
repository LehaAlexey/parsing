package events

import "time"

type PriceMeasured struct {
	EventID       string    `json:"event_id"`
	OccurredAt    time.Time `json:"occurred_at"`
	CorrelationID string    `json:"correlation_id"`
	ProductID     string    `json:"product_id,omitempty"`
	Price         int64     `json:"price"`
	Currency      string    `json:"currency"`
	ParsedAt      time.Time `json:"parsed_at"`
	SourceURL     string    `json:"source_url"`
	MetaHash      string    `json:"meta_hash,omitempty"`
}

