package traffic

import (
	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
	"os"
	"testing"
)

// Fixture returns a Har from the ./fixtures directory
func Fixture(t *testing.T, name ...*string) model.Har {
	var fixture string
	if len(name) == 0 {
		fixture = "../fixtures/browse-two-github-users.har"
	} else {
		fixture = *name[0]
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	pathToFixture := cwd + "/" + fixture
	har, err := parser.HarFrom(pathToFixture)
	if err != nil {
		t.Fatal(err)
	}
	return *har
}

// MakeEntry retrieves one of the entriers from the fixture file
func MakeEntry(t *testing.T) *model.Entry {
	return &Fixture(t).Entries[0]
}

// MakeRequest fixture
func MakeRequest(t *testing.T) *model.Request {
	return MakeEntry(t).Request
}

// MakeResponse fixture
func MakeResponse(t *testing.T) *model.Response {
	return MakeEntry(t).Response
}
