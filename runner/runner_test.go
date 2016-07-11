package runner

import (
	"testing"

	"github.com/JackDanger/traffic/model"
	util "github.com/JackDanger/traffic/test"
)

func TestPlay(t *testing.T) {
	// Given an Entry that contains a Request and a Response perform the request
	entry := util.MakeEntry(t)
	runner := &Runner{
		Har: &model.Har{
			Entries: []model.Entry{*entry},
		},
	}
	runner.Play(entry)
}
