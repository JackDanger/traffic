package transforms

import (
	"fmt"
	"regexp"

	"github.com/JackDanger/traffic/model"
)

// HeaderToHeaderTransform executes on every Request/Response
// loooking for a string in one of the response headers that should be
// extracted and used in all subsequent request headers. Once the pattern is
// found this Transform replaces itself with a HeaderInjectionTransform that
// inserts a specific header into all subsequent requests.
// The reason these are different objects is that the HeaderToHeaderTransform
// only observes requests and modifies nothing. When it encounters a matching
// value then it replaces itself with a transform that modifies requests.
//
// Example:
//
//   Given a response that contains the headers:
//     {
//       "new-session":    "Session ABC123",
//       "Account-number": "4001",
//     }
//
//   And a transform defined as:
//     HeaderToHeaderTransform{
//       ResponseKey: "new-session",
//       Pattern:     "Session (.+)",
//       RequestKey:  "X-AUTHORIZATION",
//       Before:      "api:(",
//       After:        ")",
//     }
//
//   Future requests will be made with the header:
//     {
//       "X-AUTHORIZATION: "api:(ABC123)",
//     }
//
type HeaderToHeaderTransform struct {
	Type        string `json:"type"`
	ResponseKey string `json:"response_key"` // which header to read the value out of. If blank, all headers will be checked for a matching pattern.
	Pattern     string `json:"pattern"`      // interpreted as a regular expression
	RequestKey  string `json:"request_key"`  // which header to put the matched string into
	Before      string `json:"before"`       // What to put into the header value before the match
	After       string `json:"after"`        // What to put into the header value after the match
}

// T is because I don't know how to inherit from a func
func (t HeaderToHeaderTransform) T(r *model.Request) ResponseTransform {
	// Find the string as a regular expression in the body somewhere and prepare
	// a HeaderInjectionTransform with it.
	// If no matches are found we `return t` so this same code is run again and
	// again until an appropriate match _is_ found.
	matchString := func(r *model.Response) RequestTransform {
		for _, header := range r.Headers {
			replacementTransform := t.maybeRelace(&header)
			if replacementTransform != nil {
				return replacementTransform
			}
		}
		return t
	}

	// wrap this func in a responseProcessor just so it gets run not right now
	// during the request but later during the response.
	return responseProcessor{
		Tmethod: matchString,
	}
}

func (t HeaderToHeaderTransform) maybeRelace(header *model.SingleItemMap) *HeaderInjectionTransform {
	regex := regexp.MustCompile(t.Pattern)

	var found string

	// If t.ResponseKey is nil then we try to match Pattern against any header header
	if t.ResponseKey != "" && *header.Key != t.ResponseKey {
		fmt.Println("transforms/header_to_header.go:76 ", t.ResponseKey, "didn't match ", *header.Key)
		return nil
	}

	// extract the replacement string via regex
	matches := regex.FindAllStringSubmatch(*header.Value, -1)
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
		// no match, return the original transform and try again
		return nil
	}
	found = t.Before + found + t.After
	return &HeaderInjectionTransform{
		Key:   t.RequestKey,
		Value: found,
	}
}
