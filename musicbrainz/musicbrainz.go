// Package musicbrainz is the library behind the musicbrainz command line:
// the HTTP client, request shaping, and the typed data models for the MusicBrainz
// open music metadata database (artists, recordings).
//
// MusicBrainz is rate-limited to 1 request per second. The Client paces
// requests at 1100 ms intervals and retries 503/429/5xx with exponential
// backoff. A descriptive User-Agent header is required by the API terms.
package musicbrainz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "musicbrainz.org"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with the MusicBrainz public API base URL,
// a descriptive User-Agent, and a 1100 ms rate floor (10% above the 1 req/s
// limit).
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://musicbrainz.org/ws/2",
		UserAgent: "musicbrainz-cli/dev (https://github.com/tamnd/musicbrainz-cli; contact@example.com)",
		Rate:      1100 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
	}
}

// Client talks to MusicBrainz over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Artists searches MusicBrainz for artists matching query.
// It returns at most limit results (pass 0 for the default of 10; max 100).
func (c *Client) Artists(ctx context.Context, query string, limit int) ([]Artist, error) {
	n := limit
	if n <= 0 {
		n = 10
	}
	if n > 100 {
		n = 100
	}
	u := fmt.Sprintf("%s/artist?query=%s&fmt=json&limit=%d",
		c.cfg.BaseURL, neturl.QueryEscape(query), n)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp artistSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode artists: %w", err)
	}
	out := make([]Artist, 0, len(resp.Artists))
	for i, a := range resp.Artists {
		out = append(out, Artist{
			Rank:    i + 1,
			ID:      a.ID,
			Name:    a.Name,
			Type:    a.Type,
			Country: a.Country,
			Score:   a.Score,
			URL:     "https://musicbrainz.org/artist/" + a.ID,
		})
	}
	if limit > 0 && limit < len(out) {
		out = out[:limit]
	}
	return out, nil
}

// Recordings searches MusicBrainz for recordings matching query.
// It returns at most limit results (pass 0 for the default of 10; max 100).
func (c *Client) Recordings(ctx context.Context, query string, limit int) ([]Recording, error) {
	n := limit
	if n <= 0 {
		n = 10
	}
	if n > 100 {
		n = 100
	}
	u := fmt.Sprintf("%s/recording?query=%s&fmt=json&limit=%d",
		c.cfg.BaseURL, neturl.QueryEscape(query), n)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp recordingSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode recordings: %w", err)
	}
	out := make([]Recording, 0, len(resp.Recordings))
	for i, r := range resp.Recordings {
		artist := ""
		if len(r.ArtistCredit) > 0 {
			artist = r.ArtistCredit[0].Name
		}
		out = append(out, Recording{
			Rank:     i + 1,
			ID:       r.ID,
			Title:    r.Title,
			Artist:   artist,
			LengthMs: r.Length,
			Score:    r.Score,
			URL:      "https://musicbrainz.org/recording/" + r.ID,
		})
	}
	if limit > 0 && limit < len(out) {
		out = out[:limit]
	}
	return out, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return b, err != nil, err
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}
