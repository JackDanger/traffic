package runner

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/transforms"
	"github.com/JackDanger/traffic/util"
)

func TestPlay(t *testing.T) {

	executor := testExecutor(t)
	entry := util.MakeEntry()
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

func TestRunWholeTar(t *testing.T) {

	har := util.Fixture()
	entryCount := len(har.Entries)
	if len(har.Entries) <= 1 {
		t.Fatalf("Expected there to be at least 2 entries, found %d", len(har.Entries))
	}

	executor := testExecutor(t)
	instance := Run(&har, executor, nil)

	select {
	case <-instance.DoneChannel:
		// wait for instance to send something on it's completion channel
	}
	if len(*executor.ProcessedRequests) < entryCount {
		t.Errorf("expected to process all %d entries, only processed %d", entryCount, len(*executor.ProcessedRequests))
	}
}

func TestRunWithComplexTranforms(t *testing.T) {
	har := util.Fixture()
	entryCount := len(har.Entries)
	if len(har.Entries) <= 1 {
		t.Fatalf("Expected there to be at least 2 entries, found %d", len(har.Entries))
	}

	executor := testExecutor(t)

	ts := []transforms.RequestTransform{}
	ts = append(ts, &transforms.ConstantTransform{
		Search:  "heddle317", // found in entry[2]
		Replace: "SingingParodies",
	})
	ts = append(ts, &transforms.ConstantTransform{
		Search:  "nehakarajgikar", // found in entry[3]
		Replace: "SugarInMyDrinks",
	})
	ts = append(ts, &transforms.BodyToHeaderTransform{
		Pattern:    `"session": (\d+)`,
		HeaderName: "SessionID",
		Before:     "prefix-",
		After:      "-suffix",
	})
	executor.Response.ContentBody = util.StringPtr(`{
		"session": 42342323,
		"other": "stuff"
	}`)
	//ts = append(ts, &transforms.HeaderToHeaderTransform{
	//	Pattern:    "github.v(\\d+)",
	//	HeaderName: "Github-VERSION",
	//	Before:     "omgifoundit-",
	//	After:      "-totes",
	//})

	instance := Run(&har, executor, ts)

	select {
	case <-instance.DoneChannel:
		// wait for instance to send something on it's completion channel
	}
	if len(*executor.ProcessedRequests) < entryCount {
		t.Errorf("expected to process all %d entries, only processed %d", entryCount, len(*executor.ProcessedRequests))
	}

	requests := *executor.ProcessedRequests

	// Expect the 3rd entry to have heddle317 swapped out
	if !strings.Contains(requests[2].URL, "SingingParodies") {
		t.Errorf("expected heddle317 to be replaced")
	}
	// Expect the 4th entry to have nehakarajgikar swapped out
	if !strings.Contains(requests[3].URL, "SugarInMyDrinks") {
		t.Errorf("expected nehakarajgikar to be replaced")
	}
	// Expect the first request doesn't have any session header
	if util.Any(requests[0].Headers, func(key, _ *string) bool {
		return *key == "SessionID"
	}) {
		t.Error("Github version header was in first request")
	}
	for _, request := range []mockRequest{
		requests[1],
		requests[2],
		requests[3],
	} {
		// Expect the subsequent requests to ALL have the header set
		if !util.Any(request.Headers, func(key, value *string) bool {
			return *key == "SessionID" && *value == "prefix-42342323-suffix"
		}) {
			t.Errorf("Github version header was not found in request! %#v", request)
		}
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
		Response:          *util.MakeResponse(),
		ProcessedRequests: &[]mockRequest{},
	}
}

// Given a request, copy it to a local list of processed reqeusts.
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

func (e mockExecutor) Get(r model.Request) (*model.Response, error) {
	e.clone("GET", r)
	return &e.Response, nil
}
func (e mockExecutor) Post(r model.Request) (*model.Response, error) {
	e.clone("POST", r)
	return &e.Response, nil
}
func (e mockExecutor) Put(r model.Request) (*model.Response, error) {
	e.clone("PUT", r)
	return &e.Response, nil
}
func (e mockExecutor) Delete(r model.Request) (*model.Response, error) {
	e.clone("DELETE", r)
	return &e.Response, nil
}
func (e mockExecutor) Head(r model.Request) (*model.Response, error) {
	e.clone("HEAD", r)
	return &e.Response, nil
}
func (e mockExecutor) Patch(r model.Request) (*model.Response, error) {
	e.clone("PATCH", r)
	return &e.Response, nil
}

var _ Executor = mockExecutor{}
