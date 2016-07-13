package transforms

import (
	"github.com/JackDanger/traffic/model"
	"strings"
)

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
			if strings.Contains(*pair.Key, t.Search) {
				*pair.Key = strings.Replace(*pair.Key, t.Search, t.Replace, -1)
			}
			if strings.Contains(*pair.Value, t.Search) {
				*pair.Value = strings.Replace(*pair.Value, t.Search, t.Replace, -1)
			}
		}
	}

	return passthrough{requestTransform: t}

}
