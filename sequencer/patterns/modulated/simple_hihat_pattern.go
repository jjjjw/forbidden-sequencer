package modulated

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// SimpleHihatPattern fires hihat events in the first half of the phrase (0-50%)
type SimpleHihatPattern struct {
	conductor    *conductors.PhraseConductor
	name         string  // event name
	velocity     float64 // event velocity
	subdivision  int     // number of times to fire per tick
	paused       bool
}

// NewSimpleHihatPattern creates a new simple hihat pattern
func NewSimpleHihatPattern(
	conductor *conductors.PhraseConductor,
	name string,
	velocity float64,
	subdivision int,
) *SimpleHihatPattern {
	return &SimpleHihatPattern{
		conductor:   conductor,
		name:        name,
		velocity:    velocity,
		subdivision: subdivision,
		paused:      true,
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

// SetSubdivision updates the subdivision value (how many times per tick to fire)
func (h *SimpleHihatPattern) SetSubdivision(subdivision int) {
	h.subdivision = subdivision
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

		// Generate events based on subdivision
		var scheduledEvents []events.ScheduledEvent

		// Calculate time between subdivisions and duration for each note
		subdivisionDuration := tickDuration / time.Duration(h.subdivision)
		noteDuration := time.Duration(float64(subdivisionDuration) * 0.75)

		// Create an event for each subdivision
		for i := 0; i < h.subdivision; i++ {
			eventTime := nextTickTime.Add(subdivisionDuration * time.Duration(i))

			scheduledEvents = append(scheduledEvents, events.ScheduledEvent{
				Event: events.Event{
					Name: h.name,
					Type: events.EventTypeNote,
					A:    float32(note),
					B:    float32(h.velocity),
				},
				Timing: events.Timing{
					Timestamp: eventTime,
					Duration:  noteDuration,
				},
			})
		}

		return scheduledEvents
	}

	// Outside range - return no events
	return nil
}
