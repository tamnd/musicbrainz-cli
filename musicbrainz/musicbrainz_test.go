package musicbrainz_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/musicbrainz-cli/musicbrainz"
)

// --- test data ---

const fakeArtistsJSON = `{
  "count": 2,
  "artists": [
    {
      "id": "0383dadf-2a4e-4d10-a46a-e9e041da8eb3",
      "name": "Queen",
      "type": "Group",
      "country": "GB",
      "area": {"name": "United Kingdom"},
      "life-span": {"begin": "1970", "ended": false},
      "score": 100
    },
    {
      "id": "f27ec8db-af05-4f36-916e-3d57f91ecf7e",
      "name": "Michael Jackson",
      "type": "Person",
      "country": "US",
      "area": {"name": "United States"},
      "life-span": {"begin": "1958-08-29", "end": "2009-06-25", "ended": true},
      "score": 85
    }
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
      "artist-credit": [{"name": "Pink Floyd", "artist": {"id": "83d91898-7763-47d7-b03b-b92132375c47", "name": "Pink Floyd"}}],
      "releases": [{"title": "The Wall", "date": "1979"}]
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

const fakeReleasesJSON = `{
  "count": 1,
  "releases": [
    {
      "id": "b84ee12a-1dc4-4a84-b1c0-b7b159e18174",
      "title": "OK Computer",
      "status": "Official",
      "date": "1997-05-21",
      "country": "GB",
      "label-info": [{"label": {"name": "Parlophone"}}],
      "release-group": {"primary-type": "Album"},
      "artist-credit": [{"name": "Radiohead", "artist": {"id": "a74b1b7f-71a5-4011-9441-d0b5e4122711", "name": "Radiohead"}}],
      "score": 100
    }
  ]
}`

const fakeReleaseGroupsJSON = `{
  "count": 1,
  "release-groups": [
    {
      "id": "f5093c06-1a98-4cf3-8b75-4b0c3b0c0b0c",
      "title": "The Dark Side of the Moon",
      "primary-type": "Album",
      "first-release-date": "1973-03-01",
      "artist-credit": [{"name": "Pink Floyd", "artist": {"id": "83d91898-7763-47d7-b03b-b92132375c47", "name": "Pink Floyd"}}],
      "score": 100
    }
  ]
}`

const fakeLabelsJSON = `{
  "count": 1,
  "labels": [
    {
      "id": "2c9b9182-cb7e-4d7c-8ef9-0a9d0a16d7b8",
      "name": "Parlophone",
      "type": "Original Production",
      "country": "GB",
      "score": 100
    }
  ]
}`

const fakeArtistDetailJSON = `{
  "id": "b10bbbfc-cf9e-42e0-be17-e2c3e1d2600d",
  "name": "The Beatles",
  "type": "Group",
  "country": "GB",
  "area": {"name": "United Kingdom"},
  "life-span": {"begin": "1957-07", "end": "1970-04-10", "ended": true},
  "releases": [
    {"id": "rel1", "title": "Abbey Road", "date": "1969-09-26", "status": "Official"},
    {"id": "rel2", "title": "Let It Be", "date": "1970-05-08", "status": "Official"}
  ]
}`

// --- helpers ---

func newTestClient(ts *httptest.Server) *musicbrainz.Client {
	cfg := musicbrainz.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return musicbrainz.NewClient(cfg)
}

func serve(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, body)
	}))
}

// --- tests ---

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
	ts := serve(fakeArtistsJSON)
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
	if a.MBID != "0383dadf-2a4e-4d10-a46a-e9e041da8eb3" {
		t.Errorf("MBID = %q, unexpected", a.MBID)
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
	if a.Area != "United Kingdom" {
		t.Errorf("Area = %q, want United Kingdom", a.Area)
	}
	if a.BeginDate != "1970" {
		t.Errorf("BeginDate = %q, want 1970", a.BeginDate)
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
	ts := serve(fakeArtistsJSON)
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

func TestArtistsNonRetryable4xx(t *testing.T) {
	hits := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Artists(context.Background(), "notfound", 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if hits != 1 {
		t.Errorf("server saw %d hits, want 1 (no retry on 404)", hits)
	}
}

func TestRecordingsParsesItems(t *testing.T) {
	ts := serve(fakeRecordingsJSON)
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
	if r.MBID != "abc123" {
		t.Errorf("MBID = %q, want abc123", r.MBID)
	}
	if r.Title != "Comfortably Numb" {
		t.Errorf("Title = %q, want Comfortably Numb", r.Title)
	}
	if r.ArtistCredit != "Pink Floyd" {
		t.Errorf("ArtistCredit = %q, want Pink Floyd", r.ArtistCredit)
	}
	if r.LengthMs != 587266 {
		t.Errorf("LengthMs = %d, want 587266", r.LengthMs)
	}
	if r.ReleaseTitle != "The Wall" {
		t.Errorf("ReleaseTitle = %q, want The Wall", r.ReleaseTitle)
	}
	if r.Score != 100 {
		t.Errorf("Score = %d, want 100", r.Score)
	}
	wantURL := "https://musicbrainz.org/recording/abc123"
	if r.URL != wantURL {
		t.Errorf("URL = %q, want %q", r.URL, wantURL)
	}
}

func TestReleasesParsesItems(t *testing.T) {
	ts := serve(fakeReleasesJSON)
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Releases(context.Background(), "ok computer radiohead", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	r := items[0]
	if r.MBID != "b84ee12a-1dc4-4a84-b1c0-b7b159e18174" {
		t.Errorf("MBID = %q, unexpected", r.MBID)
	}
	if r.Title != "OK Computer" {
		t.Errorf("Title = %q, want OK Computer", r.Title)
	}
	if r.Status != "Official" {
		t.Errorf("Status = %q, want Official", r.Status)
	}
	if r.Label != "Parlophone" {
		t.Errorf("Label = %q, want Parlophone", r.Label)
	}
	if r.Type != "Album" {
		t.Errorf("Type = %q, want Album", r.Type)
	}
	if r.ArtistCredit != "Radiohead" {
		t.Errorf("ArtistCredit = %q, want Radiohead", r.ArtistCredit)
	}
	if r.Score != 100 {
		t.Errorf("Score = %d, want 100", r.Score)
	}
}

func TestReleaseGroupsParsesItems(t *testing.T) {
	ts := serve(fakeReleaseGroupsJSON)
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.ReleaseGroups(context.Background(), "dark side of the moon", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	rg := items[0]
	if rg.Title != "The Dark Side of the Moon" {
		t.Errorf("Title = %q, unexpected", rg.Title)
	}
	if rg.Type != "Album" {
		t.Errorf("Type = %q, want Album", rg.Type)
	}
	if rg.FirstReleaseDate != "1973-03-01" {
		t.Errorf("FirstReleaseDate = %q, want 1973-03-01", rg.FirstReleaseDate)
	}
	if rg.ArtistCredit != "Pink Floyd" {
		t.Errorf("ArtistCredit = %q, want Pink Floyd", rg.ArtistCredit)
	}
}

func TestLabelsParsesItems(t *testing.T) {
	ts := serve(fakeLabelsJSON)
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Labels(context.Background(), "parlophone", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	l := items[0]
	if l.Name != "Parlophone" {
		t.Errorf("Name = %q, want Parlophone", l.Name)
	}
	if l.Type != "Original Production" {
		t.Errorf("Type = %q, want Original Production", l.Type)
	}
	if l.Country != "GB" {
		t.Errorf("Country = %q, want GB", l.Country)
	}
}

func TestGetArtistParsesDetail(t *testing.T) {
	ts := serve(fakeArtistDetailJSON)
	defer ts.Close()

	c := newTestClient(ts)
	detail, err := c.GetArtist(context.Background(), "b10bbbfc-cf9e-42e0-be17-e2c3e1d2600d")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Name != "The Beatles" {
		t.Errorf("Name = %q, want The Beatles", detail.Name)
	}
	if detail.MBID != "b10bbbfc-cf9e-42e0-be17-e2c3e1d2600d" {
		t.Errorf("MBID = %q, unexpected", detail.MBID)
	}
	if detail.Country != "GB" {
		t.Errorf("Country = %q, want GB", detail.Country)
	}
	if detail.BeginDate != "1957-07" {
		t.Errorf("BeginDate = %q, want 1957-07", detail.BeginDate)
	}
	if len(detail.Releases) != 2 {
		t.Fatalf("len(Releases) = %d, want 2", len(detail.Releases))
	}
	if detail.Releases[0].Title != "Abbey Road" {
		t.Errorf("Releases[0].Title = %q, want Abbey Road", detail.Releases[0].Title)
	}
}

const fakeRecordingDetailJSON = `{
  "id": "29db0dec-2caf-4c0e-b542-20c91b8d7eec",
  "title": "Smells Like Teen Spirit",
  "length": 277000,
  "first-release-date": "1991-09-10",
  "artist-credit": [{"name": "Nirvana", "artist": {"id": "5b11f4ce-a62d-471e-81fc-a69a8278c7da", "name": "Nirvana"}}],
  "releases": [
    {"id": "a7ffd0c7-3d24-3d18-b97e-49fc83cf8c97", "title": "Nevermind", "date": "1991-09-24", "status": "Official"}
  ]
}`

const fakeReleaseDetailJSON = `{
  "id": "a7ffd0c7-3d24-3d18-b97e-49fc83cf8c97",
  "title": "Nevermind",
  "status": "Official",
  "date": "1991-09-24",
  "country": "US",
  "label-info": [{"label": {"name": "DGC Records"}}],
  "release-group": {"primary-type": "Album"},
  "artist-credit": [{"name": "Nirvana", "artist": {"id": "5b11f4ce-a62d-471e-81fc-a69a8278c7da", "name": "Nirvana"}}]
}`

func TestGetRecordingParsesDetail(t *testing.T) {
	ts := serve(fakeRecordingDetailJSON)
	defer ts.Close()

	c := newTestClient(ts)
	detail, err := c.GetRecording(context.Background(), "29db0dec-2caf-4c0e-b542-20c91b8d7eec")
	if err != nil {
		t.Fatal(err)
	}
	if detail.MBID != "29db0dec-2caf-4c0e-b542-20c91b8d7eec" {
		t.Errorf("MBID = %q, unexpected", detail.MBID)
	}
	if detail.Title != "Smells Like Teen Spirit" {
		t.Errorf("Title = %q, want Smells Like Teen Spirit", detail.Title)
	}
	if detail.LengthMs != 277000 {
		t.Errorf("LengthMs = %d, want 277000", detail.LengthMs)
	}
	if detail.FirstRelease != "1991-09-10" {
		t.Errorf("FirstRelease = %q, want 1991-09-10", detail.FirstRelease)
	}
	if detail.ArtistCredit != "Nirvana" {
		t.Errorf("ArtistCredit = %q, want Nirvana", detail.ArtistCredit)
	}
	if len(detail.Releases) != 1 {
		t.Fatalf("len(Releases) = %d, want 1", len(detail.Releases))
	}
	if detail.Releases[0].Title != "Nevermind" {
		t.Errorf("Releases[0].Title = %q, want Nevermind", detail.Releases[0].Title)
	}
	wantURL := "https://musicbrainz.org/recording/29db0dec-2caf-4c0e-b542-20c91b8d7eec"
	if detail.URL != wantURL {
		t.Errorf("URL = %q, want %q", detail.URL, wantURL)
	}
}

func TestGetReleaseParsesDetail(t *testing.T) {
	ts := serve(fakeReleaseDetailJSON)
	defer ts.Close()

	c := newTestClient(ts)
	detail, err := c.GetRelease(context.Background(), "a7ffd0c7-3d24-3d18-b97e-49fc83cf8c97")
	if err != nil {
		t.Fatal(err)
	}
	if detail.MBID != "a7ffd0c7-3d24-3d18-b97e-49fc83cf8c97" {
		t.Errorf("MBID = %q, unexpected", detail.MBID)
	}
	if detail.Title != "Nevermind" {
		t.Errorf("Title = %q, want Nevermind", detail.Title)
	}
	if detail.Status != "Official" {
		t.Errorf("Status = %q, want Official", detail.Status)
	}
	if detail.Date != "1991-09-24" {
		t.Errorf("Date = %q, want 1991-09-24", detail.Date)
	}
	if detail.Country != "US" {
		t.Errorf("Country = %q, want US", detail.Country)
	}
	if detail.Label != "DGC Records" {
		t.Errorf("Label = %q, want DGC Records", detail.Label)
	}
	if detail.Type != "Album" {
		t.Errorf("Type = %q, want Album", detail.Type)
	}
	if detail.ArtistCredit != "Nirvana" {
		t.Errorf("ArtistCredit = %q, want Nirvana", detail.ArtistCredit)
	}
	wantURL := "https://musicbrainz.org/release/a7ffd0c7-3d24-3d18-b97e-49fc83cf8c97"
	if detail.URL != wantURL {
		t.Errorf("URL = %q, want %q", detail.URL, wantURL)
	}
}
