package runner

import (
	"errors"
	"fmt"
	"github.com/JackDanger/traffic/model"
	"sync"
	"time"
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

// Runner encapsulates a single goroutine reading and replaying a HAR
type Runner struct {
	m                sync.Mutex
	Har              *model.Har
	Running          bool
	StartTime        time.Time
	operationChannel chan Operation
	requestChannel   chan *model.Entry
}

// This is the list (implemented as a map so we can use instance pointers) of
// currently-running Runner instances.
var runners struct {
	items map[*Runner]bool
	m     sync.Mutex
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
func (r *Runner) Play(entry *model.Entry) {
	fmt.Printf("performing request for %s\n", entry.Request)
	// TODO: do some http stuff
}
