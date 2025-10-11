package lanes

import "forbidden_sequencer/sequencer"

// Lane is an interface for generating events
// The sequencer manages timing/rate, lanes just generate events
type Lane interface {
	// GetNextEvent generates and returns the next event
	GetNextEvent() (sequencer.Event, error)
}
