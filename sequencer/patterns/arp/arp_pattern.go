package arp

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// RestValue is a special value in the sequence representing a rest
const RestValue = -1000

// ArpPattern plays a monophonic arpeggio sequence based on scale degrees
// Sequence values are scale degree indexes, with RestValue (-1000) representing rests
type ArpPattern struct {
	name        string     // event name
	sequence    []int      // scale degree indexes or RestValue for rests
	scale       lib.Scale  // musical scale to use
	rootNote    uint8      // base MIDI note
	transpose   int        // semitone offset (for octave/fifth shifts)
	currentStep int        // current position in sequence
	velocity    float64    // note velocity
	paused      bool
}

// NewArpPattern creates a new arpeggiator pattern
func NewArpPattern(name string, sequence []int, scale lib.Scale, rootNote uint8, velocity float64) *ArpPattern {
	return &ArpPattern{
		name:        name,
		sequence:    sequence,
		scale:       scale,
		rootNote:    rootNote,
		transpose:   0,
		currentStep: 0,
		velocity:    velocity,
		paused:      true,
	}
}

// Reset resets the pattern to the beginning
func (a *ArpPattern) Reset() {
	a.currentStep = 0
}

// Play resumes the pattern
func (a *ArpPattern) Play() {
	a.paused = false
}

// Stop pauses the pattern
func (a *ArpPattern) Stop() {
	a.paused = true
}

// ShiftOctaveUp transposes up one octave
func (a *ArpPattern) ShiftOctaveUp() {
	a.transpose += 12
}

// ShiftOctaveDown transposes down one octave
func (a *ArpPattern) ShiftOctaveDown() {
	a.transpose -= 12
}

// ShiftFifthUp transposes up a perfect fifth
func (a *ArpPattern) ShiftFifthUp() {
	a.transpose += 7
}

// ShiftFifthDown transposes down a perfect fifth
func (a *ArpPattern) ShiftFifthDown() {
	a.transpose -= 7
}

// SetRootNote changes the root note
func (a *ArpPattern) SetRootNote(note uint8) {
	a.rootNote = note
}

// GetRootNote returns the current root note
func (a *ArpPattern) GetRootNote() uint8 {
	return a.rootNote
}

// GetTranspose returns the current transpose amount
func (a *ArpPattern) GetTranspose() int {
	return a.transpose
}

// String returns a string representation of the pattern
func (a *ArpPattern) String() string {
	return fmt.Sprintf("%s (root: %d, transpose: %+d, step: %d/%d)",
		a.name, a.rootNote, a.transpose, a.currentStep, len(a.sequence))
}

// GetScheduledEventsForTick implements the Pattern interface
func (a *ArpPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if a.paused {
		return nil
	}

	// If sequence is empty, return no events
	if len(a.sequence) == 0 {
		return nil
	}

	// Get current value from sequence
	currentValue := a.sequence[a.currentStep]

	// Advance to next step
	a.currentStep = (a.currentStep + 1) % len(a.sequence)

	// Check if this is a rest
	if currentValue == RestValue {
		// Return rest event
		return []events.ScheduledEvent{{
			Event: events.Event{
				Name: a.name,
				Type: events.EventTypeRest,
			},
			Timing: events.Timing{
				Timestamp: nextTickTime,
				Duration:  tickDuration,
			},
		}}
	}

	// Convert scale degree to MIDI note
	note := a.scale.NoteAt(a.rootNote, currentValue)

	// Apply transpose
	transposedNote := int(note) + a.transpose
	if transposedNote < 0 {
		transposedNote = 0
	}
	if transposedNote > 127 {
		transposedNote = 127
	}

	// Fire note with duration = 90% of tick
	noteDuration := time.Duration(float64(tickDuration) * 0.9)

	return []events.ScheduledEvent{{
		Event: events.Event{
			Name: a.name,
			Type: events.EventTypeNote,
			A:    float32(transposedNote),
			B:    float32(a.velocity),
		},
		Timing: events.Timing{
			Timestamp: nextTickTime,
			Duration:  noteDuration,
		},
	}}
}
