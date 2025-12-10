package sequencers

import (
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/patterns/modulated"
)

// NewModulatedRhythmSequencer creates a sequencer with simple phrase-based patterns
// kick and hihat both fire in first half (0-50%)
// baseTickDuration: base time between ticks (before rate modulation)
// phraseLength: number of ticks in one phrase
// adapter: output adapter
// eventChan: channel to send events to
// debug: debug mode flag
// Returns: sequencer, conductor, kickPattern, hihatPattern
func NewModulatedRhythmSequencer(baseTickDuration time.Duration, phraseLength int, adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent, debug bool) (*Sequencer, *conductors.PhraseConductor, *modulated.SimpleKickPattern, *modulated.SimpleHihatPattern) {
	// Create phrase conductor
	phraseConductor := conductors.NewPhraseConductor(baseTickDuration, phraseLength)

	// Create simple patterns based on phrase position:
	// - Kick: fires every tick in first half (0-50%)
	// - Hihat: fires every tick in middle section (25-75%)

	kickPattern := modulated.NewSimpleKickPattern(
		phraseConductor,
		"kick",
		36,  // MIDI note (bass drum)
		0.8, // velocity
		1,   // subdivision (default: once per tick)
	)

	hihatPattern := modulated.NewSimpleHihatPattern(
		phraseConductor,
		"hihat",
		0.6, // velocity
		1,   // subdivision (default: once per tick)
	)

	patterns := []Pattern{kickPattern, hihatPattern}

	return NewSequencer(patterns, phraseConductor, adapter, eventChan, debug), phraseConductor, kickPattern, hihatPattern
}
