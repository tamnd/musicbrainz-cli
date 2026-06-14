package musicbrainz

import (
	"context"
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes musicbrainz as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/musicbrainz-cli/musicbrainz"
//
// The same Domain also builds the standalone musicbrainz binary (see cli.NewApp).
func init() { kit.Register(Domain{}) }

// Domain is the musicbrainz driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "musicbrainz",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "musicbrainz",
			Short:  "MusicBrainz music metadata — artists, recordings, and more",
			Long: `musicbrainz searches the MusicBrainz open music database for artists
and recordings. No API key required. Rate limit: 1 req/s.`,
			Site: Host,
			Repo: "https://github.com/tamnd/musicbrainz-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// artist: search for artists by name or Lucene query
	kit.Handle(app, kit.OpMeta{
		Name:    "artist",
		Group:   "read",
		List:    true,
		Summary: "Search for artists by name",
		Args:    []kit.Arg{{Name: "query", Help: "artist name or Lucene query (e.g. type:Group AND country:GB)"}},
	}, artistOp)

	// recording: search for recordings (tracks) by title or Lucene query
	kit.Handle(app, kit.OpMeta{
		Name:    "recording",
		Group:   "read",
		List:    true,
		Summary: "Search for recordings (tracks) by title",
		Args:    []kit.Arg{{Name: "query", Help: "title or Lucene query (e.g. title:comfortably+numb AND artist:pink+floyd)"}},
	}, recordingOp)
}

// newClient builds the client from host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type artistInput struct {
	Query  string        `kit:"arg" help:"artist name or Lucene query"`
	Limit  int           `kit:"flag,inherit" help:"max results (default 10, max 100)"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

type recordingInput struct {
	Query  string        `kit:"arg" help:"recording title or Lucene query"`
	Limit  int           `kit:"flag,inherit" help:"max results (default 10, max 100)"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

// --- handlers ---

func artistOp(ctx context.Context, in artistInput, emit func(Artist) error) error {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	items, err := in.Client.Artists(ctx, in.Query, limit)
	if err != nil {
		return mapErr(err)
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func recordingOp(ctx context.Context, in recordingInput, emit func(Recording) error) error {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	items, err := in.Client.Recordings(ctx, in.Query, limit)
	if err != nil {
		return mapErr(err)
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver: pure string functions, no network ---

// Classify turns an input into the canonical (type, id).
// MusicBrainz entities live at /{type}/{uuid}.
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty musicbrainz reference")
	}
	// Default to artist type for bare UUIDs or names.
	return "artist", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "artist":
		return "https://musicbrainz.org/artist/" + id, nil
	case "recording":
		return "https://musicbrainz.org/recording/" + id, nil
	default:
		return "", errs.Usage("musicbrainz has no resource type %q", uriType)
	}
}

// mapErr converts a library error into the kit error kind.
func mapErr(err error) error {
	return err
}
