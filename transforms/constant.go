package transforms

import (
	"regexp"

	"github.com/JackDanger/traffic/model"
)

// ConstantTransform replaces known constants with function calls throughout a
// Request. It's useful, for example, to turn all instances of UNIXTIME into
// the string value (without quotes) of time.Now().Unix() or to replace GUID1,
// GUID2 with specific, predefined Guids that are constant across the session.
type ConstantTransform struct {
	Search         string `json:"search"`
	Replace        string `json:"replace"`
	compiledSearch *regexp.Regexp
}

// T is because I don't know how to inherit from a func
func (t *ConstantTransform) T(r *model.Request) ResponseTransform {
	// We replace constants when they appear as string values anywhere in the
	// URL, in Headers (both keys and values) and in Cookies (both keys and
	// values)
	t.replace(&r.URL)

	// Extract the key/value from the cookies.
	var cookieMaps []model.SingleItemMap
	for _, cookie := range r.Cookies {
		cookieMaps = append(cookieMaps, cookie.SingleItemMap)
	}
	for _, pairs := range [][]model.SingleItemMap{
		r.Headers,
		cookieMaps,
		r.QueryString,
	} {
		for _, pair := range pairs {
			t.replace(pair.Key)
			t.replace(pair.Value)
		}
	}

	// Don't do anything with the response and reuse this same transformation on
	// the next request.
	return passthrough{requestTransform: t}
}

// Mutate a string to replace any instances of t.Search with t.Replace. Handles
// both static strings and regular expressions.
func (t *ConstantTransform) replace(content *string) {
	if t.compiledSearch == nil {
		t.compiledSearch = regexp.MustCompile(t.Search)
	}
	*content = t.compiledSearch.ReplaceAllString(*content, t.Replace)
}
