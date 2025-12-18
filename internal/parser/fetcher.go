package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type FetcherConfig struct {
	UserAgent            string
	RequestTimeout       time.Duration
	MaxBodyBytes         int64
	Retries              int
	MinBackoff           time.Duration
	MaxBackoff           time.Duration
	PerDomainMinInterval time.Duration
}

type Fetcher struct {
	cfg       FetcherConfig
	client    *http.Client
	limiter   *domainLimiter
}

func NewFetcher(cfg FetcherConfig) *Fetcher {
	if cfg.UserAgent == "" {
		cfg.UserAgent = "price-tracker-parsing/1.0"
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 8 * time.Second
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 5 * 1024 * 1024
	}
	if cfg.Retries <= 0 {
		cfg.Retries = 3
	}
	if cfg.MinBackoff <= 0 {
		cfg.MinBackoff = 200 * time.Millisecond
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 2 * time.Second
	}
	if cfg.PerDomainMinInterval <= 0 {
		cfg.PerDomainMinInterval = 300 * time.Millisecond
	}

	return &Fetcher{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.RequestTimeout,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		limiter: newDomainLimiter(cfg.PerDomainMinInterval),
	}
}

func (f *Fetcher) Fetch(ctx context.Context, rawURL string) ([]byte, string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, "", fmt.Errorf("empty url")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, "", fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	u.Fragment = ""

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return nil, "", fmt.Errorf("invalid url host")
	}

	var lastErr error
	for attempt := 0; attempt <= f.cfg.Retries; attempt++ {
		if err := f.limiter.Wait(ctx, host); err != nil {
			return nil, "", err
		}

		body, finalURL, err := f.fetchOnce(ctx, u.String())
		if err == nil {
			return body, finalURL, nil
		}

		lastErr = err
		if attempt == f.cfg.Retries {
			break
		}

		sleep := backoff(attempt, f.cfg.MinBackoff, f.cfg.MaxBackoff)
		timer := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, "", ctx.Err()
		case <-timer.C:
		}
	}

	return nil, "", lastErr
}

func (f *Fetcher) fetchOnce(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", f.cfg.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en;q=0.5")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("http status %d", resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, f.cfg.MaxBodyBytes)
	b, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", err
	}

	finalURL := ""
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	return bytes.Clone(b), finalURL, nil
}

func backoff(attempt int, min, max time.Duration) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	sleep := min * time.Duration(1<<attempt)
	if sleep > max {
		sleep = max
	}
	jitter := time.Duration(rand.IntN(120)) * time.Millisecond
	return sleep + jitter
}

type domainLimiter struct {
	minInterval time.Duration
	mu          sync.Mutex
	last        map[string]time.Time
}

func newDomainLimiter(minInterval time.Duration) *domainLimiter {
	return &domainLimiter{
		minInterval: minInterval,
		last:        make(map[string]time.Time),
	}
}

func (l *domainLimiter) Wait(ctx context.Context, host string) error {
	l.mu.Lock()
	last, ok := l.last[host]
	now := time.Now()
	var wait time.Duration
	if ok {
		elapsed := now.Sub(last)
		if elapsed < l.minInterval {
			wait = l.minInterval - elapsed
		}
	}
	l.last[host] = now.Add(wait)
	l.mu.Unlock()

	if wait <= 0 {
		return nil
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

