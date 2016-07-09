package parser

import (
	"encoding/json"
	"github.com/JackDanger/traffic/model"
	"io/ioutil"
)

// The HAR file contains a top-level key called "log" which we'll pretend isn't
// there.
type harWrapper struct {
	Har model.Har `json:"log"`
}

// HarFrom parses the .har file and returns a full Har instance.
func HarFrom(file string) (*model.Har, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	wrapper := &harWrapper{}
	err = json.Unmarshal(bytes, &wrapper)
	if err != nil {
		return nil, err
	}

	return &wrapper.Har, nil
}
