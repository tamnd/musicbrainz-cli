package musicbrainz_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/musicbrainz-cli/musicbrainz"
)

const fakeArtistsJSON = `{
  "count": 2,
  "artists": [
    {"id": "0383dadf-2a4e-4d10-a46a-e9e041da8eb3", "name": "Queen", "type": "Group", "country": "GB", "score": 100},
    {"id": "f27ec8db-af05-4f36-916e-3d57f91ecf7e", "name": "Michael Jackson", "type": "Person", "country": "US", "score": 85}
  ]
}`

const fakeRecordingsJSON = `{
  "count": 2,
  "recordings": [
    {
      "id": "abc123",
      "title": "Comfortably Numb",
      "score": 100,
      "length": 587266,
      "artist-credit": [{"name": "Pink Floyd", "artist": {"id": "83d91898-7763-47d7-b03b-b92132375c47", "name": "Pink Floyd"}}]
    },
    {
      "id": "def456",
      "title": "Comfortably Numb (Live)",
      "score": 95,
      "length": 473560,
      "artist-credit": [{"name": "Pink Floyd", "artist": {"id": "83d91898-7763-47d7-b03b-b92132375c47", "name": "Pink Floyd"}}]
    }
  ]
}`

func newTestClient(ts *httptest.Server) *musicbrainz.Client {
	cfg := musicbrainz.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return musicbrainz.NewClient(cfg)
}

func TestArtistsSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeArtistsJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Artists(context.Background(), "queen", 2)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestArtistsParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeArtistsJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Artists(context.Background(), "queen", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	a := items[0]
	if a.Rank != 1 {
		t.Errorf("Rank = %d, want 1", a.Rank)
	}
	if a.ID != "0383dadf-2a4e-4d10-a46a-e9e041da8eb3" {
		t.Errorf("ID = %q, want 0383dadf-...", a.ID)
	}
	if a.Name != "Queen" {
		t.Errorf("Name = %q, want Queen", a.Name)
	}
	if a.Type != "Group" {
		t.Errorf("Type = %q, want Group", a.Type)
	}
	if a.Country != "GB" {
		t.Errorf("Country = %q, want GB", a.Country)
	}
	if a.Score != 100 {
		t.Errorf("Score = %d, want 100", a.Score)
	}
	wantURL := "https://musicbrainz.org/artist/0383dadf-2a4e-4d10-a46a-e9e041da8eb3"
	if a.URL != wantURL {
		t.Errorf("URL = %q, want %q", a.URL, wantURL)
	}
}

func TestArtistsLimitRespected(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeArtistsJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Artists(context.Background(), "queen", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d, want 1", len(items))
	}
}

func TestArtistsRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeArtistsJSON)
	}))
	defer ts.Close()

	cfg := musicbrainz.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := musicbrainz.NewClient(cfg)

	_, err := c.Artists(context.Background(), "queen", 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

func TestRecordingsParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeRecordingsJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Recordings(context.Background(), "comfortably numb", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	r := items[0]
	if r.Rank != 1 {
		t.Errorf("Rank = %d, want 1", r.Rank)
	}
	if r.ID != "abc123" {
		t.Errorf("ID = %q, want abc123", r.ID)
	}
	if r.Title != "Comfortably Numb" {
		t.Errorf("Title = %q, want Comfortably Numb", r.Title)
	}
	if r.Artist != "Pink Floyd" {
		t.Errorf("Artist = %q, want Pink Floyd", r.Artist)
	}
	if r.LengthMs != 587266 {
		t.Errorf("LengthMs = %d, want 587266", r.LengthMs)
	}
	if r.Score != 100 {
		t.Errorf("Score = %d, want 100", r.Score)
	}
	wantURL := "https://musicbrainz.org/recording/abc123"
	if r.URL != wantURL {
		t.Errorf("URL = %q, want %q", r.URL, wantURL)
	}
}
