package runner

import (
	"errors"
	"sync"
	"time"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/transforms"
)

// Operation is used to send messages about stopping and starting to the HarRunner
// goroutine that's running the Har in a loop.
type Operation uint16

const (
	// Pause tells the HarRunner goroutine to sleep until another operation is sent
	Pause Operation = iota
	// Kill will halt the goroutine and remove it from the running tasks
	Kill
	// Continue will un-pause a goroutine. Has no effect on non-paused
	// goroutines.
	Continue
)

// Runner expresses the API of running an entire archive
type Runner interface {
	Run() error
	Pause()
	Continue()
	Kill()
	GetDoneChannel() chan bool
}

// HarRunner encapsulates a single goroutine reading and replaying a HAR
type HarRunner struct {
	m                      sync.Mutex
	Har                    *model.Har
	Running                bool
	Velocity               float64
	StartTime              time.Time
	operationChannel       chan Operation
	currentEntryNumChannel chan int
	DoneChannel            chan bool
	requestTransforms      []transforms.RequestTransform
	responseTransforms     []transforms.ResponseTransform
	Executor               Executor
}

var _ Runner = &HarRunner{}

// This is the list (implemented as a map so we can use instance pointers) of
// currently-running HarRunner instances.
type runnerList struct {
	items map[*HarRunner]bool
	m     sync.Mutex
}

var runners = &runnerList{
	items: map[*HarRunner]bool{},
	m:     sync.Mutex{},
}

// NewHarRunner accepts a full HAR and begins to replay the contents at the
// originally-recorded timing intervals.
func NewHarRunner(har *model.Har, executor Executor, transforms []transforms.RequestTransform, velocity float64) Runner {
	runner := &HarRunner{
		operationChannel:       make(chan Operation, 1),
		StartTime:              time.Now(),
		Har:                    har,
		Velocity:               velocity,
		Running:                false,
		DoneChannel:            make(chan bool),
		currentEntryNumChannel: make(chan int, 1),
		Executor:               executor,
		requestTransforms:      transforms,
	}

	runner.Run()

	return runner
}

// Run begins to perform http requests and listens for operational commands
func (r *HarRunner) Run() error {
	runners.m.Lock()
	defer runners.m.Unlock()
	if runners.items[r] {
		return errors.New("Attempting to start the same Runner twice, maybe you meant to Continue() it?")
	}
	// Add this runner to the list of runners
	runners.items[r] = true
	r.Running = true

	// And enqueue processing of the first entry
	r.currentEntryNumChannel <- 0
	r.operationChannel <- Continue

	// This is the main goroutine that runs all of the entries in the HAR once it
	// finishes the last entry it exits.
	go func() {
		for {
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
				} else if operation == Kill {
					r.m.Lock()
					defer r.m.Unlock()

					r.Running = false

					runners.m.Lock()
					defer runners.m.Unlock()
					defer delete(runners.items, r) // Remove this instance from the list
					r.DoneChannel <- true
					return // This is where we shut the whole routine down
				}
			// Check if there's another request to make. If so, play it (play()
			// spawns a goroutine and returns immediately)
			case entryIndex := <-r.currentEntryNumChannel:
				if len(r.Har.Entries) > entryIndex {
					r.play(entryIndex)
				} else {
					// We're done!
					r.operationChannel <- Kill
				}
			}
		}

	}()
	return nil
}

func (r *HarRunner) play(index int) {
	entry := r.Har.Entries[index]
	go func() {
		r.Play(&entry)
		time.Sleep(r.SleepFor(&entry))
		r.currentEntryNumChannel <- index + 1
	}()
}

// Play performs the request described in the Entry
func (r *HarRunner) Play(entry *model.Entry) error {
	transformedRequest := r.transformRequest(*entry.Request)

	var err error
	var response *model.Response

	switch entry.Request.Method {
	case "GET":
		response, err = r.Executor.Get(transformedRequest)
	case "POST":
		response, err = r.Executor.Post(transformedRequest)
	case "PUT":
		response, err = r.Executor.Put(transformedRequest)
	case "DELETE":
		response, err = r.Executor.Delete(transformedRequest)
	case "HEAD":
		response, err = r.Executor.Head(transformedRequest)
	case "PATCH":
		response, err = r.Executor.Patch(transformedRequest)
	default:
		return errors.New("No HTTP verb matched")
	}

	if response != nil {
		r.updateTransformsFromResponse(response)
	}

	return err
}

// Pause halts this runner and the goroutine waits for a Continue() or a Kill()
func (r *HarRunner) Pause() {
	r.operationChannel <- Pause
}

// Continue halts this runner if a Pause was sent. It's not an error to call
// this multiple times, it's idempotent.
func (r *HarRunner) Continue() {
	r.operationChannel <- Continue
}

// Kill halts this runner and stops the running goroutine
func (r *HarRunner) Kill() {
	r.operationChannel <- Kill
}

// GetDoneChannel exposes doneness
func (r *HarRunner) GetDoneChannel() chan bool {
	return r.DoneChannel
}

// transformRequest modifies the request object and sets up a list of
// transforms to execute against the upcoming response.
func (r *HarRunner) transformRequest(request model.Request) model.Request {
	// Clear out any ResponseTransforms set by previous requests. There will
	// always be exactly as many response transforms as request transforms
	// because each kind produces the other.
	r.responseTransforms = make([]transforms.ResponseTransform, len(r.requestTransforms))

	for i, requestTransform := range r.requestTransforms {
		responseTransform := requestTransform.T(&request)
		if responseTransform == nil {
			panic("a transform should never ever return anything but another transform")
		}
		r.responseTransforms[i] = responseTransform
	}

	// The RequestTransform instances may have modified the request
	return request
}

// updateTransformsFromResponse is the inverse of transformRequest. It's not
// really a "transformResponse" though because we don't bother modifying a
// response we get from a remote server, we just use the data in the response
// to produce new RequestTransform instances to use in the future.
func (r *HarRunner) updateTransformsFromResponse(response *model.Response) {

	// There are always exactly as many requestTransforms as responseTransforms
	r.requestTransforms = make([]transforms.RequestTransform, len(r.responseTransforms))
	for i, transform := range r.responseTransforms {
		requestTransform := transform.T(response)
		if requestTransform == nil {
			panic("a transform's .T() should never ever return anything but another transform")
		}
		r.requestTransforms[i] = requestTransform
	}
}

// SleepFor calculates how long has passed since the runner started and,
// considering how long after the HAR recording this particular entry
// happened, returns a time duration that we should sleep so the next request
// happens at the right time.
// This value is adjusted by the Runner's `Velocity`
// Higher `Velocity` equals a shorter sleep between requests
func (r *HarRunner) SleepFor(entry *model.Entry) time.Duration {
	pauseDuration := time.Duration(entry.TimeMs/r.Velocity) * time.Millisecond
	return r.StartTime.Add(pauseDuration).Sub(time.Now())
}
