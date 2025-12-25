package parse_requested_processor_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/LehaAlexey/Parsing/internal/models"
	"github.com/LehaAlexey/Parsing/internal/models/events"
	parse_requested_processor "github.com/LehaAlexey/Parsing/internal/services/processors/parse_requested_processor"
	processorMocks "github.com/LehaAlexey/Parsing/internal/services/processors/parse_requested_processor/mocks"
	kafkaMocks "github.com/LehaAlexey/Parsing/internal/kafka/mocks"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandle_Success(t *testing.T) {
	t.Parallel()

	extractor := processorMocks.NewMockExtractor(t)
	fetcher := processorMocks.NewMockFetcher(t)
	writer := kafkaMocks.NewMockWriter(t)

	fetcher.EXPECT().
		Fetch(mock.Anything, "https://example.com").
		Return([]byte("<html></html>"), "https://final.example.com", nil)
	extractor.EXPECT().
		Extract([]byte("<html></html>")).
		Return(int64(12345), "USD", true)
	writer.EXPECT().
		WriteMessages(mock.Anything, mock.Anything).
		Run(func(_ context.Context, msgs ...kafka.Message) {
			require.Len(t, msgs, 1)
			require.Equal(t, []byte("product-1"), msgs[0].Key)

			var pm events.PriceMeasured
			require.NoError(t, json.Unmarshal(msgs[0].Value, &pm))
			require.Equal(t, models.Sha256Hex("PriceMeasured|evt-1"), pm.EventID)
			require.Equal(t, "corr-1", pm.CorrelationID)
			require.Equal(t, "product-1", pm.ProductID)
			require.Equal(t, int64(12345), pm.Price)
			require.Equal(t, "USD", pm.Currency)
			require.Equal(t, "https://final.example.com", pm.SourceURL)
			require.Equal(t, models.Sha256Hex("https://final.example.com|12345|USD"), pm.MetaHash)
			require.False(t, pm.OccurredAt.IsZero())
			require.True(t, pm.OccurredAt.Equal(pm.ParsedAt))
		}).
		Return(nil)

	processor := parse_requested_processor.New(extractor, fetcher, writer)

	req := &events.ParseRequested{
		EventID:       "evt-1",
		CorrelationID: "corr-1",
		ProductID:     "product-1",
		URL:           " https://example.com ",
	}
	require.NoError(t, processor.Handle(context.Background(), req))
}

func TestHandle_DefaultCurrencyAndKey(t *testing.T) {
	t.Parallel()

	extractor := processorMocks.NewMockExtractor(t)
	fetcher := processorMocks.NewMockFetcher(t)
	writer := kafkaMocks.NewMockWriter(t)

	fetcher.EXPECT().
		Fetch(mock.Anything, "https://example.com/item").
		Return([]byte("<html></html>"), "https://example.com/item", nil)
	extractor.EXPECT().
		Extract([]byte("<html></html>")).
		Return(int64(99), "", true)
	writer.EXPECT().
		WriteMessages(mock.Anything, mock.Anything).
		Run(func(_ context.Context, msgs ...kafka.Message) {
			require.Len(t, msgs, 1)
			require.Equal(t, []byte(models.Sha256Hex("https://example.com/item")), msgs[0].Key)

			var pm events.PriceMeasured
			require.NoError(t, json.Unmarshal(msgs[0].Value, &pm))
			require.Equal(t, "RUB", pm.Currency)
			require.Equal(t, models.Sha256Hex("https://example.com/item|99|RUB"), pm.MetaHash)
		}).
		Return(nil)

	processor := parse_requested_processor.New(extractor, fetcher, writer)

	req := &events.ParseRequested{
		EventID:       "evt-2",
		CorrelationID: "corr-2",
		URL:           "https://example.com/item",
	}
	require.NoError(t, processor.Handle(context.Background(), req))
}

func TestHandle_EmptyURL(t *testing.T) {
	t.Parallel()

	extractor := processorMocks.NewMockExtractor(t)
	fetcher := processorMocks.NewMockFetcher(t)
	writer := kafkaMocks.NewMockWriter(t)

	processor := parse_requested_processor.New(extractor, fetcher, writer)

	err := processor.Handle(context.Background(), &events.ParseRequested{
		URL: "   ",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty url")

	fetcher.AssertNotCalled(t, "Fetch", mock.Anything, mock.Anything)
	extractor.AssertNotCalled(t, "Extract", mock.Anything)
	writer.AssertNotCalled(t, "WriteMessages", mock.Anything, mock.Anything)
}

func TestHandle_FetchError(t *testing.T) {
	t.Parallel()

	extractor := processorMocks.NewMockExtractor(t)
	fetcher := processorMocks.NewMockFetcher(t)
	writer := kafkaMocks.NewMockWriter(t)

	fetcher.EXPECT().
		Fetch(mock.Anything, "https://example.com").
		Return(nil, "", assertError("boom"))

	processor := parse_requested_processor.New(extractor, fetcher, writer)

	err := processor.Handle(context.Background(), &events.ParseRequested{
		URL: "https://example.com",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "fetch:")

	extractor.AssertNotCalled(t, "Extract", mock.Anything)
	writer.AssertNotCalled(t, "WriteMessages", mock.Anything, mock.Anything)
}

func TestHandle_PriceNotFound(t *testing.T) {
	t.Parallel()

	extractor := processorMocks.NewMockExtractor(t)
	fetcher := processorMocks.NewMockFetcher(t)
	writer := kafkaMocks.NewMockWriter(t)

	fetcher.EXPECT().
		Fetch(mock.Anything, "https://example.com").
		Return([]byte("<html></html>"), "https://example.com", nil)
	extractor.EXPECT().
		Extract([]byte("<html></html>")).
		Return(int64(0), "", false)

	processor := parse_requested_processor.New(extractor, fetcher, writer)

	err := processor.Handle(context.Background(), &events.ParseRequested{
		URL: "https://example.com",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "price not found")

	writer.AssertNotCalled(t, "WriteMessages", mock.Anything, mock.Anything)
}

type assertError string

func (e assertError) Error() string {
	return string(e)
}
