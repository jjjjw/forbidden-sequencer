package event_generators

import "forbidden_sequencer/sequencer"

// EventGenerator is an interface for generating events
// The sequencer manages timing, generators just produce events
type EventGenerator interface {
	// GetNextEvent generates and returns the next event
	GetNextEvent() (sequencer.Event, error)
}
