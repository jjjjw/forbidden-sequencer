package sequencers

import (
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/patterns/modulated"
)

// NewModulatedRhythmSequencer creates a sequencer with modulated timing and gated patterns
// baseTickDuration: base time between ticks (before rate modulation)
// phraseLength: number of ticks in one phrase
// adapter: MIDI or other output adapter
// debug: debug mode flag
func NewModulatedRhythmSequencer(baseTickDuration time.Duration, phraseLength int, adapter adapters.EventAdapter, debug bool) (*Sequencer, *conductors.PhraseConductor) {
	// Create modulated conductor
	conductor := conductors.NewPhraseConductor(baseTickDuration, phraseLength)

	// Create patterns with overlapping ranges
	// Example: 16 tick phrase
	// Kick: ticks 0-12 (fires), 12-16 (rests)
	// Synced: ticks 4-12 (fires), 0-4 and 12-16 (rests)
	// Overlap: ticks 4-12 (both fire together)

	kickPattern := modulated.NewGatedPattern(
		conductor,
		"kick",
		36,            // MIDI note (bass drum)
		0.8,           // velocity
		0,             // startTick
		phraseLength*3/4, // endTick (first 75% of phrase)
	)

	syncedPattern := modulated.NewGatedPattern(
		conductor,
		"hihat",
		42,               // MIDI note (closed hihat)
		0.6,              // velocity
		phraseLength/4,   // startTick (starts at 25%)
		phraseLength*3/4, // endTick (ends at 75%)
	)

	patterns := []Pattern{kickPattern, syncedPattern}

	return NewSequencer(patterns, conductor, adapter, debug), conductor
}
