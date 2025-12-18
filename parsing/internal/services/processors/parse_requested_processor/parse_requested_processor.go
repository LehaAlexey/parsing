package parse_requested_processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/LehaAlexey/Parsing/internal/models"
	"github.com/LehaAlexey/Parsing/internal/models/events"
	"github.com/LehaAlexey/Parsing/internal/parser"
	kafkago "github.com/segmentio/kafka-go"
)

type Processor struct {
	extractor *parser.Extractor
	fetcher   *parser.Fetcher
	writer    *kafkago.Writer
}

func New(extractor *parser.Extractor, fetcher *parser.Fetcher, writer *kafkago.Writer) *Processor {
	return &Processor{extractor: extractor, fetcher: fetcher, writer: writer}
}

func (p *Processor) Handle(ctx context.Context, req *events.ParseRequested) error {
	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		return fmt.Errorf("empty url")
	}

	if req.EventID == "" {
		req.EventID = models.NewEventID()
	}
	if req.CorrelationID == "" {
		req.CorrelationID = req.EventID
	}

	body, finalURL, err := p.fetcher.Fetch(ctx, req.URL)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	price, currency, ok := p.extractor.Extract(body)
	if !ok {
		return fmt.Errorf("price not found")
	}
	if currency == "" {
		currency = "RUB"
	}

	parsedAt := time.Now().UTC()
	pm := events.PriceMeasured{
		EventID:       models.Sha256Hex("PriceMeasured|" + req.EventID),
		OccurredAt:    parsedAt,
		CorrelationID: req.CorrelationID,
		ProductID:     req.ProductID,
		Price:         price,
		Currency:      currency,
		ParsedAt:      parsedAt,
		SourceURL:     firstNonEmpty(finalURL, req.URL),
		MetaHash:      models.Sha256Hex(firstNonEmpty(finalURL, req.URL) + "|" + strconv.FormatInt(price, 10) + "|" + currency),
	}

	payload, err := json.Marshal(&pm)
	if err != nil {
		return fmt.Errorf("marshal price_measured: %w", err)
	}

	key := req.ProductID
	if key == "" {
		key = models.Sha256Hex(pm.SourceURL)
	}

	if err := p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(key),
		Value: payload,
	}); err != nil {
		return fmt.Errorf("kafka write: %w", err)
	}

	slog.Info("price measured published",
		"product_id", req.ProductID,
		"price", price,
		"currency", currency,
		"url", pm.SourceURL,
		"correlation_id", pm.CorrelationID,
	)

	return nil
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}
