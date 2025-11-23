package modulated

import (
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// GatedPattern fires events only when within a specified tick range of the phrase
type GatedPattern struct {
	conductor *conductors.ModulatedConductor
	name      string  // event name (e.g., "kick", "hihat")
	note      uint8   // MIDI note number
	velocity  float64 // event velocity
	startTick int     // first tick in phrase when pattern fires (inclusive)
	endTick   int     // last tick in phrase when pattern fires (exclusive)
	paused    bool
}

// NewGatedPattern creates a new gated pattern
// startTick: first tick to fire (inclusive)
// endTick: last tick to fire (exclusive)
func NewGatedPattern(
	conductor *conductors.ModulatedConductor,
	name string,
	note uint8,
	velocity float64,
	startTick int,
	endTick int,
) *GatedPattern {
	return &GatedPattern{
		conductor: conductor,
		name:      name,
		note:      note,
		velocity:  velocity,
		startTick: startTick,
		endTick:   endTick,
		paused:    true,
	}
}

// Reset resets the pattern state
func (g *GatedPattern) Reset() {
	// No internal state to reset
}

// Play resumes the pattern
func (g *GatedPattern) Play() {
	g.paused = false
}

// Stop pauses the pattern
func (g *GatedPattern) Stop() {
	g.paused = true
}

// GetScheduledEventsForTick implements the Pattern interface
func (g *GatedPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if g.paused {
		return nil
	}

	// Get tick position from conductor
	nextTickInPhrase := g.conductor.GetNextTickInPhrase()

	// Check if next tick is in the active range
	inRange := nextTickInPhrase >= g.startTick && nextTickInPhrase < g.endTick

	if inRange {
		// Fire event with duration = 75% of tick
		noteDuration := time.Duration(float64(tickDuration) * 0.75)
		return []events.ScheduledEvent{{
			Event: events.Event{
				Name: g.name,
				Type: events.EventTypeNote,
				A:    float32(g.note),
				B:    float32(g.velocity),
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
