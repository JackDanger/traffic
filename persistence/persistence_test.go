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

func TestArchiveCreate(t *testing.T) {
}

func TestTransformCreate(t *testing.T) {
}

func TestListArchives(t *testing.T) {
	db, err := NewDb()
	if err != nil {
		t.Fatal(err)
	}
	// Get a har from a test fixture
	bytes, err := ioutil.ReadFile(util.Root() + "fixtures/browse-two-github-users.har")
	// We trim it because we're going to do a byte-for-byte comparison of the
	// source later on to ensure it can roundtrip to the database without
	// unexpected modifications
	fixtureSource := strings.TrimRight(string(bytes), "\n")
	if err != nil {
		t.Fatal(err)
	}

	har, err := parser.HarFromFile(util.Root() + "fixtures/browse-two-github-users.har")
	if err != nil {
		t.Fatal(err)
	}

	// Store it in the db
	archive, err := MakeArchive("some name", "any description", har)
	if err != nil {
		t.Fatal(err)
	}
	if err := archive.Create(db); err != nil {
		t.Fatal(err)
	}
	if archive == nil {
		t.Fatal("Archive not expected to be nil")
	}
	if archive.Name != "some name" {
		t.Errorf("Expected archive.Name to be \"some name\", got: %s", archive.Name)
	}
	if archive.Description != "any description" {
		t.Errorf("Expected archive.Description to be \"any description\", got: %s", archive.Description)
	}
	if parser.UnquoteJSON(archive.Source) != fixtureSource {
		t.Errorf("Unexpected archive.Source length: %d, original: %d", len(archive.Source), len(fixtureSource))
	}
	// If the time since it was created is less than "-1 * time.Second" or "1 second ago"
	if archive.CreatedAt.Sub(time.Now()) < -time.Second {
		t.Errorf("Unexpected archive.CreatedAt: %#v, duration: %#v (%#v)", archive.CreatedAt, archive.CreatedAt.Sub(time.Now()), time.Second)
	}
	// The updated_at should be about the same as created_at
	if archive.UpdatedAt.Sub(time.Now()) < -time.Second {
		t.Errorf("Unexpected archive.UpdatedAt: %#v, duration: %#v (%#v)", archive.UpdatedAt, archive.UpdatedAt.Sub(time.Now()), time.Second)
	}

	// Then retrieve everything in there
	archives, err := db.ListArchives()
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) < 1 {
		t.Error("Did not retrieve any archives")
	}
	retrieved := archives[0]

	if retrieved.Source != archive.Source {
		t.Errorf("Unexpected Source retrieved: %d / %d", len(retrieved.Source), len(archive.Source))
	}
	if retrieved.CreatedAt != archive.CreatedAt {
		t.Errorf("Unexpected CreatedAt retrieved: %s, expected: %s", retrieved.CreatedAt, archive.CreatedAt)
	}
}

func TestListTransformsFor(t *testing.T) {
	// TODO
}
