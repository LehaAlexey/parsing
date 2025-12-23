package bootstrap

import (
	"context"
	"errors"
	"log/slog"
)

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 3)

	go func() {
		slog.Info("grpc server starting", "addr", a.server.Addr())
		if err := a.server.Run(ctx); err != nil {
			errCh <- err
		}
	}()

	go func() {
		slog.Info("health server starting", "addr", a.http.Addr())
		if err := a.http.Run(ctx); err != nil {
			errCh <- err
		}
	}()

	go func() {
		slog.Info("kafka consumer starting")
		if err := a.consumer.Consume(ctx); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
}

