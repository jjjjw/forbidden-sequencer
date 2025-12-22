package modules

import (
	"forbidden_sequencer/sequencer/lib"
	"forbidden_sequencer/sequencer/patterns/arp"
	"forbidden_sequencer/sequencer/sequencers"
)

// NewArpModule creates an arpeggiator pattern
// sequence: array of scale degrees (use arp.RestValue for rests)
// scale: musical scale to use (lib.MajorScale or lib.MinorScale)
// rootNote: base MIDI note (e.g., 60 for middle C)
// Returns: slice of patterns, arp pattern reference for TUI
func NewArpModule(
	sequence []int,
	scale lib.Scale,
	rootNote uint8,
) ([]sequencers.Pattern, *arp.ArpPattern) {
	// Create arp pattern
	arpPattern := arp.NewArpPattern(
		"arp",
		sequence,
		scale,
		rootNote,
		0.8, // velocity
	)

	patterns := []sequencers.Pattern{arpPattern}

	return patterns, arpPattern
}
