package musicbrainz

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string functions.
// HTTP behaviour is covered in musicbrainz_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "musicbrainz" {
		t.Errorf("Scheme = %q, want musicbrainz", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "musicbrainz" {
		t.Errorf("Identity.Binary = %q, want musicbrainz", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	typ, id, err := Domain{}.Classify("0383dadf-2a4e-4d10-a46a-e9e041da8eb3")
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}
	if typ != "artist" {
		t.Errorf("type = %q, want artist", typ)
	}
	if id != "0383dadf-2a4e-4d10-a46a-e9e041da8eb3" {
		t.Errorf("id = %q, want 0383dadf-...", id)
	}
}

func TestClassifyEmpty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestLocateArtist(t *testing.T) {
	got, err := Domain{}.Locate("artist", "0383dadf-2a4e-4d10-a46a-e9e041da8eb3")
	want := "https://musicbrainz.org/artist/0383dadf-2a4e-4d10-a46a-e9e041da8eb3"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateRecording(t *testing.T) {
	got, err := Domain{}.Locate("recording", "abc123")
	want := "https://musicbrainz.org/recording/abc123"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("label", "xyz")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}
