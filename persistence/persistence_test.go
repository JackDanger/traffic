package persistence

import (
	"io/ioutil"
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
	fixtureSource, err := ioutil.ReadFile(util.Root() + "fixtures/browse-two-github-users.har")
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
	if parser.UnquoteJSON(archive.Source) != string(fixtureSource) {
		t.Errorf("Unexpected archive.Source length: %d, original: %d", len(parser.UnquoteJSON(archive.Source)), len(string(fixtureSource)))
	}
	if archive.CreatedAt.Sub(time.Now()) < 1*time.Second {
		t.Errorf("Unexpected archive.CreatedAt: %#v", archive.CreatedAt)
	}

	// Then retrieve everything in there
}
