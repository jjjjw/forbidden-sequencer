package modules

import (
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/patterns/modulated"
	"forbidden_sequencer/sequencer/sequencers"
)

// NewModulatedRhythmModule creates simple phrase-based patterns
// kick and hihat both fire in first half (0-50%)
// conductor: global conductor for timing
// phraseLength: number of ticks in one phrase
// Returns: slice of patterns, kick pattern, hihat pattern (for TUI control)
func NewModulatedRhythmModule(
	conductor *conductors.Conductor,
	phraseLength int,
) ([]sequencers.Pattern, *modulated.SimpleKickPattern, *modulated.SimpleHihatPattern) {
	// Create simple patterns based on phrase position:
	// - Kick: fires every tick in first half (0-50%)
	// - Hihat: fires every tick in middle section (25-75%)

	kickPattern := modulated.NewSimpleKickPattern(
		conductor,
		"kick",
		36,           // MIDI note (bass drum)
		0.8,          // velocity
		1,            // subdivision (default: once per tick)
		phraseLength, // phrase length
	)

	hihatPattern := modulated.NewSimpleHihatPattern(
		conductor,
		"hihat",
		0.6,          // velocity
		1,            // subdivision (default: once per tick)
		phraseLength, // phrase length
	)

	patterns := []sequencers.Pattern{kickPattern, hihatPattern}

	return patterns, kickPattern, hihatPattern
}
