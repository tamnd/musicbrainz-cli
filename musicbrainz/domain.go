package musicbrainz

import (
	"context"

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
			Short:  "MusicBrainz music metadata — artists, recordings, releases, labels, and more",
			Long: `musicbrainz searches the MusicBrainz open music database.

Search artists, recordings (tracks), releases (albums), release groups, and
labels. Look up any entity by its MBID (MusicBrainz Identifier).
No API key required. Rate limit: 1 req/s.`,
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

	// release: search for releases (albums) by title or Lucene query
	kit.Handle(app, kit.OpMeta{
		Name:    "release",
		Group:   "read",
		List:    true,
		Summary: "Search for releases (albums) by title",
		Args:    []kit.Arg{{Name: "query", Help: "title or Lucene query (e.g. title:OK+Computer AND artist:radiohead)"}},
	}, releaseOp)

	// release-group: search for release groups
	kit.Handle(app, kit.OpMeta{
		Name:    "release-group",
		Group:   "read",
		List:    true,
		Summary: "Search for release groups by title",
		Args:    []kit.Arg{{Name: "query", Help: "title or Lucene query"}},
	}, releaseGroupOp)

	// label: search for labels
	kit.Handle(app, kit.OpMeta{
		Name:    "label",
		Group:   "read",
		List:    true,
		Summary: "Search for record labels by name",
		Args:    []kit.Arg{{Name: "query", Help: "label name or Lucene query"}},
	}, labelOp)

	// get artist: lookup artist by MBID with releases
	kit.Handle(app, kit.OpMeta{
		Name:    "get-artist",
		Group:   "read",
		Single:  true,
		Summary: "Get full artist record by MBID (includes releases)",
		Args:    []kit.Arg{{Name: "mbid", Help: "MusicBrainz Identifier (UUID) for the artist"}},
	}, getArtistOp)
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

type queryInput struct {
	Query  string  `kit:"arg" help:"search query"`
	Limit  int     `kit:"flag,inherit" help:"max results (default 10, max 100)"`
	Client *Client `kit:"inject"`
}

type mbidInput struct {
	MBID   string  `kit:"arg" help:"MusicBrainz Identifier (UUID)"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func artistOp(ctx context.Context, in queryInput, emit func(*Artist) error) error {
	items, err := in.Client.Artists(ctx, in.Query, in.Limit)
	if err != nil {
		return err
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

func recordingOp(ctx context.Context, in queryInput, emit func(*Recording) error) error {
	items, err := in.Client.Recordings(ctx, in.Query, in.Limit)
	if err != nil {
		return err
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

func releaseOp(ctx context.Context, in queryInput, emit func(*Release) error) error {
	items, err := in.Client.Releases(ctx, in.Query, in.Limit)
	if err != nil {
		return err
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

func releaseGroupOp(ctx context.Context, in queryInput, emit func(*ReleaseGroup) error) error {
	items, err := in.Client.ReleaseGroups(ctx, in.Query, in.Limit)
	if err != nil {
		return err
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

func labelOp(ctx context.Context, in queryInput, emit func(*Label) error) error {
	items, err := in.Client.Labels(ctx, in.Query, in.Limit)
	if err != nil {
		return err
	}
	for i := range items {
		if err := emit(&items[i]); err != nil {
			return err
		}
	}
	return nil
}

func getArtistOp(ctx context.Context, in mbidInput, emit func(*ArtistDetail) error) error {
	detail, err := in.Client.GetArtist(ctx, in.MBID)
	if err != nil {
		return err
	}
	return emit(detail)
}

// --- Resolver: pure string functions, no network ---

// Classify turns an input into the canonical (type, id).
// MusicBrainz entities live at /{type}/{uuid}.
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty musicbrainz reference")
	}
	return "artist", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "artist":
		return "https://musicbrainz.org/artist/" + id, nil
	case "recording":
		return "https://musicbrainz.org/recording/" + id, nil
	case "release":
		return "https://musicbrainz.org/release/" + id, nil
	case "release-group":
		return "https://musicbrainz.org/release-group/" + id, nil
	case "label":
		return "https://musicbrainz.org/label/" + id, nil
	default:
		return "", errs.Usage("musicbrainz has no resource type %q", uriType)
	}
}
