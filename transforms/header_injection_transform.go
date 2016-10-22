package transforms

import (
	"github.com/JackDanger/traffic/model"
)

// HeaderInjectionTransform is used to add a specific header to all requests.
type HeaderInjectionTransform struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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
