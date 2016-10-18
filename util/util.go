package util

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path"
	"runtime"
	"time"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
)

// Fixture returns a Har from the ./fixtures directory
func Fixture() model.Har {
	fixture := "browse-two-github-users.har"

	har, err := parser.HarFromFile(Root() + "fixtures/" + fixture)
	if err != nil {
		panic(err) // should not happen
	}
	return *har
}

// MakeEntry retrieves one of the entriers from the fixture file
func MakeEntry() *model.Entry {
	return &Fixture().Entries[0]
}

// MakeRequest fixture
func MakeRequest() *model.Request {
	return MakeEntry().Request
}

// MakeResponse fixture
func MakeResponse() *model.Response {
	return MakeEntry().Response
}

// StringPtr is a simply way to get a pointer to a string literal
func StringPtr(ss string) *string {
	return &ss
}

// TimePtr is a reference to a time
func TimePtr(tt time.Time) *time.Time {
	return &tt
}

type pairwiseFunc func(key, val *string) bool

// Any retunrs true if any of the pairs provided match the function
func Any(pairs []model.SingleItemMap, f pairwiseFunc) bool {
	for _, pair := range pairs {
		if f(pair.Key, pair.Value) {
			return true
		}
	}
	return false
}

// Root points to this project's root directory
func Root() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename) + "/../"
}

// UUID generates a 16-byte globally unique token
func UUID() string {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		panic("how did we get a wrong-sized uuid?: " + string(uuid))
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

// EnvironmentGuess attempts to say whether we're running in the test
// environment. Should be replaced with `Environment()` (i.e. not a guess)
func EnvironmentGuess() string {
	if flag.Lookup("test.v") != nil {
		return "test"
	}
	// TODO: really, let's just use CLI flags for this.
	if uname() == "Darwin\n" {
		return "development"
	}
	return "production"
}

func uname() string {
	cmd := exec.Command("uname")

	var output bytes.Buffer
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		return ""
	}
	return output.String()
}
