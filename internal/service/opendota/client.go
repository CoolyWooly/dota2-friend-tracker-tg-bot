package opendota

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/yerlan/dota2/internal/constant"
)

type Client struct {
	httpClient *http.Client
	limiter    *rate.Limiter
	apiKey     string
	baseURL    string
}

func New(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: constant.HTTPClientTimeout},
		limiter:    rate.NewLimiter(rate.Limit(constant.OpenDotaRPS), 1),
		apiKey:     apiKey,
		baseURL:    constant.OpenDotaBaseURL,
	}
}

func (c *Client) GetPlayer(ctx context.Context, accountID int64) (*Player, error) {
	var res Player
	if err := c.get(ctx, fmt.Sprintf("/players/%d", accountID), &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) GetRecentMatches(ctx context.Context, accountID int64) ([]*RecentMatch, error) {
	var res []*RecentMatch
	if err := c.get(ctx, fmt.Sprintf("/players/%d/recentMatches", accountID), &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) GetHeroStats(ctx context.Context) ([]*HeroStat, error) {
	var res []*HeroStat
	if err := c.get(ctx, "/heroStats", &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) get(ctx context.Context, path string, dst any) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}

	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if c.apiKey != "" {
		q := u.Query()
		q.Set("api_key", c.apiKey)
		u.RawQuery = q.Encode()
	}

	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return fmt.Errorf("new request: %w", err)
		}

		urlStr := redactKey(u)
		started := time.Now()
		slog.Info("opendota request", "method", http.MethodGet, "url", urlStr, "attempt", attempt+1)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("do: %w", err)
			slog.Warn("opendota request failed", "url", urlStr, "attempt", attempt+1, "duration", time.Since(started), "error", err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		slog.Info("opendota response", "url", urlStr, "status", resp.StatusCode, "bytes", len(body), "duration", time.Since(started))

		switch {
		case resp.StatusCode == http.StatusOK:
			if err := json.Unmarshal(body, dst); err != nil {
				return fmt.Errorf("decode: %w; body=%s", err, truncate(body, 200))
			}
			return nil
		case resp.StatusCode == http.StatusNotFound:
			return ErrNotFound
		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			lastErr = fmt.Errorf("status %d: %s", resp.StatusCode, truncate(body, 200))
			slog.Warn("opendota retryable", "status", resp.StatusCode, "attempt", attempt+1)
			continue
		default:
			return fmt.Errorf("status %d: %s", resp.StatusCode, truncate(body, 200))
		}
	}
	return lastErr
}

var ErrNotFound = errors.New("opendota: not found")

func redactKey(u *url.URL) string {
	q := u.Query()
	if q.Get("api_key") != "" {
		q.Set("api_key", "***")
	}
	u2 := *u
	u2.RawQuery = q.Encode()
	return u2.String()
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..." + " (truncated, total " + strconv.Itoa(len(b)) + " bytes)"
}
