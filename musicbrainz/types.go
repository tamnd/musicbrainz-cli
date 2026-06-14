package musicbrainz

// Artist is one result from a MusicBrainz artist search.
type Artist struct {
	Rank      int    `json:"rank"`
	MBID      string `json:"mbid"`
	Name      string `json:"name"`
	Type      string `json:"type"`       // Group, Person, Orchestra, Choir, Other
	Country   string `json:"country"`    // ISO 3166-1 alpha-2 country code
	Area      string `json:"area"`       // area name
	BeginDate string `json:"begin_date"` // YYYY or YYYY-MM or YYYY-MM-DD
	EndDate   string `json:"end_date"`
	Score     int    `json:"score"` // relevance 0-100
	URL       string `json:"url"`   // https://musicbrainz.org/artist/{id}
}

// Recording is one result from a MusicBrainz recording search.
type Recording struct {
	Rank         int    `json:"rank"`
	MBID         string `json:"mbid"`
	Title        string `json:"title"`
	ArtistCredit string `json:"artist_credit"` // primary artist-credit name(s)
	ReleaseTitle string `json:"release_title"`
	ReleaseDate  string `json:"release_date"`
	LengthMs     int    `json:"length_ms"` // duration in ms (0 if unknown)
	Score        int    `json:"score"`     // relevance 0-100
	URL          string `json:"url"`       // https://musicbrainz.org/recording/{id}
}

// Release is one result from a MusicBrainz release search.
type Release struct {
	Rank         int    `json:"rank"`
	MBID         string `json:"mbid"`
	Title        string `json:"title"`
	Status       string `json:"status"`        // Official, Promotional, Bootleg, Pseudo-Release
	Date         string `json:"date"`          // YYYY or YYYY-MM or YYYY-MM-DD
	Country      string `json:"country"`       // ISO 3166-1 alpha-2
	Label        string `json:"label"`         // first label name
	Type         string `json:"type"`          // Album, Single, EP, Compilation, etc.
	ArtistCredit string `json:"artist_credit"` // primary artist-credit name(s)
	Score        int    `json:"score"`
	URL          string `json:"url"` // https://musicbrainz.org/release/{id}
}

// ReleaseGroup is one result from a MusicBrainz release-group search.
type ReleaseGroup struct {
	Rank             int    `json:"rank"`
	MBID             string `json:"mbid"`
	Title            string `json:"title"`
	Type             string `json:"type"`               // Album, Single, EP, Broadcast, Other
	FirstReleaseDate string `json:"first_release_date"` // YYYY or YYYY-MM or YYYY-MM-DD
	ArtistCredit     string `json:"artist_credit"`
	Score            int    `json:"score"`
	URL              string `json:"url"` // https://musicbrainz.org/release-group/{id}
}

// Label is one result from a MusicBrainz label search.
type Label struct {
	Rank    int    `json:"rank"`
	MBID    string `json:"mbid"`
	Name    string `json:"name"`
	Type    string `json:"type"`    // Original Production, Bootleg Production, etc.
	Country string `json:"country"` // ISO 3166-1 alpha-2
	Score   int    `json:"score"`
	URL     string `json:"url"` // https://musicbrainz.org/label/{id}
}

// ArtistDetail is the full artist record from a lookup by MBID.
type ArtistDetail struct {
	MBID      string          `json:"mbid"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Country   string          `json:"country"`
	Area      string          `json:"area"`
	BeginDate string          `json:"begin_date"`
	EndDate   string          `json:"end_date"`
	URL       string          `json:"url"`
	Releases  []ReleaseStub   `json:"releases"`
}

// ReleaseStub is a brief release summary embedded in ArtistDetail.
type ReleaseStub struct {
	MBID   string `json:"mbid"`
	Title  string `json:"title"`
	Date   string `json:"date"`
	Status string `json:"status"`
}

// --- wire types (unexported, only used inside musicbrainz.go for JSON decode) ---

type artistSearchResponse struct {
	Count   int          `json:"count"`
	Artists []artistItem `json:"artists"`
}

type artistItem struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	Country  string       `json:"country"`
	Area     wireArea     `json:"area"`
	LifeSpan wireLifeSpan `json:"life-span"`
	Score    int          `json:"score"`
}

type wireArea struct {
	Name string `json:"name"`
}

type wireLifeSpan struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
	Ended bool   `json:"ended"`
}

type recordingSearchResponse struct {
	Count      int             `json:"count"`
	Recordings []recordingItem `json:"recordings"`
}

type recordingItem struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Score        int          `json:"score"`
	Length       int          `json:"length"`
	ArtistCredit []creditItem `json:"artist-credit"`
	Releases     []wireRelStub `json:"releases"`
}

type creditItem struct {
	Name   string       `json:"name"`
	Artist creditArtist `json:"artist"`
}

type creditArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type wireRelStub struct {
	Title string `json:"title"`
	Date  string `json:"date"`
}

type releaseSearchResponse struct {
	Count    int           `json:"count"`
	Releases []releaseItem `json:"releases"`
}

type releaseItem struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	Status       string        `json:"status"`
	Date         string        `json:"date"`
	Country      string        `json:"country"`
	LabelInfo    []wireLabelInfo `json:"label-info"`
	ReleaseGroup wireReleaseGroup `json:"release-group"`
	ArtistCredit []creditItem  `json:"artist-credit"`
	Score        int           `json:"score"`
}

type wireLabelInfo struct {
	Label wireLabel `json:"label"`
}

type wireLabel struct {
	Name string `json:"name"`
}

type wireReleaseGroup struct {
	PrimaryType string `json:"primary-type"`
}

type releaseGroupSearchResponse struct {
	Count         int                `json:"count"`
	ReleaseGroups []releaseGroupItem `json:"release-groups"`
}

type releaseGroupItem struct {
	ID               string       `json:"id"`
	Title            string       `json:"title"`
	PrimaryType      string       `json:"primary-type"`
	FirstReleaseDate string       `json:"first-release-date"`
	ArtistCredit     []creditItem `json:"artist-credit"`
	Score            int          `json:"score"`
}

type labelSearchResponse struct {
	Count  int         `json:"count"`
	Labels []labelItem `json:"labels"`
}

type labelItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Country string `json:"country"`
	Score   int    `json:"score"`
}

type wireArtistDetail struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	Country  string       `json:"country"`
	Area     wireArea     `json:"area"`
	LifeSpan wireLifeSpan `json:"life-span"`
	Releases []wireRelease `json:"releases"`
}

type wireRelease struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Date   string `json:"date"`
	Status string `json:"status"`
}
