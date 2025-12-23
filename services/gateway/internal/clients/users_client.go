package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/LehaAlexey/Gateway/internal/pb/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UsersClient struct {
	conn   *grpc.ClientConn
	client users.UsersServiceClient
}

func NewUsersClient(addr string) (*UsersClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("users grpc addr is empty")
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &UsersClient{conn: conn, client: users.NewUsersServiceClient(conn)}, nil
}

func (c *UsersClient) Close() error {
	return c.conn.Close()
}

func (c *UsersClient) CreateUser(ctx context.Context, email string, name string) (*users.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	resp, err := c.client.CreateUser(ctx, &users.CreateUserRequest{Email: email, Name: name})
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

func (c *UsersClient) GetUser(ctx context.Context, id string) (*users.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	resp, err := c.client.GetUser(ctx, &users.GetUserRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

func (c *UsersClient) AddURL(ctx context.Context, userID string, url string, intervalSeconds int32) (*users.UserURL, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	resp, err := c.client.AddUrl(ctx, &users.AddUrlRequest{UserId: userID, Url: url, PollingIntervalSeconds: intervalSeconds})
	if err != nil {
		return nil, err
	}
	return resp.Url, nil
}

func (c *UsersClient) ListURLs(ctx context.Context, userID string, limit int32) ([]*users.UserURL, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	resp, err := c.client.ListUrls(ctx, &users.ListUrlsRequest{UserId: userID, Limit: limit})
	if err != nil {
		return nil, err
	}
	return resp.Urls, nil
}
