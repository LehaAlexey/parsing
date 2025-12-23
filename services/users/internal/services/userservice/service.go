package userservice

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/LehaAlexey/Users/internal/models"
)

type Storage interface {
	CreateUser(ctx context.Context, email string, name string) (*models.User, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	AddURL(ctx context.Context, userID string, url string, normalizedURL string, intervalSeconds int) (*models.UserURL, error)
	ListUserURLs(ctx context.Context, userID string, limit int) ([]models.UserURL, error)
}

type Service struct {
	storage Storage
	defaultIntervalSeconds int
}

func New(storage Storage, defaultIntervalSeconds int) *Service {
	if defaultIntervalSeconds <= 0 {
		defaultIntervalSeconds = 3600
	}
	return &Service{storage: storage, defaultIntervalSeconds: defaultIntervalSeconds}
}

type CreateUserRequest struct {
	Email string
	Name  string
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*models.User, error) {
	email := strings.TrimSpace(req.Email)
	name := strings.TrimSpace(req.Name)
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	return s.storage.CreateUser(ctx, email, name)
}

func (s *Service) GetUser(ctx context.Context, userID string) (*models.User, error) {
	id := strings.TrimSpace(userID)
	if id == "" {
		return nil, fmt.Errorf("user id is required")
	}

	return s.storage.GetUserByID(ctx, id)
}

func (s *Service) AddURL(ctx context.Context, userID string, rawURL string, intervalSeconds int) (*models.UserURL, error) {
	id := strings.TrimSpace(userID)
	if id == "" {
		return nil, fmt.Errorf("user id is required")
	}
	u, err := NormalizeURL(rawURL)
	if err != nil {
		return nil, err
	}

	if intervalSeconds <= 0 {
		intervalSeconds = s.defaultIntervalSeconds
	}
	return s.storage.AddURL(ctx, id, rawURL, u, intervalSeconds)
}

func (s *Service) ListUserURLs(ctx context.Context, userID string, limit int) ([]models.UserURL, error) {
	id := strings.TrimSpace(userID)
	if id == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	return s.storage.ListUserURLs(ctx, id, limit)
}

func NormalizeURL(rawURL string) (string, error) {
	clean := strings.TrimSpace(rawURL)
	if clean == "" {
		return "", fmt.Errorf("url is required")
	}

	parsed, err := url.Parse(clean)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	parsed.Fragment = ""

	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return "", fmt.Errorf("invalid url host")
	}

	path := strings.TrimRight(parsed.EscapedPath(), "/")
	if path == "" {
		path = "/"
	}

	normalized := fmt.Sprintf("%s://%s%s", parsed.Scheme, host, path)
	if parsed.RawQuery != "" {
		normalized = normalized + "?" + parsed.RawQuery
	}

	return normalized, nil
}
