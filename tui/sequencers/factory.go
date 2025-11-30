package sequencers

import (
	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/events"
)

// SequencerFactory creates sequencer configs on demand
type SequencerFactory interface {
	// GetName returns the display name for this sequencer type
	GetName() string

	// Create creates a new sequencer config instance
	Create(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) SequencerConfig
}
