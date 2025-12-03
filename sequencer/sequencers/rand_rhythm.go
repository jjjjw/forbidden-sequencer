package sequencers

import (
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	randpatterns "forbidden_sequencer/sequencer/patterns/rand"
)

// NewRandRhythmSequencer creates a sequencer with randomized patterns using Markov chains
// baseTickDuration: base time between ticks (before rate modulation)
// phraseLength: number of ticks in one phrase
// adapter: MIDI or other output adapter
// eventChan: channel to send events to
// debug: debug mode flag
func NewRandRhythmSequencer(baseTickDuration time.Duration, phraseLength int, adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent, debug bool) (*Sequencer, *conductors.ModulatedRhythmConductor) {
	// Create phrase conductor
	phraseConductor := conductors.NewPhraseConductor(baseTickDuration, phraseLength)

	// Create rhythm decision conductor that wraps the phrase conductor
	conductor := conductors.NewModulatedRhythmConductor(phraseConductor)

	// Create patterns with conditional logic:
	// - Kick: Markov chain (50% keep playing, 50% start playing), silences after snare
	// - Snare: fires at 3/4 point, 33% chance per phrase
	// - Hihat: Markov chain (30% keep playing, 50% start playing), silences after snare
	//          uses MIDI note 42 (closed hihat)
	//          each successive hit is delayed exponentially later in the tick

	kickPattern := randpatterns.NewKickPattern(
		phraseConductor,
		conductor,
		"kick",
		36,  // MIDI note (bass drum)
		0.8, // velocity
	)

	snarePattern := randpatterns.NewSnarePattern(
		conductor,
		"snare",
		37,  // MIDI note (snare)
		0.7, // velocity
	)

	hihatPattern := randpatterns.NewHihatPattern(
		conductor,
		"hihat",
		0.6, // velocity
	)

	patterns := []Pattern{kickPattern, snarePattern, hihatPattern}

	return NewSequencer(patterns, phraseConductor, adapter, eventChan, debug), conductor
}
