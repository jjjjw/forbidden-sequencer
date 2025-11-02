package sequencers

import (
	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/patterns/techno"
)

// NewTechnoSequencer creates a techno sequencer with kick and hihat patterns
// bpm: beats per minute
// ticksPerBeat: tick resolution (4 = 16th notes, 8 = 32nd notes)
// adapter: MIDI or other output adapter
// debug: debug mode flag
func NewTechnoSequencer(bpm float64, ticksPerBeat int, adapter adapters.EventAdapter, debug bool) *Sequencer {
	// Create conductor
	conductor := conductors.NewCommonTimeConductor(ticksPerBeat, bpm)

	// Create patterns with conductor reference
	kick := techno.NewKickPattern(conductor)
	hihat := techno.NewHihatPattern(conductor)

	// Assemble into sequencer
	patternsSlice := []Pattern{kick, hihat}

	return NewSequencer(patternsSlice, conductor, adapter, debug)
}
