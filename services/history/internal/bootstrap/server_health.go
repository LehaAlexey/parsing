package bootstrap

import (
	"context"
	"net"
	"net/http"
	"time"
)

type HealthServer struct {
	addr string
}

func NewHealthServer(addr string) *HealthServer {
	if addr == "" {
		addr = ":8072"
	}
	return &HealthServer{addr: addr}
}

func (s *HealthServer) Addr() string { return s.addr }

func (s *HealthServer) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	return srv.Serve(lis)
}

