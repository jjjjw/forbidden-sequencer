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
func NewModulatedRhythmSequencer(baseTickDuration time.Duration, phraseLength int, adapter adapters.EventAdapter, debug bool) (*Sequencer, *conductors.ModulatedRhythmConductor) {
	// Create phrase conductor
	phraseConductor := conductors.NewPhraseConductor(baseTickDuration, phraseLength)

	// Create rhythm decision conductor that wraps the phrase conductor
	conductor := conductors.NewModulatedRhythmConductor(phraseConductor)

	// Create patterns with conditional logic:
	// - Kick: fires in bursts of 3-4 hits, then pauses 2-3 ticks (randomized per burst)
	// - Snare: fires at 3/4 point, 33% chance per phrase
	// - Hihat: fires 0-50% if snare triggers, 25%-50% otherwise
	//          randomly selects closed (42) or open (43) per phrase (75% closed)
	//          each successive hit is delayed exponentially later in the tick

	kickPattern := modulated.NewBurstKickPattern(
		phraseConductor,
		"kick",
		36,  // MIDI note (bass drum)
		0.8, // velocity
	)

	snarePattern := modulated.NewSnarePattern(
		conductor,
		"snare",
		37,  // MIDI note (snare)
		0.7, // velocity
	)

	hihatPattern := modulated.NewConditionalHihatPattern(
		conductor,
		"hihat",
		0.6, // velocity
	)

	patterns := []Pattern{kickPattern, snarePattern, hihatPattern}

	return NewSequencer(patterns, phraseConductor, adapter, debug), conductor
}
