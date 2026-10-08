package dash

import (
	"regexp"
	"testing"
)

// The page shares an origin with the admin API, so every script and stylesheet
// it pulls from a CDN must name an exact version and carry an SRI hash.
func TestIndexPinsExternalAssets(t *testing.T) {
	page, err := Assets.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}
	tags := regexp.MustCompile(`(?s)<(?:script|link)\b[^>]*https://[^>]*>`).FindAll(page, -1)
	if len(tags) == 0 {
		t.Fatal("index.html references no external assets; update this test")
	}
	pinned := regexp.MustCompile(`https://[^"]+@\d+\.\d+\.\d+/`)
	integrity := regexp.MustCompile(`integrity="sha(256|384|512)-[^"]+"`)
	for _, tag := range tags {
		if !pinned.Match(tag) || !integrity.Match(tag) {
			t.Errorf("external asset is not pinned to an exact version with integrity: %s", tag)
		}
	}
}
