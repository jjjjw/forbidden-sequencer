package modulated

import (
	"fmt"
	"math"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// ConditionalHihatPattern fires hihat events using a Markov chain
// States: "playing" and "silent"
// Transition probabilities:
//   - playing → playing: 0.3 (30% keep playing)
//   - playing → silent: 0.7 (70% go silent)
//   - silent → silent: 0.5 (50% stay silent)
//   - silent → playing: 0.5 (50% start playing)
// The hihat type (closed/open) is determined by the conductor each phrase
// Each successive hit is delayed exponentially later within the tick
// Silences after snare fires in the phrase
type ConditionalHihatPattern struct {
	conductor         *conductors.ModulatedRhythmConductor
	name              string           // event name
	velocity          float64          // event velocity
	paused            bool
	ticksInActiveRange int              // counter for exponential delay calculation
	wasInRange         bool             // tracks if we were in range last tick
	chain              *lib.MarkovChain // Markov chain for play/silent decisions
}

// NewConditionalHihatPattern creates a new conditional hihat pattern
func NewConditionalHihatPattern(
	conductor *conductors.ModulatedRhythmConductor,
	name string,
	velocity float64,
) *ConditionalHihatPattern {
	// Create Markov chain with two states: playing and silent
	chain := lib.NewMarkovChain(43) // Different seed from kick

	// Set transition probabilities (less busy than kick)
	// When playing: 30% keep playing, 70% go silent
	chain.SetTransitionProbability("playing", "playing", 0.3)
	chain.SetTransitionProbability("playing", "silent", 0.7)

	// When silent: 50% stay silent, 50% start playing
	chain.SetTransitionProbability("silent", "silent", 0.5)
	chain.SetTransitionProbability("silent", "playing", 0.5)

	return &ConditionalHihatPattern{
		conductor: conductor,
		name:      name,
		velocity:  velocity,
		paused:    true,
		chain:     chain,
	}
}

// Reset resets the pattern state
func (c *ConditionalHihatPattern) Reset() {
	c.ticksInActiveRange = 0
	c.wasInRange = false
	c.chain.Reset()
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
	snareTriggerTick := c.conductor.GetSnareTriggerTick()

	// If snare will trigger and we're at or past the snare tick, stay silent
	if c.conductor.WillSnareTrigger() && nextTickInPhrase >= snareTriggerTick {
		return nil
	}

	// Use Markov chain to decide whether to play this tick
	state, err := c.chain.Next()
	if err != nil {
		return nil
	}

	// Track range transitions for exponential delay
	inRange := state == "playing"
	if inRange && !c.wasInRange {
		// Just started playing, reset counter
		c.ticksInActiveRange = 0
	} else if !inRange && c.wasInRange {
		// Just stopped playing, reset counter
		c.ticksInActiveRange = 0
	}
	c.wasInRange = inRange

	// Only fire if we're in the "playing" state
	if state == "playing" {
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

	// Silent state - return no events
	return nil
}
