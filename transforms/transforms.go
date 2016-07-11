package transforms

import (
	"github.com/JackDanger/traffic/model"
	"strings"
)

// Transform encapsulates a single operation that should be performed on a
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
// itself by returning a new Transform instance.
//
type Transform interface {
	T(*model.Request) ResponseTransform
}

// ResponseTransform (returned from Transform.T()) takes a response and may
// modify it or do nothing. Then it either returns a new Transform that will
// replace the Transform that called it on the next Request or it returns nil,
// signaling that the original Transform should continue to run.
type ResponseTransform interface {
	T(*model.Response) *Transform
}

// ConstantTransform replaces known constants with function calls throughout a
// Request. It's useful, for example, to turn all instances of UNIXTIME into
// the string value (without quotes) of time.Now().Unix() or to replace GUID1,
// GUID2 with specific, predefined Guids that are constant across the session.
type ConstantTransform struct {
	Search  string
	Replace string
}

// If a Transform has no intention of being replaced by the return value of its
// ResponseTransform it can return this.
type noopResponseTransform struct{}

func (t *noopResponseTransform) T(*model.Response) *Transform {
	return nil
}

var noop = &noopResponseTransform{}

// T exists because I don't know of a way to make Transform and
// ResponseTransform subclassed from functions directly.
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
