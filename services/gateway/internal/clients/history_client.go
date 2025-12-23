package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/LehaAlexey/Gateway/internal/pb/history"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type HistoryClient struct {
	conn   *grpc.ClientConn
	client history.HistoryServiceClient
}

func NewHistoryClient(addr string) (*HistoryClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("history grpc addr is empty")
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &HistoryClient{conn: conn, client: history.NewHistoryServiceClient(conn)}, nil
}

func (c *HistoryClient) Close() error {
	return c.conn.Close()
}

func (c *HistoryClient) GetHistory(ctx context.Context, productID string, fromUnix int64, toUnix int64, limit int32) ([]*history.HistoryPoint, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	resp, err := c.client.GetHistory(ctx, &history.GetHistoryRequest{
		ProductId: productID,
		FromUnix:  fromUnix,
		ToUnix:    toUnix,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	return resp.Points, nil
}

