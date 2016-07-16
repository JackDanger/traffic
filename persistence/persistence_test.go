package persistence

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/util"
)

func TestMain(m *testing.M) {
	// Creates the database if necessary
	db, err := NewDb()
	if err != nil {
		panic(err)
	}

	m.Run()

	db.Truncate()
}

func TestListArchives(t *testing.T) {
	db, err := NewDb()
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
	archive, err := db.Create(MakeArchive(har))
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
	archives, err := db.ListArchives()
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) < 1 {
		t.Error("Did not retrieve any records")
	}
	retrieved := archives[0]

	if retrieved.Token != archive.Token {
		t.Errorf("Unexpected Token retrieved: %s, expected: %s", retrieved.Token, archive.Token)
	}
	if retrieved.Source != archive.Source {
		t.Errorf("Unexpected Source retrieved: %d / %d", len(retrieved.Source), len(archive.Source))
	}
	if retrieved.CreatedAt != archive.CreatedAt {
		t.Errorf("Unexpected CreatedAt retrieved: %s, expected: %s", retrieved.CreatedAt, archive.CreatedAt)
	}
}
