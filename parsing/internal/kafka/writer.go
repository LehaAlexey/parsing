package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Writer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

func NewWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.Hash{},
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
	}
}
