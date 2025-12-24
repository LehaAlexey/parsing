package parser

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type Extractor struct {
	priceRe *regexp.Regexp
}

func NewExtractor() *Extractor {
	return &Extractor{
		priceRe: regexp.MustCompile(`(?i)(?:price|amount)[^0-9]{0,20}([0-9][0-9\s.,]{0,20})`),
	}
}

func (e *Extractor) Extract(htmlBytes []byte) (int64, string, bool) {
	if len(htmlBytes) == 0 {
		return 0, "", false
	}

	priceStr, currency, ok := extractFromMeta(htmlBytes)
	if ok {
		if p, ok := parsePriceInt64(priceStr); ok {
			return p, normalizeCurrency(currency), true
		}
	}

	priceStr, currency, ok = extractFromJSONLD(htmlBytes)
	if ok {
		if p, ok := parsePriceInt64(priceStr); ok {
			return p, normalizeCurrency(currency), true
		}
	}

	priceStr, currency, ok = extractFromScriptJSON(htmlBytes)
	if ok {
		if p, ok := parsePriceInt64(priceStr); ok {
			return p, normalizeCurrency(currency), true
		}
	}

	priceStr, currency, ok = extractFromTextWithCurrency(htmlBytes)
	if ok {
		if p, ok := parsePriceInt64(priceStr); ok {
			return p, normalizeCurrency(currency), true
		}
	}

	if m := e.priceRe.FindSubmatch(htmlBytes); len(m) >= 2 {
		if p, ok := parsePriceInt64(string(m[1])); ok {
			return p, "", true
		}
	}

	return 0, "", false
}

func extractFromMeta(b []byte) (string, string, bool) {
	z := html.NewTokenizer(bytes.NewReader(b))
	foundCurrency := ""
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return "", foundCurrency, foundCurrency != ""
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if strings.ToLower(t.Data) != "meta" {
				continue
			}
			var (
				itemprop string
				property string
				content  string
			)
			for _, a := range t.Attr {
				switch strings.ToLower(a.Key) {
				case "itemprop":
					itemprop = strings.ToLower(strings.TrimSpace(a.Val))
				case "property":
					property = strings.ToLower(strings.TrimSpace(a.Val))
				case "content":
					content = strings.TrimSpace(a.Val)
				}
			}
			switch itemprop {
			case "price":
				return content, foundCurrency, content != ""
			case "pricecurrency":
				if content != "" {
					foundCurrency = content
				}
			}
			switch property {
			case "product:price:amount", "og:price:amount":
				return content, foundCurrency, content != ""
			case "product:price:currency", "og:price:currency":
				if content != "" {
					foundCurrency = content
				}
			}
		}
	}
}

func extractFromJSONLD(b []byte) (string, string, bool) {
	z := html.NewTokenizer(bytes.NewReader(b))
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return "", "", false
		case html.StartTagToken:
			t := z.Token()
			if strings.ToLower(t.Data) != "script" {
				continue
			}
			isJSONLD := false
			for _, a := range t.Attr {
				if strings.ToLower(a.Key) == "type" && strings.Contains(strings.ToLower(a.Val), "ld+json") {
					isJSONLD = true
					break
				}
			}
			if !isJSONLD {
				continue
			}

			if z.Next() != html.TextToken {
				continue
			}
			raw := strings.TrimSpace(string(z.Text()))
			if raw == "" {
				continue
			}

			var v any
			if err := json.Unmarshal([]byte(raw), &v); err != nil {
				continue
			}

			price, currency, ok := findPriceCurrency(v)
			if ok && price != "" {
				return price, currency, true
			}
		}
	}
}

func extractFromScriptJSON(b []byte) (string, string, bool) {
	z := html.NewTokenizer(bytes.NewReader(b))
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return "", "", false
		case html.StartTagToken:
			t := z.Token()
			if strings.ToLower(t.Data) != "script" {
				continue
			}
			if z.Next() != html.TextToken {
				continue
			}
			raw := strings.TrimSpace(string(z.Text()))
			if raw == "" {
				continue
			}

			if price, currency, ok := parseEmbeddedJSON(raw); ok {
				return price, currency, true
			}
		}
	}
}

