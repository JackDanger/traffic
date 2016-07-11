package model

// Har represents the top-level, single  `log` key of the .har file
type Har struct {
	Version string   `json:"version"`
	Creator creator  `json:"creator"`
	Pages   []string `json:"pages"` // Not impleneted
	Entries []Entry  `json:"entries"`
}

// Entry is a single request & response
type Entry struct {
	Start    string   `json:"startedDateTime"`
	TimeMs   uint32   `json:"time"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Cache    cache    `json:"cache"`
	Timings  timings  `json:"timings"`
}

// Request represents a single HTTP request
type Request struct {
	Method      string          `json:"method"`
	URL         string          `json:"url"`
	HTTPVersion string          `json:"httpVersion"`
	Headers     []SingleItemMap `json:"headers"`
	QueryString []SingleItemMap `json:"queryString"`
	Cookies     []SingleItemMap `json:"cookies"`
	HeaderSize  uint32          `json:"headersSize"`
	BodySize    uint32          `json:"bodySize"`
}

// Response represents a single HTTP response
type Response struct {
	Status      uint32          `json:"status"`
	StatusText  string          `json:"statusText"`
	HTTPVersion string          `json:"httpVersion"`
	Headers     []SingleItemMap `json:"headers"`
	Cookies     []SingleItemMap `json:"cookies"`
	Content     content         `json:"content"`
	RedirectUTL string          `json:"redirectURL"`
	HeadersSize uint32          `json:"headersSize"`
	BodySize    uint32          `json:"bodySize"`
}

// SingleItemMap is a single key-value pair because that's how HAR represents
// headers and cookies. The headers are a list of single-element maps, not a
// single unified map.
type SingleItemMap struct {
	Key   string `json:"name"`
	Value string `json:"value"`
}

type creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type content struct {
	Size        uint64 `json:"size"`
	MimeType    string `json:"mimeType"`
	Compression uint32 `json:"compression"`
}

type cache struct{}

type timings struct {
	Blocked int32 `json:"blocked"`
	DNS     int32 `json:"dns"`
	Connect int32 `json:"connect"`
	Send    int32 `json:"send"`
	Wait    int32 `json:"wait"`
	Receive int32 `json:"receive"`
	SSL     int32 `json:"ssl"`
}
