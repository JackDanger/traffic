package parser

import (
	"encoding/json"
	"io/ioutil"

	"github.com/JackDanger/traffic/model"
)

// HarWrapper exists because the HAR file contains a top-level key called "log"
// which most of the time we'll pretend isn't there.
type HarWrapper struct {
	Har model.Har `json:"log"`
}

// HarFrom parses the .har file and returns a full Har instance.
func HarFrom(path string) (*model.Har, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	wrapper := &HarWrapper{}
	err = json.Unmarshal(bytes, &wrapper)
	if err != nil {
		return nil, err
	}
	return &wrapper.Har, nil
}
