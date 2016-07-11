package transforms

import (
	"github.com/JackDanger/traffic/model"
	"regexp"
	"strings"
)

// RequestTransform encapsulates a single operation that should be performed on a
// request, a response, or a whole session. This allows a user to specify, for
// example, which part of the HTTP response should be reused in future requests
// after getting a response to a successful sign in.
//
// # Types of transforms:
//
// * A function that takes a request and replaces known CONSTANT by calling a
//   specific function provided by source code. Returns a function that
//   takes a response returns nothing. (i.e. it only modifies
//   requests outbound and does not re-enqueue any function to be run next
//   time)
// * A function that takes a request and returns a function that, given a
//   response, modifies that response somehow. It then returns a function
//   which is to be enqueued in the list of further transformations.
//
// So a transform is an object that will either run every time just as it is
// (if ResponseTransform returns nil) or it will replace
// itself by returning a new RequestTransform instance.
//
type RequestTransform interface {
	T(*model.Request) ResponseTransform
}

// ResponseTransform (returned from RequestTransform#T()) takes a response and may
// modify it or do nothing. Then it either returns a new RequestTransform that will
// replace the RequestTransform that called it on the next Request or it returns nil,
// signaling that the original RequestTransform should continue to run.
type ResponseTransform interface {
	T(*model.Response) RequestTransform
}

// Wraps a RequestTransform in a ResponseTransform that simply returns it. If
// no requestTransform is set then the T() method returns nil.
type passthrough struct {
	requestTransform RequestTransform
}

func (t passthrough) T(*model.Response) RequestTransform {
	return t.requestTransform
}

var noop ResponseTransform = &passthrough{}

// A little convenience object that wraps up the method you want to run later.
type responseProcessor struct {
	Tmethod func(*model.Response) RequestTransform
}

func (t responseProcessor) T(r *model.Response) RequestTransform {
	return t.Tmethod(r)
}

// ConstantTransform replaces known constants with function calls throughout a
// Request. It's useful, for example, to turn all instances of UNIXTIME into
// the string value (without quotes) of time.Now().Unix() or to replace GUID1,
// GUID2 with specific, predefined Guids that are constant across the session.
type ConstantTransform struct {
	Search  string
	Replace string
}

// T is because I don't know how to inherit from a func
func (t *ConstantTransform) T(r *model.Request) ResponseTransform {
	// We replace constants when they appear as string values anywhere in the
	// URL, in Headers (both keys and values) and in Cookies (both keys and
	// values)
	if strings.Contains(r.URL, t.Search) {
		r.URL = strings.Replace(r.URL, t.Search, t.Replace, -1)
	}

	for _, pairs := range [][]model.SingleItemMap{
		r.Headers,
		r.Cookies,
		r.QueryString,
	} {
		for _, pair := range pairs {
			if strings.Contains(*pair.Key, t.Search) {
				*pair.Key = strings.Replace(*pair.Key, t.Search, t.Replace, -1)
			}
			if strings.Contains(*pair.Value, t.Search) {
				*pair.Value = strings.Replace(*pair.Value, t.Search, t.Replace, -1)
			}
		}
	}

	return noop
}

// HeaderInjectionTransform is used to add a specific header to all requests.
type HeaderInjectionTransform struct {
	Key   string
	Value string
}

var _ RequestTransform = HeaderInjectionTransform{}

// T is because I don't know how to inherit from a func
func (t HeaderInjectionTransform) T(r *model.Request) ResponseTransform {
	r.Headers = append(r.Headers, model.SingleItemMap{
		Key:   &t.Key,
		Value: &t.Value,
	})
	return noop
}

// ResponseBodyToRequestHeaderTransform executes on every Request/Response
// loooking for a string in the response body that should be extracted and used
// in all subsequent request headers. Once the pattern is found this Transform
// replaces itself with a HeaderInjectionTransform that inserts a specific
// header into all subsequent requests.
type ResponseBodyToRequestHeaderTransform struct {
	Pattern    string // interpreted as a regular expression
	HeaderName string // which header to put the matched string into
}

// T is because I don't know how to inherit from a func
func (t ResponseBodyToRequestHeaderTransform) T(r *model.Request) ResponseTransform {
	regex := regexp.MustCompile(t.Pattern)

	// Find the string as a regular expression in the body somewhere and prepare
	// a HeaderInjectionTransform with it.
	matchString := func(r *model.Response) RequestTransform {
		if r.ContentBody == nil {
			return nil
		}

		var found string
		matches := regex.FindAllStringSubmatch(*r.ContentBody, -1)
		if len(matches) == 0 {
			return nil
		}
		firstMatch := matches[0]
		switch {
		case len(firstMatch) == 0:
			return nil
		case len(firstMatch) == 1:
			// There are no capture groups but the whole thing matched okay
			found = firstMatch[0]
		case len(firstMatch) > 1:
			// just use the first capture group, ignore the rest (TODO: disallow more than one)
			found = firstMatch[1]
		}
		if found == "" {
			return nil
		}

		return HeaderInjectionTransform{
			Key:   t.HeaderName,
			Value: found,
		}
	}

	// wrap this func in a responseProcessor just so it gets run not right now
	// during the request but later during the response.
	return responseProcessor{
		Tmethod: matchString,
	}
}
