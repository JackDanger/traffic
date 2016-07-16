package parser

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	fixture := "../fixtures/browse-two-github-users.har"
	out := "../fixtures/browse-two-github-users.har.roundtrip"
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	pathToFixture := cwd + "/" + fixture
	pathToOut := cwd + "/" + out

	jsonSource, err := ioutil.ReadFile(pathToFixture)
	if err != nil {
		t.Fatal(err)
	}

	instance, err := HarFromFile(pathToFixture)
	if err != nil {
		t.Fatal(err)
	}

	wrapper := &HarWrapper{Har: instance}
	roundtrip, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	normalizedSource := normalizedJSON(jsonSource, t)
	normalizedRoundtrip := normalizedJSON(roundtrip, t)
	if normalizedSource != normalizedRoundtrip {
		ioutil.WriteFile(pathToOut, []byte(normalizedRoundtrip), 0600)
		ioutil.WriteFile(pathToOut+".escapedoriginal", []byte(normalizedSource), 0600)
		t.Errorf("the json source wasn't the same.\n compare with: \ndiff -w fixtures/browse-two-github-users.har.roundtrip*")
	}
}

func TestUnquoteJSON(t *testing.T) {
	type structure struct {
		Link string `json:"link"`
	}
	source := `{"link":"<a href=\"/substitution.html\">search & replace</a>"}`
	withQuotes := `{"link":"\u003ca href=\"/substitution.html\"\u003esearch \u0026 replace\u003c/a\u003e"}`
	holdsIt := &structure{}
	err := json.Unmarshal([]byte(source), holdsIt)
	if err != nil {
		t.Fatal(err)
	}

	quoted, err := json.Marshal(holdsIt)
	if err != nil {
		t.Fatal(err)
	}

	if string(quoted) != withQuotes {
		t.Errorf("should have been equal:\n%s\n%s", quoted, withQuotes)
	}

	unquoted := UnquoteJSON(string(quoted))
	if unquoted != source {
		t.Errorf("should have been equal:\n%s\n%s", unquoted, source)
	}

}

// test helpers

func normalizedJSON(input []byte, t *testing.T) string {
	output := bytes.Buffer{}
	err := json.Compact(&output, input)
	if err != nil {
		t.Fatal(err)
	}
	return UnquoteJSON(output.String())
}
