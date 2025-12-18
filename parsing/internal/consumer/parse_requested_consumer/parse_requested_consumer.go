package parse_requested_consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/LehaAlexey/Parsing/internal/models/events"
	"github.com/segmentio/kafka-go"
)

type Processor interface {
	Handle(ctx context.Context, req *events.ParseRequested) error
}

type Config struct {
	Brokers []string
	GroupID string
	Topic   string
}

type ParseRequestedConsumer struct {
	cfg       Config
	processor Processor
}

func New(cfg Config, processor Processor) *ParseRequestedConsumer {
	return &ParseRequestedConsumer{cfg: cfg, processor: processor}
}

func (c *ParseRequestedConsumer) Consume(ctx context.Context) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           c.cfg.Brokers,
		GroupID:           c.cfg.GroupID,
		Topic:             c.cfg.Topic,
		HeartbeatInterval: 3 * time.Second,
		SessionTimeout:    30 * time.Second,
		MinBytes:          1,
		MaxBytes:          10e6,
	})
	defer r.Close()

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			slog.Error("parse_requested_consumer: read message", "error", err.Error())
			continue
		}

		var req events.ParseRequested
		if err := json.Unmarshal(msg.Value, &req); err != nil {
			slog.Error("parse_requested_consumer: unmarshal", "error", err.Error())
			continue
		}

		if err := c.processor.Handle(ctx, &req); err != nil {
			slog.Error("parse_requested_consumer: handle", "error", err.Error(), "url", req.URL, "product_id", req.ProductID, "event_id", req.EventID)
		}
	}
}
