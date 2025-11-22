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
	velocity float64,
	startTick int,
	endTick int,
) *GatedPattern {
	return &GatedPattern{
		conductor: conductor,
		name:      name,
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

// GetNextScheduledEvent implements the Pattern interface
func (g *GatedPattern) GetNextScheduledEvent() (events.ScheduledEvent, error) {
	// When paused, return short rests
	if g.paused {
		return events.ScheduledEvent{
			Event: events.Event{
				Name: "rest",
				Type: events.EventTypeRest,
			},
			Timing: events.Timing{
				Delta:    10 * time.Millisecond,
				Duration: 0,
			},
		}, nil
	}

	// Get timing from conductor
	nextTickTime := g.conductor.GetNextTickTime()
	nextTickInPhrase := g.conductor.GetNextTickInPhrase()

	// Calculate delta to next tick
	delta := time.Until(nextTickTime)

	// Check if next tick is in the active range
	inRange := nextTickInPhrase >= g.startTick && nextTickInPhrase < g.endTick

	if inRange {
		// Fire event
		return events.ScheduledEvent{
			Event: events.Event{
				Name: g.name,
				Type: events.EventTypeNote,
				A:    float32(g.velocity),
			},
			Timing: events.Timing{
				Delta:    delta,
				Duration: 50 * time.Millisecond,
			},
		}, nil
	}

	// Outside range - return rest until next tick
	return events.ScheduledEvent{
		Event: events.Event{
			Name: "rest",
			Type: events.EventTypeRest,
		},
		Timing: events.Timing{
			Delta:    delta,
			Duration: 0,
		},
	}, nil
}
