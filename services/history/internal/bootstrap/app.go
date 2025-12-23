package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/LehaAlexey/History/config"
	"github.com/LehaAlexey/History/internal/api/grpcserver"
	"github.com/LehaAlexey/History/internal/consumer/price_measured_consumer"
	"github.com/LehaAlexey/History/internal/services/historyservice"
	"github.com/LehaAlexey/History/internal/storage/pgstorage"
	"github.com/LehaAlexey/History/internal/storage/redisstore"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type App struct {
	consumer Consumer
	server   GRPCServerRunner
	http     HealthServerRunner
}

func InitApp(cfg *config.Config) (*App, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgx pool: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		DB:   cfg.Redis.DB,
	})

	store := pgstorage.New(pool)
	cache := redisstore.New(redisClient, cfg.Redis.RecentLimit, time.Duration(cfg.Redis.RecentTTLSeconds)*time.Second)
	service := historyservice.New(store, cache)

	grpcSrv := grpc.NewServer()
	grpcHandler := grpcserver.New(service)
	hServer := NewGRPCServer(cfg.GRPC.Addr, grpcSrv, grpcHandler)
	health := NewHealthServer(cfg.HTTP.Addr)

	brokers := []string{fmt.Sprintf("%s:%d", cfg.Kafka.Host, cfg.Kafka.Port)}
	consumer := price_measured_consumer.New(price_measured_consumer.Config{
		Brokers: brokers,
		GroupID: cfg.Kafka.GroupID,
		Topic:   cfg.Kafka.PriceMeasuredTopic,
	}, service)

	return &App{consumer: consumer, server: hServer, http: health}, nil
}

type Consumer interface {
	Consume(ctx context.Context) error
}

type GRPCServerRunner interface {
	Addr() string
	Run(ctx context.Context) error
}

type HealthServerRunner interface {
	Addr() string
	Run(ctx context.Context) error
}
