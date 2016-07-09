package runner

import (
	"github.com/JackDanger/traffic/model"
	"testing"
)

func TestPlay(t *testing.T) {
	// Given an Entry that contains a Request and a Response perform the request
	entry := model.Entry{
		Request: model.Request{},
	}
	runner := &Runner{
		Har: &model.Har{
			Entries: []model.Entry{entry},
		},
	}
	runner.Play(&entry)
}
