package runner

import (
	"errors"
	"sync"
	"time"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/transforms"
)

// Operation is used to send messages about stopping and starting to the Runner
// goroutine that's running the Har in a loop.
type Operation uint16

const (
	// Pause tells the Runner goroutine to sleep until another operation is sent
	Pause Operation = 0
	// Stop will halt the goroutine and remove it from the running tasks
	Stop
	// Continue will un-pause a goroutine. Has no effect on non-paused goroutines.
	Continue
)

// Executor performs the actual HTTP actions (mocked in tests)
type Executor interface {
	Get(model.Request) (model.Response, error)
	Post(model.Request) (model.Response, error)
	Put(model.Request) (model.Response, error)
	Delete(model.Request) (model.Response, error)
	Head(model.Request) (model.Response, error)
	Patch(model.Request) (model.Response, error)
}

// Runner encapsulates a single goroutine reading and replaying a HAR
type Runner struct {
	m                  sync.Mutex
	Har                *model.Har
	Running            bool
	StartTime          time.Time
	operationChannel   chan Operation
	requestChannel     chan *model.Entry
	requestTransforms  []transforms.RequestTransform
	responseTransforms []transforms.ResponseTransform
	Executor           Executor
}

// This is the list (implemented as a map so we can use instance pointers) of
// currently-running Runner instances.
type runnerList struct {
	items map[*Runner]bool
	m     sync.Mutex
}

var runners = &runnerList{
	items: map[*Runner]bool{},
	m:     sync.Mutex{},
}

// Run accepts a full HAR and begins to replay the contents at the
// originally-recorded timing intervals.
func Run(har *model.Har) *Runner {
	replay := &Runner{
		operationChannel: make(chan Operation),
		StartTime:        time.Now(),
		Har:              har,
		Running:          false,
	}

	replay.begin()

	return replay
}

func (r *Runner) begin() error {
	runners.m.Lock()
	defer runners.m.Unlock()
	if runners.items[r] {
		return errors.New("Attempting to start the same Runner twice, maybe you meant to Continue() it?")
	}
	// Add this runner to the list of runners
	runners.items[r] = true
	r.Running = true

	go func() {
		select {
		// Check if we've been asked to pause or continue or shut down
		case operation := <-r.operationChannel:
			if operation == Pause {
				r.m.Lock()
				r.Running = false
				r.m.Unlock()
			} else if operation == Continue {
				r.m.Lock()
				r.Running = true
				r.m.Unlock()
			} else if operation == Stop {
				r.m.Lock()
				defer r.m.Unlock()

				r.Running = false

				runners.m.Lock()
				defer runners.m.Unlock()
				defer delete(runners.items, r) // Remove this instance from the list
				return                         // This is where we shut the whole routine down
			}
		// Check if there's another request to make. If so, play it (play() spawns
		// a goroutine and returns immediately)
		case entry := <-r.requestChannel:
			r.play(entry)
		}

	}()
	return nil
}

func (r *Runner) play(entry *model.Entry) {
	go func() {
		r.Play(entry)
	}()
}

// Play performs the request described in the Entry
func (r *Runner) Play(entry *model.Entry) error {
	transformedRequest := r.transformRequest(entry.Request)

	var err error
	var response model.Response

	switch entry.Request.Method {
	case "GET":
		response, err = r.Executor.Get(*transformedRequest)
	case "POST":
		response, err = r.Executor.Post(*transformedRequest)
	case "PUT":
		response, err = r.Executor.Put(*transformedRequest)
	case "DELETE":
		response, err = r.Executor.Delete(*transformedRequest)
	case "HEAD":
		response, err = r.Executor.Head(*transformedRequest)
	case "PATCH":
		response, err = r.Executor.Patch(*transformedRequest)
	}

	r.updateTransformsFromResponse(&response)

	if &response == nil {
		return errors.New("No HTTP verb matched")
	}
	return err
}

// transformRequest modifies the request object and sets up a list of
// transforms to execute against the upcoming response.
func (r *Runner) transformRequest(request *model.Request) *model.Request {
	var _responseTransforms []transforms.ResponseTransform
	for i, transform := range r.requestTransforms {
		responseTransform := transform.T(request)
		if responseTransform == nil {
			panic("a transform should never ever return anything but another transform")
		}
		_responseTransforms[i] = responseTransform
	}
	// We replace the response transforms every single request
	r.responseTransforms = _responseTransforms
	return request
}

// updateTransformsFromResponse executes transforms which may read response
// object and which return a set of request transforms to be used in the next
// request.
func (r *Runner) updateTransformsFromResponse(response *model.Response) {
	var _requestTransforms []transforms.RequestTransform
	for i, transform := range r.responseTransforms {
		requestTransform := transform.T(response)
		if requestTransform == nil {
			panic("a transform should never ever return anything but another transform")
		}
		_requestTransforms[i] = requestTransform
	}
	r.requestTransforms = _requestTransforms
}
