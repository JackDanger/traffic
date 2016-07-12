package runner

import (
	"encoding/json"
	"testing"

	"github.com/JackDanger/traffic/model"
	util "github.com/JackDanger/traffic/test"
)

type mockRequest struct {
	Verb        string
	URL         string
	Headers     []model.SingleItemMap
	QueryString []model.SingleItemMap
	Cookies     []model.SingleItemMap
}
type mockExecutor struct {
	Requests *[]mockRequest
	t        *testing.T
	Response model.Response
}

func (e mockExecutor) clone(verb string, original model.Request) {
	// Roundtrip the request through JSON to make a deep copy
	r := model.Request{}
	in, err := json.Marshal(original)
	if err != nil {
		e.t.Fatal(err)
	}
	err = json.Unmarshal(in, &r)
	if err != nil {
		e.t.Fatal(err)
	}

	*e.Requests = append(*e.Requests, mockRequest{
		Verb:        verb,
		URL:         r.URL,
		Headers:     r.Headers,
		QueryString: r.QueryString,
	})
}

func (e mockExecutor) Get(r model.Request) (model.Response, error) {
	e.clone("GET", r)
	return e.Response, nil
}
func (e mockExecutor) Post(r model.Request) (model.Response, error) {
	e.clone("POST", r)
	return e.Response, nil
}
func (e mockExecutor) Put(r model.Request) (model.Response, error) {
	e.clone("PUT", r)
	return e.Response, nil
}
func (e mockExecutor) Delete(r model.Request) (model.Response, error) {
	e.clone("DELETE", r)
	return e.Response, nil
}
func (e mockExecutor) Head(r model.Request) (model.Response, error) {
	e.clone("HEAD", r)
	return e.Response, nil
}
func (e mockExecutor) Patch(r model.Request) (model.Response, error) {
	e.clone("PATCH", r)
	return e.Response, nil
}

var _ Executor = mockExecutor{}

func TestPlay(t *testing.T) {

	testExecutor := mockExecutor{
		t:        t,
		Response: *util.MakeResponse(t),
		Requests: &[]mockRequest{},
	}

	entry := util.MakeEntry(t)
	runner := &Runner{
		Har: &model.Har{
			Entries: []model.Entry{*entry},
		},
		Executor: testExecutor,
	}
	runner.Play(entry)

	if len(*testExecutor.Requests) == 0 {
		t.Errorf("Expected number of performed requests to be higher than 0, was: %d", len(*testExecutor.Requests))
	}
}
