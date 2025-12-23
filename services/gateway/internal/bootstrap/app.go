package bootstrap

import (
	"context"
	"fmt"

	"github.com/LehaAlexey/Gateway/config"
	"github.com/LehaAlexey/Gateway/internal/api/httpapi"
	"github.com/LehaAlexey/Gateway/internal/clients"
)

type App struct {
	server  HTTPServerRunner
	users   UsersClient
	history HistoryClient
}

func InitApp(cfg *config.Config) (*App, error) {
	usersClient, err := clients.NewUsersClient(cfg.UsersGRPC.Addr)
	if err != nil {
		return nil, fmt.Errorf("users grpc: %w", err)
	}
	historyClient, err := clients.NewHistoryClient(cfg.HistoryGRPC.Addr)
	if err != nil {
		return nil, fmt.Errorf("history grpc: %w", err)
	}

	handler := httpapi.New(usersClient, historyClient)
	server := NewHTTPServer(cfg.HTTP.Addr, handler.Routes())

	return &App{server: server, users: usersClient, history: historyClient}, nil
}

type HTTPServerRunner interface {
	Run(ctx context.Context) error
}

type UsersClient interface {
	Close() error
}

type HistoryClient interface {
	Close() error
}
