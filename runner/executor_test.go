package runner

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/util"
)

type handler struct{}

var responseHeaders = map[string]string{}
var responseBody = "default response body"
var performedRequest *http.Request
var performedRequestBody string

// This boots a test server that listens on the local host and response to any
// test client requests with headers and content stored in global variables.
func (h *handler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	// Store the incoming request so we can check it
	performedRequest = request
	// And read the request body so we can store that
	body, _ := ioutil.ReadAll(request.Body)
	performedRequestBody = string(body)

	// set any headers defined in this test
	for key, value := range responseHeaders {
		w.Header().Set(key, value)
	}
	// set the body defined in this test
	w.Write([]byte(responseBody))
}

// This starts a server and immediately backgrounds it via a goroutine
func startServer() string {
	server := http.Server{
		Addr:    "127.0.0.1:9797",
		Handler: &handler{},
	}
	go server.ListenAndServe()
	// Wait a moment so the server can boot
	time.Sleep(100 * time.Millisecond)
	return "server is running"
}

var started = startServer()

func TestGet(t *testing.T) {
	println(started)
	executor := NewHTTPExecutor(os.Stdout)

	// Set some headers that our server will sent back to us - we'll check that
	// they return in the right format
	responseHeaders["Time"] = "noonish"
	responseHeaders["Session"] = "valid-session"
	// And define a body we hope to retrieve
	responseBody = "This should be the body"

	request := model.Request{
		URL: "http://localhost:9797/",
		Headers: []model.SingleItemMap{
			{
				Key:   util.StringPtr("X-REQUEST-HEADER"), // NB: net/http will modify the case of this string
				Value: util.StringPtr("Val"),
			},
		},
	}
	response, err := executor.Get(request)
	if err != nil {
		t.Fatal(err)
	}

	if !hasHeader(stdLibHeadersToModel(performedRequest.Header), "X-Request-Header", "Val") {
		t.Errorf("Expected our headers to make it to the server, got: \n%s", printSingleItemMaps(stdLibHeadersToModel(performedRequest.Header)))
	}
	if !hasHeader(response.Headers, "Time", "noonish") {
		t.Errorf("Couldn't find the first test header in\n %s", printSingleItemMaps(response.Headers))
	}
	if !hasHeader(response.Headers, "Session", "valid-session") {
		t.Errorf("Couldn't find the first second header in\n %s", printSingleItemMaps(response.Headers))
	}
	if *response.ContentBody != "This should be the body" {
		t.Errorf("Unexpected body content: %s", response.ContentBody)
	}
}

func TestGetWithElaborateHeadersAndQueryString(t *testing.T) {
	println(started)
	executor := NewHTTPExecutor(os.Stdout).(*HTTPExecutor)

	request := model.Request{
		URL: "http://localhost:9797/kate/heddleston?query=string",
		QueryString: []model.SingleItemMap{
			{
				Key:   util.StringPtr("Second"),
				Value: util.StringPtr("QueryValue"),
			},
		},
		Headers: []model.SingleItemMap{
			{
				Key:   util.StringPtr("Official-Title"),
				Value: util.StringPtr("Software Warrior Princess/CEO"),
			},
			{
				Key:   util.StringPtr("Content-Type"),
				Value: util.StringPtr("application/x-set-in-header-manually"),
			},
		},
	}
	_, err := executor.Get(request)
	if err != nil {
		t.Fatal(err)
	}
	req := executor.GetLastRequest()

	if !hasHeader(stdLibHeadersToModel(req.Header), "Official-Title", "Software Warrior Princess/CEO") {
		t.Errorf("Expected our headers to make it to the server, got: \n%s", printSingleItemMaps(stdLibHeadersToModel(req.Header)))
	}
	if !hasHeader(stdLibHeadersToModel(req.Header), "Content-Type", "application/x-set-in-header-manually") {
		t.Errorf("Expected our headers to make it to the server, got: \n%s", printSingleItemMaps(stdLibHeadersToModel(req.Header)))
	}
	if req.URL.Host != "localhost:9797" {
		t.Errorf("host was not properly incorporated into the URL: %s, %s", req.URL.Host, req.URL)
	}
	if req.URL.Path != "/kate/heddleston" {
		t.Errorf("path was not properly incorporated into the URL: %s, %s", req.URL.Path, req.URL)
	}
	if req.URL.RawQuery != "query=string" {
		t.Errorf("query string was not properly incorporated into the URL: %s, %s", req.URL.RawQuery, req.URL)
	}
}

func TestPost(t *testing.T) {
	println(started)
	executor := NewHTTPExecutor(os.Stdout)

	// Set some headers that our server will sent back to us - we'll check that
	// they return in the right format
	responseHeaders["Time"] = "afternoon"
	responseHeaders["Content-Type"] = "application/x-protobuf"
	// And define a body we hope to retrieve
	responseBody = "This should be the body"

	response, err := executor.Post(model.Request{
		URL: "http://localhost:9797/",
		PostData: &model.PostData{
			MimeType: "application/json",
			Text:     `{"Nickname": "Jenny", "Role": "Team Captain"}`,
		},
	})

	if response.Status != 200 {
		t.Errorf("Expected HTTP 200 OK, got: %#v", response.Status)
	}

	if performedRequestBody != `{"Nickname": "Jenny", "Role": "Team Captain"}` {
		t.Errorf("Expected a particular request body, got: %s ", performedRequestBody)
	}

	if err != nil {
		t.Fatal(err)
	}

	if !hasHeader(response.Headers, "Time", "afternoon") {
		t.Errorf("Couldn't find the first test header in\n %s", printSingleItemMaps(response.Headers))
	}

	if !hasHeader(response.Headers, "Content-Type", "application/x-protobuf") {
		t.Errorf("Couldn't find the second test header in\n %s", printSingleItemMaps(response.Headers))
	}

	if response.ContentBody != nil && *response.ContentBody != "This should be the body" {
		t.Errorf("Unexpected body content: %s", response.ContentBody)
	}
}

// Test Helperrs

func stdLibHeadersToModel(header http.Header) []model.SingleItemMap {
	var m []model.SingleItemMap
	for key, values := range header {
		for _, value := range values {
			m = append(m, model.SingleItemMap{
				Key:   util.StringPtr(key),
				Value: util.StringPtr(value),
			})
		}
	}
	return m
}

func hasHeader(headers []model.SingleItemMap, findKey, findValue string) bool {
	return util.Any(headers, func(key, value *string) bool {
		return *key == findKey && *value == findValue
	})
}

func printSingleItemMaps(headers []model.SingleItemMap) string {
	reorganized := map[string]string{}

	for _, h := range headers {
		reorganized[*h.Key] = *h.Value
	}
	str, _ := json.Marshal(reorganized)
	return string(str)
}
