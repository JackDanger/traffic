package parser

import (
	"encoding/json"
	"io/ioutil"

	"github.com/JackDanger/traffic/model"
)

// HarWrapper exists because the HAR file contains a top-level key called "log"
// which most of the time we'll pretend isn't there.
type HarWrapper struct {
	Har *model.Har `json:"log"`
}

// HarFrom parses the .har file and returns a full Har instance.
func HarFrom(source string) (*model.Har, error) {
	wrapper := &HarWrapper{}
	err := json.Unmarshal([]byte(source), &wrapper)
	if err != nil {
		return nil, err
	}
	return wrapper.Har, nil
}

// HarFromFile parses the .har file at the given path and returns a full Har instance.
func HarFromFile(path string) (*model.Har, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return HarFrom(string(bytes))
}

// HarToJSON does the opposite of HarFrom
func HarToJSON(har *model.Har) string {
	wrapper := &HarWrapper{Har: har}
	j, _ := json.MarshalIndent(wrapper, "", "  ")
	return string(j)
}
