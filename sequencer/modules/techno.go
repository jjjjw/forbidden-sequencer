package modules

import (
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/patterns/techno"
	"forbidden_sequencer/sequencer/sequencers"
)

// NewTechnoModule creates a techno pattern with kick and hihat
// conductor: global conductor for timing
// ticksPerBeat: tick resolution (4 = 16th notes, 8 = 32nd notes)
// Returns: slice of patterns
func NewTechnoModule(conductor *conductors.Conductor, ticksPerBeat int) []sequencers.Pattern {
	// Create pattern with conductor reference
	pattern := techno.NewTechnoPattern(conductor, ticksPerBeat)

	// Assemble into pattern slice
	return []sequencers.Pattern{pattern}
}
