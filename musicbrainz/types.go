package musicbrainz

// Artist is one result from a MusicBrainz artist search.
type Artist struct {
	Rank    int    `json:"rank"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`    // "Group", "Person", "Orchestra", "Choir", …
	Country string `json:"country"` // ISO 3166-1 alpha-2 country code
	Score   int    `json:"score"`   // relevance 0-100
	URL     string `json:"url"`     // https://musicbrainz.org/artist/{id}
}

// Recording is one result from a MusicBrainz recording search.
type Recording struct {
	Rank     int    `json:"rank"`
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`    // primary artist-credit name
	LengthMs int    `json:"length_ms"` // duration in ms (0 if unknown)
	Score    int    `json:"score"`     // relevance 0-100
	URL      string `json:"url"`       // https://musicbrainz.org/recording/{id}
}

// unexported: only used inside musicbrainz.go for JSON decode

type artistSearchResponse struct {
	Count   int          `json:"count"`
	Artists []artistItem `json:"artists"`
}

type artistItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Country string `json:"country"`
	Score   int    `json:"score"`
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
}

type creditItem struct {
	Name   string       `json:"name"`
	Artist creditArtist `json:"artist"`
}

type creditArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
