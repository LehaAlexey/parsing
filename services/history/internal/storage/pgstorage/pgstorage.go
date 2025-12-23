package pgstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/LehaAlexey/History/internal/models/events"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) InsertPrice(ctx context.Context, m *events.PriceMeasured) error {
	const q = `
		INSERT INTO price_history (event_id, product_id, price, currency, parsed_at, source_url, meta_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (event_id) DO NOTHING;
	`
	_, err := s.pool.Exec(ctx, q, m.EventID, m.ProductID, m.Price, m.Currency, m.ParsedAt, m.SourceURL, m.MetaHash)
	if err != nil {
		return fmt.Errorf("insert price: %w", err)
	}
	return nil
}

func (s *Storage) GetHistory(ctx context.Context, productID string, from time.Time, to time.Time, limit int) ([]events.PriceMeasured, error) {
	const q = `
		SELECT event_id, product_id, price, currency, parsed_at, source_url, meta_hash
		FROM price_history
		WHERE product_id = $1 AND parsed_at >= $2 AND parsed_at <= $3
		ORDER BY parsed_at DESC
		LIMIT $4;
	`
	rows, err := s.pool.Query(ctx, q, productID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	defer rows.Close()

	result := make([]events.PriceMeasured, 0, 64)
	for rows.Next() {
		var m events.PriceMeasured
		if err := rows.Scan(&m.EventID, &m.ProductID, &m.Price, &m.Currency, &m.ParsedAt, &m.SourceURL, &m.MetaHash); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		result = append(result, m)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}

	return result, nil
}

