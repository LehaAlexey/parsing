package bootstrap

import (
	"fmt"
	"time"

	"github.com/LehaAlexey/Parsing/config"
	"github.com/LehaAlexey/Parsing/internal/consumer/parse_requested_consumer"
	"github.com/LehaAlexey/Parsing/internal/kafka"
	"github.com/LehaAlexey/Parsing/internal/parser"
	"github.com/LehaAlexey/Parsing/internal/services/processors/parse_requested_processor"
)

type App struct {
	consumer *parse_requested_consumer.ParseRequestedConsumer
	server   *HealthServer
}

func InitApp(cfg *config.Config) (*App, error) {
	configuration := cfg
	brokers := []string{fmt.Sprintf("%v:%v", cfg.Kafka.Host, cfg.Kafka.Port)}

	writer := kafka.NewWriter(brokers, configuration.Kafka.PriceMeasuredTopic)
	extractor := parser.NewExtractor()
	fetcher := parser.NewFetcher(parser.FetcherConfig{
		UserAgent:              configuration.Parser.UserAgent,
		RequestTimeout:         time.Duration(configuration.Parser.RequestTimeoutMS) * time.Millisecond,
		MaxBodyBytes:           configuration.Parser.MaxBodyBytes,
		Retries:                configuration.Parser.Retries,
		MinBackoff:             time.Duration(configuration.Parser.MinBackoffMS) * time.Millisecond,
		MaxBackoff:             time.Duration(configuration.Parser.MaxBackoffMS) * time.Millisecond,
		PerDomainMinInterval:   time.Duration(configuration.Parser.PerDomainMinIntervalMS) * time.Millisecond,
	})

	processor := parse_requested_processor.New(extractor, fetcher, writer)
	consumer := parse_requested_consumer.New(parse_requested_consumer.Config{
		Brokers: brokers,
		GroupID: configuration.Kafka.GroupID,
		Topic:   configuration.Kafka.ParseRequestedTopic,
	}, processor)

	server := NewHealthServer(configuration.HTTP.Addr)
	return &App{consumer: consumer, server: server}, nil
}
