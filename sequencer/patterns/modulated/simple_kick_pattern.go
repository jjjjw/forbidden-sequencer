package modulated

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// SimpleKickPattern fires kick events in the first half of the phrase (0-50%)
type SimpleKickPattern struct {
	conductor *conductors.PhraseConductor
	name      string  // event name
	note      uint8   // MIDI note number
	velocity  float64 // event velocity
	paused    bool
}

// NewSimpleKickPattern creates a new simple kick pattern
func NewSimpleKickPattern(
	conductor *conductors.PhraseConductor,
	name string,
	note uint8,
	velocity float64,
) *SimpleKickPattern {
	return &SimpleKickPattern{
		conductor: conductor,
		name:      name,
		note:      note,
		velocity:  velocity,
		paused:    true,
	}
}

// Reset resets the pattern state
func (k *SimpleKickPattern) Reset() {
	// No state to reset
}

// Play resumes the pattern
func (k *SimpleKickPattern) Play() {
	k.paused = false
}

// Stop pauses the pattern
func (k *SimpleKickPattern) Stop() {
	k.paused = true
}

// String returns a string representation of the pattern
func (k *SimpleKickPattern) String() string {
	return fmt.Sprintf("%s (0-50%% of phrase)", k.name)
}

// GetScheduledEventsForTick implements the Pattern interface
func (k *SimpleKickPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if k.paused {
		return nil
	}

	// Get tick position from conductor
	nextTickInPhrase := k.conductor.GetNextTickInPhrase()
	phraseLength := k.conductor.GetPhraseLength()

	// Fire if in first half of phrase (0-50%)
	if float64(nextTickInPhrase) < float64(phraseLength)*0.5 {
		// Fire event with duration = 75% of tick
		noteDuration := time.Duration(float64(tickDuration) * 0.75)

		return []events.ScheduledEvent{{
			Event: events.Event{
				Name: k.name,
				Type: events.EventTypeNote,
				A:    float32(k.note),
				B:    float32(k.velocity),
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
