package parser

import (
	"strings"
)

// UnquoteJSON removes the extra level of escaping that encoding/json adds via
// json.Marshal
func UnquoteJSON(s string) string {
	return strings.NewReplacer(
		"\\u003c", "<",
		"\\u003e", ">",
		"\\u0026", "&",
	).Replace(s)
}
