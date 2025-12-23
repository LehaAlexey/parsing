package grpcserver

import (
	"context"
	"time"

	"github.com/LehaAlexey/History/internal/models/events"
	"github.com/LehaAlexey/History/internal/pb/history"
)

type Service interface {
	GetHistory(ctx context.Context, productID string, from time.Time, to time.Time, limit int) ([]events.PriceMeasured, error)
}

type Server struct {
	history.UnimplementedHistoryServiceServer
	service Service
}

func New(service Service) *Server {
	return &Server{service: service}
}

func (s *Server) GetHistory(ctx context.Context, req *history.GetHistoryRequest) (*history.GetHistoryResponse, error) {
	from := time.Unix(req.FromUnix, 0).UTC()
	to := time.Unix(req.ToUnix, 0).UTC()
	points, err := s.service.GetHistory(ctx, req.ProductId, from, to, int(req.Limit))
	if err != nil {
		return nil, err
	}

	resp := &history.GetHistoryResponse{Points: make([]*history.HistoryPoint, 0, len(points))}
	for _, p := range points {
		resp.Points = append(resp.Points, &history.HistoryPoint{
			ProductId: p.ProductID,
			Price:     p.Price,
			Currency:  p.Currency,
			ParsedAt:  p.ParsedAt.Unix(),
			SourceUrl: p.SourceURL,
			EventId:   p.EventID,
		})
	}

	return resp, nil
}

