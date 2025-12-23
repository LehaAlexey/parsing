package bootstrap

import (
	"context"
	"errors"
)

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := a.server.Run(ctx); err != nil {
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

