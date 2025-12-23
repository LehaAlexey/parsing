package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/LehaAlexey/Users/config"
	"github.com/LehaAlexey/Users/internal/api/grpcserver"
	"github.com/LehaAlexey/Users/internal/api/httpapi"
	"github.com/LehaAlexey/Users/internal/kafka"
	"github.com/LehaAlexey/Users/internal/scheduler"
	"github.com/LehaAlexey/Users/internal/services/userservice"
	"github.com/LehaAlexey/Users/internal/storage/pgstorage"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type App struct {
	server    HTTPServerRunner
	scheduler SchedulerRunner
	grpcServer GRPCServerRunner
}

func InitApp(configuration *config.Config) (*App, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		configuration.Database.Username,
		configuration.Database.Password,
		configuration.Database.Host,
		configuration.Database.Port,
		configuration.Database.DBName,
		configuration.Database.SSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgx pool: %w", err)
	}

	storage := pgstorage.New(pool)
	service := userservice.New(storage, configuration.Scheduler.DefaultIntervalSeconds)
	handler := httpapi.New(service)

	server := NewHTTPServer(configuration.HTTP.Addr, handler.Routes())

	grpcSrv := grpc.NewServer()
	grpcHandler := grpcserver.New(service)
	grpcServer := NewGRPCServer(configuration.GRPC.Addr, grpcSrv, grpcHandler)

	kafkaBrokers := []string{fmt.Sprintf("%s:%d", configuration.Kafka.Host, configuration.Kafka.Port)}
	writer := kafka.NewWriter(kafkaBrokers, configuration.Kafka.ParseRequestedTopic)
	sched := scheduler.New(storage, writer, time.Duration(configuration.Scheduler.TickSeconds)*time.Second, configuration.Scheduler.DefaultIntervalSeconds, configuration.Scheduler.MaxBatch)

	return &App{server: server, scheduler: sched, grpcServer: grpcServer}, nil
}

type HTTPServerRunner interface {
	Run(ctx context.Context) error
}

type SchedulerRunner interface {
	Run(ctx context.Context) error
}

type GRPCServerRunner interface {
	Run(ctx context.Context) error
}
