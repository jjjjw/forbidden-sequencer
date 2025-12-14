package modulated

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// ChordPattern fires multiple notes simultaneously to create a chord
type ChordPattern struct {
	conductor *conductors.PhraseConductor
	name      string    // event name (for MIDI channel mapping)
	notes     []uint8   // MIDI note numbers for the chord
	velocity  float64   // event velocity
	startTick int       // first tick in phrase when pattern fires (inclusive)
	endTick   int       // last tick in phrase when pattern fires (exclusive)
	paused    bool
}

// NewChordPattern creates a new chord pattern
// notes: array of MIDI note numbers to play simultaneously
// startTick: first tick to fire (inclusive)
// endTick: last tick to fire (exclusive)
func NewChordPattern(
	conductor *conductors.PhraseConductor,
	name string,
	notes []uint8,
	velocity float64,
	startTick int,
	endTick int,
) *ChordPattern {
	return &ChordPattern{
		conductor: conductor,
		name:      name,
		notes:     notes,
		velocity:  velocity,
		startTick: startTick,
		endTick:   endTick,
		paused:    true,
	}
}

// Reset resets the pattern state
func (c *ChordPattern) Reset() {
	// No internal state to reset
}

// Play resumes the pattern
func (c *ChordPattern) Play() {
	c.paused = false
}

// Stop pauses the pattern
func (c *ChordPattern) Stop() {
	c.paused = true
}

// String returns a string representation of the pattern
func (c *ChordPattern) String() string {
	return fmt.Sprintf("%s chord (ticks %d-%d, %d notes)", c.name, c.startTick, c.endTick, len(c.notes))
}

// GetScheduledEventsForTick implements the Pattern interface
func (c *ChordPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if c.paused {
		return nil
	}

	// Get tick position from conductor
	nextTickInPhrase := c.conductor.GetNextTickInPhrase()

	// Check if next tick is in the active range
	inRange := nextTickInPhrase >= c.startTick && nextTickInPhrase < c.endTick

	if inRange {
		// Fire all notes in the chord simultaneously
		// Create a short stab duration (25% of tick for a percussive chord hit)
		noteDuration := time.Duration(float64(tickDuration) * 0.25)

		scheduledEvents := make([]events.ScheduledEvent, len(c.notes))
		for i, note := range c.notes {
			scheduledEvents[i] = events.ScheduledEvent{
				Event: events.Event{
					Name: c.name,
					Type: events.EventTypeNote,
					Params: map[string]float32{
						"midi_note": float32(note),
						"amp":       float32(c.velocity),
					},
				},
				Timing: events.Timing{
					Timestamp: nextTickTime,
					Duration:  noteDuration,
				},
			}
		}
		return scheduledEvents
	}

	// Outside range - return no events
	return nil
}
