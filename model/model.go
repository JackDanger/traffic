package model

import (
	"encoding/json"
)

// Har represents the top-level, single  `log` key of the .har file
type Har struct {
	Name    string   // Internal use only, not part of the HAR file
	Version string   `json:"version"`
	Creator creator  `json:"creator"`
	Pages   []string `json:"pages"` // Not impleneted
	Entries []Entry  `json:"entries"`
}

// Entry is a single request & response
type Entry struct {
	Start           string    `json:"startedDateTime"`
	TimeMs          float64   `json:"time"`
	Request         *Request  `json:"request"`
	Response        *Response `json:"response"`
	Cache           cache     `json:"cache"`
	Timings         timings   `json:"timings"`
	ServerIPAddress string    `json:"serverIPAddress,omitempty"`
	Pageref         string    `json:"pageref,omitempty"`
}

// Request represents a single HTTP request
type Request struct {
	Method      string          `json:"method"`
	URL         string          `json:"url"`
	HTTPVersion string          `json:"httpVersion"`
	Headers     []SingleItemMap `json:"headers"`
	QueryString []SingleItemMap `json:"queryString"`
	Cookies     []Cookie        `json:"cookies"`
	HeaderSize  int             `json:"headersSize"`
	BodySize    int             `json:"bodySize"`
	PostData    *PostData       `json:"postData,omitempty"`
}

// Response represents a single HTTP response
type Response struct {
	Status       int             `json:"status"`
	StatusText   string          `json:"statusText"`
	HTTPVersion  string          `json:"httpVersion"`
	Headers      []SingleItemMap `json:"headers"`
	Cookies      []SingleItemMap `json:"cookies"` // Chrome produces HARs with this but response cookies make no sense
	Content      content         `json:"content"`
	ContentBody  *string         `json:"body,omitempty"` // ContentBody is not present in HAR files
	RedirectURL  string          `json:"redirectURL"`
	HeadersSize  int             `json:"headersSize"`
	BodySize     int             `json:"bodySize"`
	TransferSize *int            `json:"_transferSize,omitempty"`
}

// SingleItemMap is a single key-value pair because that's how HAR represents
// headers and cookies. The headers are a list of single-element maps, not a
// single unified map.
type SingleItemMap struct {
	Key   *string `json:"name"`
	Value *string `json:"value"`
}

// Cookie is a slightly more complex SingleItemMap
type Cookie struct {
	SingleItemMap
	Expires  nullString `json:"expires"`
	HTTPOnly bool       `json:"httpOnly"`
	Secure   bool       `json:"secure"`
}

type creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type content struct {
	Size        int    `json:"size"`
	MimeType    string `json:"mimeType"`
	Compression int    `json:"compression,omitempty"`
}

// PostData represents the content type and then two ways to look at the data that's submitted with a POST request
type PostData struct {
	MimeType string          `json:"mimeType,omitempty"`
	Text     string          `json:"text,omitempty"`
	Params   []SingleItemMap `json:"params,omitempty"`
}

type cache struct{}

type timings struct {
	Blocked float64 `json:"blocked"`
	DNS     float64 `json:"dns"`
	Connect float64 `json:"connect"`
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
	SSL     float64 `json:"ssl"`
}

// This is a string that is represented as "null" in JSON
type nullString string

func (n *nullString) MarshalJSON() ([]byte, error) {
	if string(*n) == "" {
		return []byte("null"), nil
	}
	return json.Marshal(*n)
}
func (n *nullString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	return json.Unmarshal(b, (*string)(n))
}
