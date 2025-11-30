package sequencers

import (
	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/patterns/techno"
)

// NewTechnoSequencer creates a techno sequencer with kick and hihat patterns
// bpm: beats per minute
// adapter: MIDI or other output adapter
// eventChan: channel to send events to
// debug: debug mode flag
func NewTechnoSequencer(bpm float64, adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent, debug bool) *Sequencer {
	// Create conductor
	// ticksPerBeat: tick resolution (4 = 16th notes, 8 = 32nd notes)
	conductor := conductors.NewCommonTimeConductor(bpm, 4)

	// Create pattern with conductor reference
	pattern := techno.NewTechnoPattern(conductor)

	// Assemble into sequencer
	patternsSlice := []Pattern{pattern}

	return NewSequencer(patternsSlice, conductor, adapter, eventChan, debug)
}
