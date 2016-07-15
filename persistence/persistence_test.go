package persistence

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/util"
)

func TestListArchives(t *testing.T) {
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
	archive, err := Store(har)
	if err != nil {
		t.Fatal(err)
	}
	if archive == nil {
		t.Fatal("Archive not expected to be nil")
	}
	if archive.Token == "" {
		t.Error("Expected archive.Token to not be blank")
	}
	if archive.Source != string(fixtureSource) {
		t.Errorf("Unexpected archive.Source: %s", archive.Source)
	}
	if archive.CreatedAt.Sub(time.Now()) < 1*time.Second {
		t.Errorf("Unexpected archive.Source: %s", archive.Source)
	}

	// Then retrieve everything in there
}
