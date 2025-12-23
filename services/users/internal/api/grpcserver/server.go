package grpcserver

import (
	"context"

	"github.com/LehaAlexey/Users/internal/models"
	"github.com/LehaAlexey/Users/internal/pb/users"
	"github.com/LehaAlexey/Users/internal/services/userservice"
)

type Service interface {
	CreateUser(ctx context.Context, req userservice.CreateUserRequest) (*models.User, error)
	GetUser(ctx context.Context, userID string) (*models.User, error)
	AddURL(ctx context.Context, userID string, rawURL string, intervalSeconds int) (*models.UserURL, error)
	ListUserURLs(ctx context.Context, userID string, limit int) ([]models.UserURL, error)
}

type Server struct {
	users.UnimplementedUsersServiceServer
	service Service
}

func New(service Service) *Server {
	return &Server{service: service}
}

func (s *Server) CreateUser(ctx context.Context, req *users.CreateUserRequest) (*users.CreateUserResponse, error) {
	u, err := s.service.CreateUser(ctx, userservice.CreateUserRequest{Email: req.Email, Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &users.CreateUserResponse{User: mapUser(u)}, nil
}

func (s *Server) GetUser(ctx context.Context, req *users.GetUserRequest) (*users.GetUserResponse, error) {
	u, err := s.service.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &users.GetUserResponse{User: mapUser(u)}, nil
}

func (s *Server) AddUrl(ctx context.Context, req *users.AddUrlRequest) (*users.AddUrlResponse, error) {
	u, err := s.service.AddURL(ctx, req.UserId, req.Url, int(req.PollingIntervalSeconds))
	if err != nil {
		return nil, err
	}
	return &users.AddUrlResponse{Url: mapUserURL(u)}, nil
}

func (s *Server) ListUrls(ctx context.Context, req *users.ListUrlsRequest) (*users.ListUrlsResponse, error) {
	items, err := s.service.ListUserURLs(ctx, req.UserId, int(req.Limit))
	if err != nil {
		return nil, err
	}
	resp := &users.ListUrlsResponse{Urls: make([]*users.UserURL, 0, len(items))}
	for _, item := range items {
		itemCopy := item
		resp.Urls = append(resp.Urls, mapUserURL(&itemCopy))
	}
	return resp, nil
}

func mapUser(u *models.User) *users.User {
	if u == nil {
		return nil
	}
	return &users.User{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Unix(),
	}
}

func mapUserURL(u *models.UserURL) *users.UserURL {
	if u == nil {
		return nil
	}
	return &users.UserURL{
		Id:            u.ID,
		UserId:        u.UserID,
		Url:           u.URL,
		NormalizedUrl: u.NormalizedURL,
		PollingIntervalSeconds: int32(u.PollingIntervalSeconds),
		CreatedAt:     u.CreatedAt.Unix(),
	}
}
