package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/LehaAlexey/Gateway/config"
	"github.com/LehaAlexey/Gateway/internal/bootstrap"
)

func main() {
	configuration, err := config.LoadConfig(os.Getenv("configPath"))
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	app, err := bootstrap.InitApp(configuration)
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		panic(err)
	}
}

