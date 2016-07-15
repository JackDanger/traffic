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

	if normalizedJSON(escaped(jsonSource), t) != normalizedJSON(roundtrip, t) {
		ioutil.WriteFile(pathToOut, escaped(roundtrip), 0600)
		ioutil.WriteFile(pathToOut+".escapedoriginal", escaped(jsonSource), 0600)
		t.Errorf("the json source wasn't the same.\n compare with: \ndiff -w fixtures/browse-two-github-users.har.roundtrip*")
	}
}

// test helpers

func normalizedJSON(input []byte, t *testing.T) string {
	output := bytes.Buffer{}
	err := json.Compact(&output, input)
	if err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func escaped(input []byte) []byte {
	buf := &bytes.Buffer{}
	json.HTMLEscape(buf, input)
	return buf.Bytes()
}
