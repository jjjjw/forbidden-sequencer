package sequencers

import (
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/lib"
	"forbidden_sequencer/sequencer/patterns/arp"
)

// NewArpSequencer creates an arpeggiator sequencer
// baseTickDuration: base time between ticks
// sequence: array of scale degrees (use arp.RestValue for rests)
// scale: musical scale to use (lib.MajorScale or lib.MinorScale)
// rootNote: base MIDI note (e.g., 60 for middle C)
// adapter: MIDI or other output adapter
// debug: debug mode flag
func NewArpSequencer(
	baseTickDuration time.Duration,
	sequence []int,
	scale lib.Scale,
	rootNote uint8,
	adapter adapters.EventAdapter,
	debug bool,
) (*Sequencer, *arp.ArpPattern) {
	// Create phrase conductor
	phraseConductor := conductors.NewPhraseConductor(baseTickDuration, len(sequence))

	// Create arp pattern
	arpPattern := arp.NewArpPattern(
		"arp",
		sequence,
		scale,
		rootNote,
		0.8, // velocity
	)

	patterns := []Pattern{arpPattern}

	return NewSequencer(patterns, phraseConductor, adapter, debug), arpPattern
}