func parseEmbeddedJSON(raw string) (string, string, bool) {
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err == nil {
		if price, currency, ok := findPriceCurrency(v); ok && price != "" {
			return price, currency, true
		}
	}

	start := strings.Index(raw, "{")
	if start == -1 {
		return "", "", false
	}
	end := strings.LastIndex(raw, "}")
	if end <= start {
		return "", "", false
	}
	fragment := strings.TrimSpace(raw[start : end+1])

	if err := json.Unmarshal([]byte(fragment), &v); err != nil {
		return "", "", false
	}
	if price, currency, ok := findPriceCurrency(v); ok && price != "" {
		return price, currency, true
	}

	return "", "", false
}

func extractFromTextWithCurrency(b []byte) (string, string, bool) {
	patterns := []struct {
		re       *regexp.Regexp
		currency string
	}{
		{regexp.MustCompile(`(?i)(?:rub|rur)\s*([0-9][0-9\s.,]{0,20})`), "RUB"},
		{regexp.MustCompile(`(?i)([0-9][0-9\s.,]{0,20})\s*(?:rub|rur)`), "RUB"},
		{regexp.MustCompile(`(?i)(?:usd|\$|dollars?)\s*([0-9][0-9\s.,]{0,20})`), "USD"},
		{regexp.MustCompile(`(?i)([0-9][0-9\s.,]{0,20})\s*(?:usd|\$|dollars?)`), "USD"},
		{regexp.MustCompile(`(?i)(?:eur|euros?)\s*([0-9][0-9\s.,]{0,20})`), "EUR"},
		{regexp.MustCompile(`(?i)([0-9][0-9\s.,]{0,20})\s*(?:eur|euros?)`), "EUR"},
	}
	text := string(b)
	for _, p := range patterns {
		if m := p.re.FindStringSubmatch(text); len(m) >= 2 {
			return m[1], p.currency, true
		}
	}
	return "", "", false
}

func findPriceCurrency(v any) (string, string, bool) {
	switch x := v.(type) {
	case map[string]any:
		if p, ok := firstKey(x, "price", "priceValue", "price_value", "priceNumeric", "price_num", "amount", "value"); ok {
			price := toString(p)
			if price == "" {
				return "", "", false
			}
			cur := ""
			if c, ok := firstKey(x, "priceCurrency", "price_currency", "currency", "currencyCode", "currency_code", "currencyId", "currency_id"); ok {
				cur = toString(c)
			}
			return price, cur, true
		}
		if o, ok := x["offers"]; ok {
			if price, currency, ok := findPriceCurrency(o); ok {
				return price, currency, true
			}
		}
		for _, v2 := range x {
			if price, currency, ok := findPriceCurrency(v2); ok {
				return price, currency, true
			}
		}
	case []any:
		for _, v2 := range x {
			if price, currency, ok := findPriceCurrency(v2); ok {
				return price, currency, true
			}
		}
	}
	return "", "", false
}

func firstKey(m map[string]any, keys ...string) (any, bool) {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v, true
		}
	}
	return nil, false
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return ""
	}
}

func normalizeCurrency(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return ""
	}
	if s == "RUR" {
		return "RUB"
	}
	if strings.Contains(s, "RUB") {
		return "RUB"
	}
	return s
}

func parsePriceInt64(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	s = strings.ReplaceAll(s, "\u00A0", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "RUB", "")
	s = strings.ReplaceAll(s, "RUR", "")
	s = strings.ReplaceAll(s, "USD", "")
	s = strings.ReplaceAll(s, "EUR", "")
	s = strings.ReplaceAll(s, ",", ".")

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' {
			b.WriteRune(r)
		}
	}
	clean := b.String()
	if clean == "" {
		return 0, false
	}
	if strings.Count(clean, ".") > 1 {
		parts := strings.Split(clean, ".")
		clean = strings.Join(parts[:len(parts)-1], "") + "." + parts[len(parts)-1]
	}
	f, err := strconv.ParseFloat(clean, 64)
	if err != nil || f <= 0 {
		return 0, false
	}
	return int64(f + 0.5), true
}
