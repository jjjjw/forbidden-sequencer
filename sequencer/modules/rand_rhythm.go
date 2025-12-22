package modules

import (
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/lib"
	randpatterns "forbidden_sequencer/sequencer/patterns/rand"
	"forbidden_sequencer/sequencer/sequencers"
)

// NewRandRhythmModule creates randomized rhythm patterns using Markov chains
// conductor: global conductor for timing
// phraseLength: number of ticks in one phrase
// Returns: slice of patterns
func NewRandRhythmModule(conductor *conductors.Conductor, phraseLength int) []sequencers.Pattern {
	// Create snare pattern first (other patterns reference it for trigger state)
	snarePattern := randpatterns.NewSnarePattern(
		conductor,
		"snare",
		0.7,          // velocity
		phraseLength, // phrase length
	)

	kickPattern := randpatterns.NewKickPattern(
		conductor,
		snarePattern, // reference to snare for trigger state
		"kick",
		50.0, // frequency in Hz (bass drum)
		0.8,  // velocity
	)

	hihatPattern := randpatterns.NewHihatPattern(
		conductor,
		snarePattern, // reference to snare for trigger state
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

	return []sequencers.Pattern{snarePattern, kickPattern, hihatPattern, fm1Pattern, fm2Pattern}
}
