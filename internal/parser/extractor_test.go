package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ExtractorSuite struct {
	suite.Suite
	extractor *Extractor
}

func (s *ExtractorSuite) SetupTest() {
	s.extractor = NewExtractor()
}

func (s *ExtractorSuite) TestExtract_Empty() {
	price, currency, ok := s.extractor.Extract(nil)
	s.False(ok)
	s.Equal(int64(0), price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestExtract_MetaItemprop() {
	html := `<html><head>
		<meta itemprop="priceCurrency" content="RUB">
		<meta itemprop="price" content="12 345">
	</head></html>`

	price, currency, ok := s.extractor.Extract([]byte(html))
	s.True(ok)
	s.Equal(int64(12345), price)
	s.Equal("RUB", currency)
}

func (s *ExtractorSuite) TestExtract_MetaProperty() {
	html := `<html><head>
		<meta property="product:price:currency" content="RUR">
		<meta property="product:price:amount" content="999">
	</head></html>`

	price, currency, ok := s.extractor.Extract([]byte(html))
	s.True(ok)
	s.Equal(int64(999), price)
	s.Equal("RUB", currency)
}

func (s *ExtractorSuite) TestExtract_JSONLD_Offers() {
	html := `<html><head>
		<script type="application/ld+json">
		{"offers":{"price":"19990","priceCurrency":"USD"}}
		</script>
	</head></html>`

	price, currency, ok := s.extractor.Extract([]byte(html))
	s.True(ok)
	s.Equal(int64(19990), price)
	s.Equal("USD", currency)
}

func (s *ExtractorSuite) TestExtract_ScriptJSON() {
	html := `<html><head>
		<script>var product = {"price":"321","currency":"EUR"};</script>
	</head></html>`

	price, currency, ok := s.extractor.Extract([]byte(html))
	s.True(ok)
	s.Equal(int64(321), price)
	s.Equal("EUR", currency)
}

func (s *ExtractorSuite) TestExtract_TextWithCurrency() {
	html := `usd 10000`
	price, currency, ok := s.extractor.Extract([]byte(html))
	s.True(ok)
	s.Equal(int64(10000), price)
	s.Equal("USD", currency)
}

func (s *ExtractorSuite) TestExtract_RegexFallback() {
	html := `<html><body>price: 54321</body></html>`
	price, currency, ok := s.extractor.Extract([]byte(html))
	s.True(ok)
	s.Equal(int64(54321), price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestExtract_NoMatches() {
	html := `<html><body>nothing here</body></html>`
	price, currency, ok := s.extractor.Extract([]byte(html))
	s.False(ok)
	s.Equal(int64(0), price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestExtractFromMeta_CurrencyOnly() {
	html := `<html><head><meta itemprop="priceCurrency" content="USD"></head></html>`
	price, currency, ok := extractFromMeta([]byte(html))
	s.True(ok)
	s.Equal("", price)
	s.Equal("USD", currency)
}

func (s *ExtractorSuite) TestExtractFromJSONLD_InvalidJSON() {
	html := `<html><head><script type="application/ld+json">{bad json}</script></head></html>`
	price, currency, ok := extractFromJSONLD([]byte(html))
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestExtractFromJSONLD_NotJSONLD() {
	html := `<html><head><script type="text/plain">{"price":"1"}</script></head></html>`
	price, currency, ok := extractFromJSONLD([]byte(html))
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestExtractFromJSONLD_EmptyScript() {
	html := `<html><head><script type="application/ld+json"></script></head></html>`
	price, currency, ok := extractFromJSONLD([]byte(html))
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestExtractFromJSONLD_WhitespaceScript() {
	html := `<html><head><script type="application/ld+json">   </script></head></html>`
	price, currency, ok := extractFromJSONLD([]byte(html))
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestExtractFromScriptJSON_NoJSON() {
	html := `<html><head><script>var test = no_json_here;</script></head></html>`
	price, currency, ok := extractFromScriptJSON([]byte(html))
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestExtractFromScriptJSON_EmptyScript() {
	html := `<html><head><script></script></head></html>`
	price, currency, ok := extractFromScriptJSON([]byte(html))
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestExtractFromScriptJSON_WhitespaceScript() {
	html := `<html><head><script>   </script></head></html>`
	price, currency, ok := extractFromScriptJSON([]byte(html))
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestParseEmbeddedJSON() {
	price, currency, ok := parseEmbeddedJSON(`{"priceValue":"555","currencyCode":"USD"}`)
	s.True(ok)
	s.Equal("555", price)
	s.Equal("USD", currency)

	price, currency, ok = parseEmbeddedJSON(`prefix {"price":"777","currency":"EUR"} suffix`)
	s.True(ok)
	s.Equal("777", price)
	s.Equal("EUR", currency)

	price, currency, ok = parseEmbeddedJSON(`no braces at all`)
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)

	price, currency, ok = parseEmbeddedJSON(`{`)
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)

	price, currency, ok = parseEmbeddedJSON(`{"no_price":123}`)
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)

	price, currency, ok = parseEmbeddedJSON(`bad {json}`)
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)
}

func (s *ExtractorSuite) TestFindPriceCurrency() {
	price, currency, ok := findPriceCurrency(map[string]any{
		"priceValue":   "888",
		"currencyCode": "USD",
	})
	s.True(ok)
	s.Equal("888", price)
	s.Equal("USD", currency)

	price, currency, ok = findPriceCurrency(map[string]any{
		"offers": map[string]any{
			"price":         123.4,
			"priceCurrency": "EUR",
		},
	})
	s.True(ok)
	s.Equal("123.4", price)
	s.Equal("EUR", currency)

	price, currency, ok = findPriceCurrency([]any{
		map[string]any{"value": "9", "currency": "USD"},
	})
	s.True(ok)
	s.Equal("9", price)
	s.Equal("USD", currency)

	price, currency, ok = findPriceCurrency(map[string]any{
		"price": struct{}{},
	})
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)

	price, currency, ok = findPriceCurrency(map[string]any{})
	s.False(ok)
	s.Equal("", price)
	s.Equal("", currency)

	price, currency, ok = findPriceCurrency(map[string]any{
		"offers": map[string]any{"no_price": "x"},
		"data":   map[string]any{"price": "42", "currency": "USD"},
	})
	s.True(ok)
	s.Equal("42", price)
	s.Equal("USD", currency)
}

func (s *ExtractorSuite) TestFirstKeyToStringNormalizeAndParse() {
	_, ok := firstKey(map[string]any{"a": 1}, "b", "c")
	s.False(ok)

	s.Equal("abc", toString(" abc "))
	s.Equal("10.5", toString(float64(10.5)))
	s.Equal("7", toString(int(7)))
	s.Equal("8", toString(int64(8)))
	s.Equal("", toString(struct{}{}))

	s.Equal("", normalizeCurrency(""))
	s.Equal("RUB", normalizeCurrency("RUR"))
	s.Equal("RUB", normalizeCurrency("rub"))
	s.Equal("USD", normalizeCurrency("usd"))

	price, ok := parsePriceInt64(" 1 234,56 ")
	s.True(ok)
	s.Equal(int64(1235), price)

	price, ok = parsePriceInt64("1.234.56")
	s.True(ok)
	s.Equal(int64(1235), price)

	price, ok = parsePriceInt64("RUB 2 000")
	s.True(ok)
	s.Equal(int64(2000), price)

	price, ok = parsePriceInt64("0")
	s.False(ok)
	s.Equal(int64(0), price)

	price, ok = parsePriceInt64("abc")
	s.False(ok)
	s.Equal(int64(0), price)

	price, ok = parsePriceInt64(".")
	s.False(ok)
	s.Equal(int64(0), price)

	price, ok = parsePriceInt64("100")
	s.True(ok)
	s.Equal(int64(100), price)

	price, ok = parsePriceInt64("")
	s.False(ok)
	s.Equal(int64(0), price)
}

func TestExtractorSuite(t *testing.T) {
	suite.Run(t, new(ExtractorSuite))
}
