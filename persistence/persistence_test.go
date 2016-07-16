package persistence

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/util"
)

func TestListArchives(t *testing.T) {
	db, err := NewDb("test")
	if err != nil {
		t.Fatal(err)
	}
	// Get a har from a test fixture
	bytes, err := ioutil.ReadFile(util.Root() + "fixtures/browse-two-github-users.har")
	fixtureSource := strings.TrimRight(string(bytes), "\n")
	if err != nil {
		t.Fatal(err)
	}

	har, err := parser.HarFromFile(util.Root() + "fixtures/browse-two-github-users.har")
	if err != nil {
		t.Fatal(err)
	}

	// Store it in the db
	archive, err := db.Store(har)
	if err != nil {
		t.Fatal(err)
	}
	if archive == nil {
		t.Fatal("Archive not expected to be nil")
	}
	if archive.Token == "" {
		t.Error("Expected archive.Token to not be blank")
	}
	if parser.UnquoteJSON(archive.Source) != fixtureSource {
		t.Errorf("Unexpected archive.Source length: %d, original: %d", len(archive.Source), len(fixtureSource))
	}
	// If the time since it was created is less than "-1 * time.Second" or "1 second ago"
	if archive.CreatedAt.Sub(time.Now()) < -time.Second {
		t.Errorf("Unexpected archive.CreatedAt: %#v, duration: %#v (%#v)", archive.CreatedAt, archive.CreatedAt.Sub(time.Now()), time.Second)
	}

	// Then retrieve everything in there
	hars, err := db.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(hars) < 1 {
		t.Error("Did not retrieve any records")
	}
}
