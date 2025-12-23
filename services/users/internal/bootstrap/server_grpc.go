package bootstrap

import (
	"context"
	"net"

	"github.com/LehaAlexey/Users/internal/api/grpcserver"
	"github.com/LehaAlexey/Users/internal/pb/users"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	addr    string
	server  *grpc.Server
	handler *grpcserver.Server
}

func NewGRPCServer(addr string, server *grpc.Server, handler *grpcserver.Server) *GRPCServer {
	if addr == "" {
		addr = ":50061"
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
	users.RegisterUsersServiceServer(s.server, s.handler)

	go func() {
		<-ctx.Done()
		s.server.GracefulStop()
	}()

	return s.server.Serve(lis)
}

