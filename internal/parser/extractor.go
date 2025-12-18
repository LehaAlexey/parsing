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
		priceRe: regexp.MustCompile(`(?i)(?:price|amount)[^0-9]{0,20}([0-9][0-9\\s\\u00A0.,]{0,20})`),
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
				itemprop  string
				property  string
				content   string
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

func findPriceCurrency(v any) (string, string, bool) {
	switch x := v.(type) {
	case map[string]any:
		if p, ok := x["price"]; ok {
			price := toString(p)
			currency := toString(firstKey(x, "priceCurrency", "price_currency", "currency"))
			return price, currency, price != ""
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

func firstKey(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v
		}
	}
	return nil
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
	if strings.Contains(s, "₽") {
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
	s = strings.ReplaceAll(s, "₽", "")
	s = strings.ReplaceAll(s, "руб.", "")
	s = strings.ReplaceAll(s, "руб", "")
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
