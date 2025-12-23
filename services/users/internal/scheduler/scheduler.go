package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/LehaAlexey/Users/internal/kafka"
	"github.com/LehaAlexey/Users/internal/models"
	"github.com/LehaAlexey/Users/internal/models/events"
	kafkago "github.com/segmentio/kafka-go"
)

type Storage interface {
	GetDueURLs(ctx context.Context, limit int) ([]models.UserURL, error)
	MarkScheduled(ctx context.Context, urlID string, intervalSeconds int) error
}

type Scheduler struct {
	storage     Storage
	writer      kafka.Writer
	tick        time.Duration
	intervalSec int
	maxBatch    int
}

func New(storage Storage, writer kafka.Writer, tick time.Duration, intervalSeconds int, maxBatch int) *Scheduler {
	if tick <= 0 {
		tick = 5 * time.Second
	}
	if intervalSeconds <= 0 {
		intervalSeconds = 3600
	}
	if maxBatch <= 0 {
		maxBatch = 100
	}
	return &Scheduler{storage: storage, writer: writer, tick: tick, intervalSec: intervalSeconds, maxBatch: maxBatch}
}

func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context) {
	urls, err := s.storage.GetDueURLs(ctx, s.maxBatch)
	if err != nil {
		slog.Error("scheduler: get due urls", "error", err.Error())
		return
	}
	if len(urls) == 0 {
		return
	}

	for _, item := range urls {
		msg := events.ParseRequested{
			EventID:       newEventID(),
			OccurredAt:    time.Now().UTC(),
			CorrelationID: newEventID(),
			ProductID:     item.ID,
			URL:           item.URL,
			ScheduledAt:   time.Now().UTC(),
			Priority:      0,
		}

		payload, err := json.Marshal(&msg)
		if err != nil {
			slog.Error("scheduler: marshal", "error", err.Error())
			continue
		}

		if err := s.writer.WriteMessages(ctx, kafkago.Message{
			Key:   []byte(item.ID),
			Value: payload,
		}); err != nil {
			slog.Error("scheduler: kafka write", "error", err.Error())
			continue
		}

		interval := item.PollingIntervalSeconds
		if interval <= 0 {
			interval = s.intervalSec
		}
		if err := s.storage.MarkScheduled(ctx, item.ID, interval); err != nil {
			slog.Error("scheduler: mark scheduled", "error", err.Error())
		}
	}
}

func newEventID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
