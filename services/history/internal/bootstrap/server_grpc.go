package bootstrap

import (
	"context"
	"net"

	"github.com/LehaAlexey/History/internal/api/grpcserver"
	"github.com/LehaAlexey/History/internal/pb/history"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	addr    string
	server  *grpc.Server
	handler *grpcserver.Server
}

func NewGRPCServer(addr string, server *grpc.Server, handler *grpcserver.Server) *GRPCServer {
	if addr == "" {
		addr = ":50062"
	}
	return &GRPCServer{addr: addr, server: server, handler: handler}
}

func (s *GRPCServer) Addr() string {
	return s.addr
}

func (s *GRPCServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	history.RegisterHistoryServiceServer(s.server, s.handler)

	go func() {
		<-ctx.Done()
		s.server.GracefulStop()
	}()

	return s.server.Serve(lis)
}

