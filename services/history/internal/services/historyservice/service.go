package historyservice

import (
	"context"
	"fmt"
	"time"

	"github.com/LehaAlexey/History/internal/models/events"
)

type Store interface {
	InsertPrice(ctx context.Context, m *events.PriceMeasured) error
	GetHistory(ctx context.Context, productID string, from time.Time, to time.Time, limit int) ([]events.PriceMeasured, error)
}

type Cache interface {
	SaveRecent(ctx context.Context, m *events.PriceMeasured) error
	GetRecent(ctx context.Context, productID string, from time.Time, to time.Time, limit int) ([]events.PriceMeasured, error)
}

type Service struct {
	store Store
	cache Cache
}

func New(store Store, cache Cache) *Service {
	return &Service{store: store, cache: cache}
}

func (s *Service) Handle(ctx context.Context, m *events.PriceMeasured) error {
	if m == nil {
		return fmt.Errorf("message is nil")
	}
	if m.EventID == "" || m.ProductID == "" {
		return fmt.Errorf("invalid message")
	}

	if err := s.store.InsertPrice(ctx, m); err != nil {
		return err
	}
	if err := s.cache.SaveRecent(ctx, m); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetHistory(ctx context.Context, productID string, from time.Time, to time.Time, limit int) ([]events.PriceMeasured, error) {
	if productID == "" {
		return nil, fmt.Errorf("product id is required")
	}
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}

	points, err := s.cache.GetRecent(ctx, productID, from, to, limit)
	if err == nil && len(points) > 0 {
		return points, nil
	}

	return s.store.GetHistory(ctx, productID, from, to, limit)
}

