package pgstorage

import (
	"context"
	"fmt"

	"github.com/LehaAlexey/Users/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) CreateUser(ctx context.Context, email string, name string) (*models.User, error) {
	const q = `
		INSERT INTO users (email, name)
		VALUES ($1, $2)
		RETURNING id, email, name, created_at;
	`
	row := s.pool.QueryRow(ctx, q, email, name)
	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (s *Storage) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	const q = `
		SELECT id, email, name, created_at
		FROM users
		WHERE id = $1;
	`
	row := s.pool.QueryRow(ctx, q, userID)
	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

func (s *Storage) AddURL(ctx context.Context, userID string, url string, normalizedURL string, intervalSeconds int) (*models.UserURL, error) {
	const q = `
		INSERT INTO user_urls (user_id, url, normalized_url, polling_interval_seconds, next_run_at)
		VALUES ($1, $2, $3, $4, now())
		RETURNING id, user_id, url, normalized_url, polling_interval_seconds, created_at;
	`
	row := s.pool.QueryRow(ctx, q, userID, url, normalizedURL, intervalSeconds)
	var u models.UserURL
	if err := row.Scan(&u.ID, &u.UserID, &u.URL, &u.NormalizedURL, &u.PollingIntervalSeconds, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("add url: %w", err)
	}
	return &u, nil
}

func (s *Storage) GetDueURLs(ctx context.Context, limit int) ([]models.UserURL, error) {
	const q = `
		SELECT id, user_id, url, normalized_url, polling_interval_seconds, created_at
		FROM user_urls
		WHERE next_run_at <= now()
		ORDER BY next_run_at ASC
		LIMIT $1;
	`
	rows, err := s.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("get due urls: %w", err)
	}
	defer rows.Close()

	result := make([]models.UserURL, 0, 64)
	for rows.Next() {
		var u models.UserURL
		if err := rows.Scan(&u.ID, &u.UserID, &u.URL, &u.NormalizedURL, &u.PollingIntervalSeconds, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan due url: %w", err)
		}
		result = append(result, u)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}
	return result, nil
}

func (s *Storage) MarkScheduled(ctx context.Context, urlID string, intervalSeconds int) error {
	const q = `
		UPDATE user_urls
		SET next_run_at = now() + ($2 || ' seconds')::interval
		WHERE id = $1;
	`
	_, err := s.pool.Exec(ctx, q, urlID, intervalSeconds)
	if err != nil {
		return fmt.Errorf("mark scheduled: %w", err)
	}
	return nil
}

func (s *Storage) ListUserURLs(ctx context.Context, userID string, limit int) ([]models.UserURL, error) {
	const q = `
		SELECT id, user_id, url, normalized_url, polling_interval_seconds, created_at
		FROM user_urls
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2;
	`
	rows, err := s.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list urls: %w", err)
	}
	defer rows.Close()

	result := make([]models.UserURL, 0, 16)
	for rows.Next() {
		var u models.UserURL
		if err := rows.Scan(&u.ID, &u.UserID, &u.URL, &u.NormalizedURL, &u.PollingIntervalSeconds, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan url: %w", err)
		}
		result = append(result, u)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}

	return result, nil
}
