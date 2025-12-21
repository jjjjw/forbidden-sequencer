package sequencers

import (
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
	randpatterns "forbidden_sequencer/sequencer/patterns/rand"
)

// NewRandRhythmSequencer creates a sequencer with randomized patterns using Markov chains
// baseTickDuration: base time between ticks (before rate modulation)
// phraseLength: number of ticks in one phrase
// adapter: MIDI or other output adapter
// eventChan: channel to send events to
func NewRandRhythmSequencer(baseTickDuration time.Duration, phraseLength int, adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) (*Sequencer, *conductors.ModulatedRhythmConductor) {
	// Create phrase conductor
	phraseConductor := conductors.NewPhraseConductor(baseTickDuration, phraseLength)

	// Create rhythm decision conductor that wraps the phrase conductor
	conductor := conductors.NewModulatedRhythmConductor(phraseConductor)

	kickPattern := randpatterns.NewKickPattern(
		phraseConductor,
		conductor,
		"kick",
		50.0, // frequency in Hz (bass drum)
		0.8,  // velocity
	)

	snarePattern := randpatterns.NewSnarePattern(
		conductor,
		"snare",
		0.7, // velocity
	)

	hihatPattern := randpatterns.NewHihatPattern(
		conductor,
		"hihat",
		0.6, // velocity
	)

	// FM voice 1: melodic minor scale, middle register
	// Both FM patterns use event name "fm" with max_voices=2
	fm1Pattern := randpatterns.NewFMPattern(
		0.5,                   // velocity
		60,                    // root note (C4)
		lib.MelodicMinorScale, // melodic minor scale
		2,                     // 2 octave range
		43,                    // random seed
	)

	// FM voice 2: melodic minor scale, higher register
	fm2Pattern := randpatterns.NewFMPattern(
		0.4,                   // velocity (slightly quieter)
		72,                    // root note (C5, one octave up)
		lib.MelodicMinorScale, // melodic minor scale
		2,                     // 2 octave range
		84,                    // random seed (different from fm1)
	)

	patterns := []Pattern{kickPattern, snarePattern, hihatPattern, fm1Pattern, fm2Pattern}

	return NewSequencer(patterns, phraseConductor, adapter, eventChan), conductor
}
