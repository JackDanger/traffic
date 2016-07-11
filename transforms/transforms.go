package transforms

import (
	"github.com/JackDanger/traffic/model"
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
