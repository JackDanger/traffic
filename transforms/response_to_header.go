package transforms

import (
	"github.com/JackDanger/traffic/model"
	"regexp"
)

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
	// If no matches are found we `return t` so this same code is run again and
	// again until an appropriate match _is_ found.
	matchString := func(r *model.Response) RequestTransform {
		if r.ContentBody == nil {
			return t
		}

		var found string
		matches := regex.FindAllStringSubmatch(*r.ContentBody, -1)
		if len(matches) == 0 {
			return t
		}
		firstMatch := matches[0]
		switch {
		case len(firstMatch) == 0:
			return t
		case len(firstMatch) == 1:
			// There are no capture groups but the whole thing matched okay
			found = firstMatch[0]
		case len(firstMatch) > 1:
			// just use the first capture group, ignore the rest (TODO: disallow more than one)
			found = firstMatch[1]
		}
		if found == "" {
			return t
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
