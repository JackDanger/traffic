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
// * A function that takes a request (and may modify it) and returns a function
//   that takes a response (which it may modify) that then returns the original
//   function.
//
//   Example:
//     responseTransform := requestTransform.T(aRequest)
//     if requestTransform != responseTransform.T(aResponse) {
//       panic("no, seriously, the original transform should be returned")
//     }
//
// * A function that takes a request (and may modify it) and returns a function
//   that takes a response (which it may modify) that then returns the a new
//   function. This will replace the original one and the next request/response
//   cycle there will be this new behavior.
//
//   Example:
//     responseTransform := requestTransform.T(aRequest)
//     if requestTransform == responseTransform.T(aResponse) {
//       panic("In this case we expected there to be a new transform produced")
//     }
//
// These transforms are implemented  as objects with a T() method that will
// either run every time just as they are (if the response transform returns
// the original request transform) or they will generate a new transform to
// replace themselves.
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

// Wraps a RequestTransform in a ResponseTransform that simply returns it.s
type passthrough struct {
	requestTransform RequestTransform
}

func (t passthrough) T(*model.Response) RequestTransform {
	return t.requestTransform
}

// A little convenience object that wraps up a function defined in the
// RequestTransform to be executed in the ResponseTransform
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
	// Return a transform that just returns this current one.
	return passthrough{requestTransform: t}
}
