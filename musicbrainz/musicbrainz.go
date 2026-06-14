// Package musicbrainz is the library behind the musicbrainz command line:
// the HTTP client, request shaping, and typed data models for the MusicBrainz
// open music metadata database (https://musicbrainz.org/ws/2).
//
// MusicBrainz is rate-limited to 1 request per second. The Client paces
// requests at 1100 ms intervals and retries 503/429/5xx with exponential
// backoff. A descriptive User-Agent header is required by the API Terms of Service.
package musicbrainz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
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
// a descriptive User-Agent, and a 1100 ms rate floor (10% above the 1 req/s limit).
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://musicbrainz.org/ws/2",
		UserAgent: "musicbrainz-cli/0.1.0 (github.com/tamnd/musicbrainz-cli)",
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
// Returns at most limit results (pass 0 for default of 10; max 100).
func (c *Client) Artists(ctx context.Context, query string, limit int) ([]Artist, error) {
	n := clampLimit(limit)
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
			Rank:      i + 1,
			MBID:      a.ID,
			Name:      a.Name,
			Type:      a.Type,
			Country:   a.Country,
			Area:      a.Area.Name,
			BeginDate: a.LifeSpan.Begin,
			EndDate:   a.LifeSpan.End,
			Score:     a.Score,
			URL:       "https://musicbrainz.org/artist/" + a.ID,
		})
	}
	return truncate(out, limit), nil
}

// Recordings searches MusicBrainz for recordings matching query.
// Returns at most limit results (pass 0 for default of 10; max 100).
func (c *Client) Recordings(ctx context.Context, query string, limit int) ([]Recording, error) {
	n := clampLimit(limit)
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
		artist := joinCredits(r.ArtistCredit)
		relTitle, relDate := "", ""
		if len(r.Releases) > 0 {
			relTitle = r.Releases[0].Title
			relDate = r.Releases[0].Date
		}
		out = append(out, Recording{
			Rank:         i + 1,
			MBID:         r.ID,
			Title:        r.Title,
			ArtistCredit: artist,
			ReleaseTitle: relTitle,
			ReleaseDate:  relDate,
			LengthMs:     r.Length,
			Score:        r.Score,
			URL:          "https://musicbrainz.org/recording/" + r.ID,
		})
	}
	return truncate(out, limit), nil
}

// Releases searches MusicBrainz for releases matching query.
// Returns at most limit results (pass 0 for default of 10; max 100).
func (c *Client) Releases(ctx context.Context, query string, limit int) ([]Release, error) {
	n := clampLimit(limit)
	u := fmt.Sprintf("%s/release?query=%s&fmt=json&limit=%d",
		c.cfg.BaseURL, neturl.QueryEscape(query), n)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp releaseSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode releases: %w", err)
	}
	out := make([]Release, 0, len(resp.Releases))
	for i, r := range resp.Releases {
		label := ""
		if len(r.LabelInfo) > 0 {
			label = r.LabelInfo[0].Label.Name
		}
		out = append(out, Release{
			Rank:         i + 1,
			MBID:         r.ID,
			Title:        r.Title,
			Status:       r.Status,
			Date:         r.Date,
			Country:      r.Country,
			Label:        label,
			Type:         r.ReleaseGroup.PrimaryType,
			ArtistCredit: joinCredits(r.ArtistCredit),
			Score:        r.Score,
			URL:          "https://musicbrainz.org/release/" + r.ID,
		})
	}
	return truncate(out, limit), nil
}

// ReleaseGroups searches MusicBrainz for release groups matching query.
// Returns at most limit results (pass 0 for default of 10; max 100).
func (c *Client) ReleaseGroups(ctx context.Context, query string, limit int) ([]ReleaseGroup, error) {
	n := clampLimit(limit)
	u := fmt.Sprintf("%s/release-group?query=%s&fmt=json&limit=%d",
		c.cfg.BaseURL, neturl.QueryEscape(query), n)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp releaseGroupSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode release-groups: %w", err)
	}
	out := make([]ReleaseGroup, 0, len(resp.ReleaseGroups))
	for i, rg := range resp.ReleaseGroups {
		out = append(out, ReleaseGroup{
			Rank:             i + 1,
			MBID:             rg.ID,
			Title:            rg.Title,
			Type:             rg.PrimaryType,
			FirstReleaseDate: rg.FirstReleaseDate,
			ArtistCredit:     joinCredits(rg.ArtistCredit),
			Score:            rg.Score,
			URL:              "https://musicbrainz.org/release-group/" + rg.ID,
		})
	}
	return truncate(out, limit), nil
}

// Labels searches MusicBrainz for labels matching query.
// Returns at most limit results (pass 0 for default of 10; max 100).
func (c *Client) Labels(ctx context.Context, query string, limit int) ([]Label, error) {
	n := clampLimit(limit)
	u := fmt.Sprintf("%s/label?query=%s&fmt=json&limit=%d",
		c.cfg.BaseURL, neturl.QueryEscape(query), n)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp labelSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode labels: %w", err)
	}
	out := make([]Label, 0, len(resp.Labels))
	for i, l := range resp.Labels {
		out = append(out, Label{
			Rank:    i + 1,
			MBID:    l.ID,
			Name:    l.Name,
			Type:    l.Type,
			Country: l.Country,
			Score:   l.Score,
			URL:     "https://musicbrainz.org/label/" + l.ID,
		})
	}
	return truncate(out, limit), nil
}

// GetArtist fetches a full artist record by MBID, including their releases.
func (c *Client) GetArtist(ctx context.Context, mbid string) (*ArtistDetail, error) {
	u := fmt.Sprintf("%s/artist/%s?inc=releases&fmt=json",
		c.cfg.BaseURL, neturl.PathEscape(mbid))
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var w wireArtistDetail
	if err := json.Unmarshal(body, &w); err != nil {
		return nil, fmt.Errorf("decode artist %s: %w", mbid, err)
	}
	releases := make([]ReleaseStub, 0, len(w.Releases))
	for _, r := range w.Releases {
		releases = append(releases, ReleaseStub{
			MBID:   r.ID,
			Title:  r.Title,
			Date:   r.Date,
			Status: r.Status,
		})
	}
	return &ArtistDetail{
		MBID:      w.ID,
		Name:      w.Name,
		Type:      w.Type,
		Country:   w.Country,
		Area:      w.Area.Name,
		BeginDate: w.LifeSpan.Begin,
		EndDate:   w.LifeSpan.End,
		URL:       "https://musicbrainz.org/artist/" + w.ID,
		Releases:  releases,
	}, nil
}

// --- HTTP internals ---

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

// --- helpers ---

func clampLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func truncate[T any](s []T, limit int) []T {
	if limit > 0 && len(s) > limit {
		return s[:limit]
	}
	return s
}

// joinCredits joins artist-credit names with " / ".
func joinCredits(credits []creditItem) string {
	names := make([]string, 0, len(credits))
	for _, c := range credits {
		if c.Name != "" {
			names = append(names, c.Name)
		} else if c.Artist.Name != "" {
			names = append(names, c.Artist.Name)
		}
	}
	return strings.Join(names, " / ")
}
