package modulated

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// SimpleHihatPattern fires hihat events in the first half of the phrase (0-50%)
type SimpleHihatPattern struct {
	conductor *conductors.PhraseConductor
	name      string  // event name
	velocity  float64 // event velocity
	paused    bool
}

// NewSimpleHihatPattern creates a new simple hihat pattern
func NewSimpleHihatPattern(
	conductor *conductors.PhraseConductor,
	name string,
	velocity float64,
) *SimpleHihatPattern {
	return &SimpleHihatPattern{
		conductor: conductor,
		name:      name,
		velocity:  velocity,
		paused:    true,
	}
}

// Reset resets the pattern state
func (h *SimpleHihatPattern) Reset() {
	// No state to reset
}

// Play resumes the pattern
func (h *SimpleHihatPattern) Play() {
	h.paused = false
}

// Stop pauses the pattern
func (h *SimpleHihatPattern) Stop() {
	h.paused = true
}

// String returns a string representation of the pattern
func (h *SimpleHihatPattern) String() string {
	return fmt.Sprintf("%s (0-50%% of phrase)", h.name)
}

// GetScheduledEventsForTick implements the Pattern interface
func (h *SimpleHihatPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if h.paused {
		return nil
	}

	// Get tick position from conductor
	nextTickInPhrase := h.conductor.GetNextTickInPhrase()
	phraseLength := h.conductor.GetPhraseLength()

	// Fire if in first half of phrase (0-50%)
	if float64(nextTickInPhrase) < float64(phraseLength)*0.5 {
		// Always use closed hihat (MIDI note 42)
		note := uint8(42)

		// Fire event with duration = 75% of tick
		noteDuration := time.Duration(float64(tickDuration) * 0.75)

		return []events.ScheduledEvent{{
			Event: events.Event{
				Name: h.name,
				Type: events.EventTypeNote,
				A:    float32(note),
				B:    float32(h.velocity),
			},
			Timing: events.Timing{
				Timestamp: nextTickTime,
				Duration:  noteDuration,
			},
		}}
	}

	// Outside range - return no events
	return nil
}
