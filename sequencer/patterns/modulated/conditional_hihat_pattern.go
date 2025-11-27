package modulated

import (
	"fmt"
	"math"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// ConditionalHihatPattern fires hihat events in different ranges based on snare decision:
// - If snare will trigger: fires from 0-50% of phrase
// - If snare won't trigger: fires from 25%-50% of phrase
// The hihat type (closed/open) is determined by the conductor each phrase
// Each successive hit is delayed exponentially later within the tick
type ConditionalHihatPattern struct {
	conductor         *conductors.ModulatedRhythmConductor
	name              string  // event name
	velocity          float64 // event velocity
	paused            bool
	ticksInActiveRange int     // counter for exponential delay calculation
	wasInRange         bool    // tracks if we were in range last tick
}

// NewConditionalHihatPattern creates a new conditional hihat pattern
func NewConditionalHihatPattern(
	conductor *conductors.ModulatedRhythmConductor,
	name string,
	velocity float64,
) *ConditionalHihatPattern {
	return &ConditionalHihatPattern{
		conductor: conductor,
		name:      name,
		velocity:  velocity,
		paused:    true,
	}
}

// Reset resets the pattern state
func (c *ConditionalHihatPattern) Reset() {
	c.ticksInActiveRange = 0
	c.wasInRange = false
}

// Play resumes the pattern
func (c *ConditionalHihatPattern) Play() {
	c.paused = false
}

// Stop pauses the pattern
func (c *ConditionalHihatPattern) Stop() {
	c.paused = true
}

// String returns a string representation of the pattern
func (c *ConditionalHihatPattern) String() string {
	return fmt.Sprintf("%s (conditional range)", c.name)
}

// GetScheduledEventsForTick implements the Pattern interface
func (c *ConditionalHihatPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if c.paused {
		return nil
	}

	// Get tick position from conductor
	nextTickInPhrase := c.conductor.GetNextTickInPhrase()
	phraseLength := c.conductor.GetPhraseLength()

	// Calculate range boundaries
	quarterPoint := phraseLength / 4
	halfPoint := phraseLength / 2

	var inRange bool
	if c.conductor.WillSnareTrigger() {
		// Snare will trigger: hihats fire from 0-50%
		inRange = nextTickInPhrase >= 0 && nextTickInPhrase < halfPoint
	} else {
		// No snare: hihats fire from 25%-50%
		inRange = nextTickInPhrase >= quarterPoint && nextTickInPhrase < halfPoint
	}

	// Track range transitions
	if inRange && !c.wasInRange {
		// Just entered range, reset counter
		c.ticksInActiveRange = 0
	} else if !inRange && c.wasInRange {
		// Just exited range, reset counter
		c.ticksInActiveRange = 0
	}
	c.wasInRange = inRange

	if inRange {
		// Determine which hihat note to use based on conductor's decision
		var note uint8
		if c.conductor.IsHihatClosed() {
			note = 42 // closed hihat
		} else {
			note = 43 // open hihat
		}

		// Calculate exponential delay: delay grows as 1.2^position
		// This makes each successive hit progressively later in the tick
		// Can extend beyond tick boundaries for more pronounced swing
		exponentialFactor := math.Pow(1.2, float64(c.ticksInActiveRange))
		delay := time.Duration(float64(tickDuration) * 0.1 * exponentialFactor)

		// Fire event with duration = 75% of tick
		noteDuration := time.Duration(float64(tickDuration) * 0.75)

		// Increment counter for next tick
		c.ticksInActiveRange++

		return []events.ScheduledEvent{{
			Event: events.Event{
				Name: c.name,
				Type: events.EventTypeNote,
				A:    float32(note),
				B:    float32(c.velocity),
			},
			Timing: events.Timing{
				Timestamp: nextTickTime.Add(delay),
				Duration:  noteDuration,
			},
		}}
	}

	// Outside range - return no events
	return nil
}
