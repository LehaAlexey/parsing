package price_measured_consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/LehaAlexey/History/internal/models/events"
	"github.com/segmentio/kafka-go"
)

type Processor interface {
	Handle(ctx context.Context, msg *events.PriceMeasured) error
}

type Config struct {
	Brokers []string
	GroupID string
	Topic   string
}

type Consumer struct {
	cfg       Config
	processor Processor
}

func New(cfg Config, processor Processor) *Consumer {
	return &Consumer{cfg: cfg, processor: processor}
}

func (c *Consumer) Consume(ctx context.Context) error {
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
			slog.Error("price_measured_consumer: read message", "error", err.Error())
			continue
		}

		var m events.PriceMeasured
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			slog.Error("price_measured_consumer: unmarshal", "error", err.Error())
			continue
		}

		if err := c.processor.Handle(ctx, &m); err != nil {
			slog.Error("price_measured_consumer: handle", "error", err.Error(), "product_id", m.ProductID, "event_id", m.EventID)
		}
	}
}

