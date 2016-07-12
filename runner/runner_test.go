package runner

import (
	"encoding/json"
	"testing"

	"github.com/JackDanger/traffic/model"
	util "github.com/JackDanger/traffic/test"
)

func TestPlay(t *testing.T) {

	executor := testExecutor(t)
	entry := util.MakeEntry(t)
	runner := &Runner{
		Har: &model.Har{
			Entries: []model.Entry{*entry},
		},
		Executor: executor,
	}
	runner.Play(entry)

	if len(*executor.ProcessedRequests) == 0 {
		t.Errorf("Expected number of performed requests to be higher than 0, was: %d", len(*executor.ProcessedRequests))
	}
}

func TestPlayWholeHar(t *testing.T) {

	har := util.Fixture(t)
	entryCount := len(har.Entries)
	if len(har.Entries) <= 1 {
		t.Fatalf("Expected there to be at least 2 entries, found %d", len(har.Entries))
	}

	executor := testExecutor(t)
	instance := Run(&har, executor)

	select {
	case <-instance.doneChannel:
		// wait for instance to send something on it's completion channel
	}
	if len(*executor.ProcessedRequests) < entryCount {
		t.Errorf("expected to process all %d entries, only processed %d", entryCount, len(*executor.ProcessedRequests))
	}
}

func TestPlayWholeHarWithComplexTranforms(t *testing.T) {
	har := util.Fixture(t)
	entryCount := len(har.Entries)
	if len(har.Entries) <= 1 {
		t.Fatalf("Expected there to be at least 2 entries, found %d", len(har.Entries))
	}

	executor := testExecutor(t)
	instance := Run(&har, executor)

	select {
	case <-instance.doneChannel:
		// wait for instance to send something on it's completion channel
	}
	if len(*executor.ProcessedRequests) < entryCount {
		t.Errorf("expected to process all %d entries, only processed %d", entryCount, len(*executor.ProcessedRequests))
	}
}

// TODO: Test all of
// * pausing & continuing
// * stopping and trying to continue
// * repeatedly pausing/continuing
// * two identical runners at once don't clobber each other's data (esp. Entry contents)

// TODO: implement all of
// * time-shifting (makes tests faster!)

// Helpers

type mockRequest struct {
	Verb        string
	URL         string
	Headers     []model.SingleItemMap
	QueryString []model.SingleItemMap
	Cookies     []model.SingleItemMap
}
type mockExecutor struct {
	ProcessedRequests *[]mockRequest
	t                 *testing.T
	Response          model.Response
}

func testExecutor(t *testing.T) mockExecutor {
	return mockExecutor{
		t:                 t,
		Response:          *util.MakeResponse(t),
		ProcessedRequests: &[]mockRequest{},
	}
}

// Given a request copy it to a local list of processed reqeusts.
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

	*e.ProcessedRequests = append(*e.ProcessedRequests, mockRequest{
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
