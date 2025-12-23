package modules

import (
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/lib"
	markovchord "forbidden_sequencer/sequencer/patterns/markov_chord"
	"forbidden_sequencer/sequencer/sequencers"
)

// NewMarkovChordModule creates patterns that alternate between chord and percussion sections
// conductor: global conductor for timing
// phraseLength: number of ticks in one phrase
// phrasesPerSection: how many phrases before switching sections
// Returns: slice of patterns, chord pattern reference for TUI
func NewMarkovChordModule(
	conductor *conductors.Conductor,
	phraseLength int,
	phrasesPerSection int,
) ([]sequencers.Pattern, *markovchord.ChordPattern) {
	// Create chord pattern (manages section state)
	chordPattern := markovchord.NewChordPattern(
		53,                    // root note (F3)
		lib.MelodicMinorScale, // melodic minor scale
		0.6,                   // velocity
		phraseLength,          // length of phrase in ticks
		phrasesPerSection,     // phrases per section
	)

	// Create percussion patterns (read section state from chord pattern)
	kickPattern := markovchord.NewKickPattern(
		chordPattern,
		"kick",
		50.0, // frequency in Hz
		0.8,  // velocity
	)

	snarePattern := markovchord.NewSnarePattern(
		chordPattern,
		"snare",
		0.7, // velocity
	)

	hihatPattern := markovchord.NewHihatPattern(
		chordPattern,
		"hihat",
		0.6, // velocity
	)

	patterns := []sequencers.Pattern{chordPattern, kickPattern, snarePattern, hihatPattern}

	return patterns, chordPattern
}
