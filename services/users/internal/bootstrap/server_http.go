package bootstrap

import (
	"context"
	"net"
	"net/http"
	"time"
)

type HTTPServer struct {
	addr    string
	handler http.Handler
}

func NewHTTPServer(addr string, handler http.Handler) *HTTPServer {
	if addr == "" {
		addr = ":8071"
	}
	return &HTTPServer{addr: addr, handler: handler}
}

func (s *HTTPServer) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.handler,
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

