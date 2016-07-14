package runner

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/JackDanger/traffic/model"
)

// Executor is anything that can perform HTTP requests (it's an interface so we
// can mock it in tests)
type Executor interface {
	Get(model.Request) (*model.Response, error)
	Post(model.Request) (*model.Response, error)
	Put(model.Request) (*model.Response, error)
	Delete(model.Request) (*model.Response, error)
	Head(model.Request) (*model.Response, error)
	Patch(model.Request) (*model.Response, error)
}

// HTTPExecutor has methods that accept a request from a HAR file and make a
// connection to the actual URL specified in each request. The HTTP response is
// then placed into a model.Response object and returned.
type HTTPExecutor struct {
	client      http.Client
	logger      Logger
	lastRequest *http.Request // only for testing
}

// Get performs an HTTP POST
func (e *HTTPExecutor) Get(r model.Request) (*model.Response, error) {
	req, err := http.NewRequest("GET", r.URL, nil)
	fmt.Println(req.URL.Host)
	if err != nil {
		e.handleError(err)
		return &model.Response{}, err
	}

	e.fromModelRequest(req, &r)

	return e.toModelResponse(e.client.Do(req)), nil
}

// Post performs an HTTP POST
func (e *HTTPExecutor) Post(r model.Request) (*model.Response, error) {
	req, err := http.NewRequest("POST", r.URL, bytes.NewBufferString(r.PostData.Text))
	if err != nil {
		e.handleError(err)
		return &model.Response{}, err
	}

	e.fromModelRequest(req, &r)

	return e.toModelResponse(e.client.Do(req)), nil
}

// Put performs an HTTP PUT
func (e *HTTPExecutor) Put(r model.Request) (*model.Response, error) {
	return &model.Response{}, nil
}

// Delete performs an HTTP DELETE
func (e *HTTPExecutor) Delete(r model.Request) (*model.Response, error) {
	return &model.Response{}, nil
}

// Head is the equivalent of `curl -I`
func (e *HTTPExecutor) Head(r model.Request) (*model.Response, error) {
	return &model.Response{}, nil
}

// Patch is an HTTP verb we'll rarely see
func (e *HTTPExecutor) Patch(r model.Request) (*model.Response, error) {
	return &model.Response{}, nil
}

// NewHTTPExecutor returns an object that can perform live HTTP requests
func NewHTTPExecutor(logDevice io.Writer) Executor {
	return &HTTPExecutor{
		client: http.Client{},
		logger: NewLogger(logDevice),
	}
}

func (e *HTTPExecutor) toModelResponse(h *http.Response, err error) *model.Response {
	if err != nil {
		e.handleError(err)
		return &model.Response{}
	}

	headers := []model.SingleItemMap{}
	for key, values := range h.Header {
		// HTTP allows you to send the same header key multiple times and
		// http.Header stores a collapsed version of all headers with the same key.
		// Here we explode them back out.
		for _, value := range values {
			headers = append(headers, model.SingleItemMap{
				// Make a copy of the string otherwise we store the address and this
				// loop reuses the address each time through.
				Key:   func(s string) *string { return &s }(key),
				Value: &value,
			})
		}
	}

	body, err := ioutil.ReadAll(h.Body)
	if err != nil {
		e.logger.Println("error reading http response body: ", err)
	}

	e.log(h.Status)

	return &model.Response{
		HTTPVersion: h.Proto,
		Status:      h.StatusCode,
		StatusText:  h.Status,
		Headers:     headers,
		ContentBody: func(s string) *string { return &s }(string(body)),
	}
}

func (e *HTTPExecutor) handleError(err error) {
	e.log("error: ", fmt.Sprintf("%#v", err))
}

// Reads the headers from the model.Request and applies them to the
// http.Request. Adds a Content-Type header if one is missing.
func (e *HTTPExecutor) fromModelRequest(req *http.Request, modelRequest *model.Request) {
	e.log(req.Method, ": ", req.URL)
	e.lastRequest = req

	contentTypeIsSet := false
	for _, header := range modelRequest.Headers {
		if *header.Key == "Content-Type" {
			contentTypeIsSet = true
		}
		req.Header.Set(*header.Key, *header.Value)
	}
	if contentTypeIsSet {
		return
	}

	// We must have manually provided a Content-Type header, don't try to intuit
	// one from the PostData
	if modelRequest.PostData != nil {
		req.Header.Set("Content-Type", modelRequest.PostData.MimeType)
	} else {
		req.Header.Set("Content-Type", "text/html") // can't think of a better default
	}
}

// Logger encapsulates printing to the screen or a file or a variable under test.
type Logger struct {
	device io.Writer
}

// Println sends to the logging device what fmt.Println sends to os.Stdout
func (l *Logger) Println(s ...interface{}) {
	l.device.Write([]byte(fmt.Sprintln(s...)))
}

func (e *HTTPExecutor) log(s ...interface{}) {
	e.logger.Println(s...)
}

// GetLastRequest is used in testing to assert we've properly transformed inbound values
func (e *HTTPExecutor) GetLastRequest() *http.Request {
	return e.lastRequest
}

// NewLogger produces a logger backed by the provided device
func NewLogger(device io.Writer) Logger {
	return Logger{device: device}
}
