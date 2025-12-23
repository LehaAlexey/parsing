package redisstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/LehaAlexey/History/internal/models/events"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	client        *redis.Client
	recentLimit   int
	recentTTL     time.Duration
}

func New(client *redis.Client, recentLimit int, recentTTL time.Duration) *Store {
	return &Store{client: client, recentLimit: recentLimit, recentTTL: recentTTL}
}

func (s *Store) SaveRecent(ctx context.Context, m *events.PriceMeasured) error {
	if m.ProductID == "" {
		return nil
	}

	payload, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal price: %w", err)
	}

	lastKey := fmt.Sprintf("price:last:%s", m.ProductID)
	recentKey := fmt.Sprintf("price:recent:%s", m.ProductID)

	pipe := s.client.Pipeline()
	pipe.Set(ctx, lastKey, payload, s.recentTTL)
	pipe.ZAdd(ctx, recentKey, redis.Z{
		Score:  float64(m.ParsedAt.Unix()),
		Member: payload,
	})
	pipe.ZRemRangeByRank(ctx, recentKey, 0, int64(-s.recentLimit-1))
	pipe.Expire(ctx, recentKey, s.recentTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("save recent: %w", err)
	}
	return nil
}

func (s *Store) GetRecent(ctx context.Context, productID string, from time.Time, to time.Time, limit int) ([]events.PriceMeasured, error) {
	if productID == "" {
		return nil, nil
	}
	key := fmt.Sprintf("price:recent:%s", productID)
	min := fmt.Sprintf("%d", from.Unix())
	max := fmt.Sprintf("%d", to.Unix())

	if limit <= 0 {
		limit = s.recentLimit
	}

	values, err := s.client.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:   min,
		Max:   max,
		Count: int64(limit),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("get recent: %w", err)
	}

	result := make([]events.PriceMeasured, 0, len(values))
	for _, v := range values {
		var m events.PriceMeasured
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			continue
		}
		result = append(result, m)
	}
	return result, nil
}

