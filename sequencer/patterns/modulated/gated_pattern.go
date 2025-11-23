package modulated

import (
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// GatedPattern fires events only when within a specified tick range of the phrase
type GatedPattern struct {
	conductor         *conductors.ModulatedConductor
	name              string  // event name (e.g., "kick", "hihat")
	note              uint8   // MIDI note number
	velocity          float64 // event velocity
	startTick         int     // first tick in phrase when pattern fires (inclusive)
	endTick           int     // last tick in phrase when pattern fires (exclusive)
	paused            bool
	lastScheduledTime time.Time // tracks last scheduled event time to avoid duplicates
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
				Timestamp: time.Now().Add(10 * time.Millisecond),
				Duration:  0,
			},
		}, nil
	}

	// Get timing from conductor
	nextTickTime := g.conductor.GetNextTickTime()
	nextTickInPhrase := g.conductor.GetNextTickInPhrase()

	// If we've already scheduled this tick, return a short rest
	if !g.lastScheduledTime.IsZero() && !nextTickTime.After(g.lastScheduledTime) {
		return events.ScheduledEvent{
			Event: events.Event{
				Name: "rest",
				Type: events.EventTypeRest,
			},
			Timing: events.Timing{
				Timestamp: time.Now().Add(10 * time.Millisecond),
				Duration:  0,
			},
		}, nil
	}

	// Update tracking
	g.lastScheduledTime = nextTickTime

	// Check if next tick is in the active range
	inRange := nextTickInPhrase >= g.startTick && nextTickInPhrase < g.endTick

	if inRange {
		// Fire event with duration = 75% of tick
		tickDuration := g.conductor.GetTickDuration()
		noteDuration := time.Duration(float64(tickDuration) * 0.75)
		return events.ScheduledEvent{
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
		}, nil
	}

	// Outside range - return rest until next tick
	return events.ScheduledEvent{
		Event: events.Event{
			Name: "rest",
			Type: events.EventTypeRest,
		},
		Timing: events.Timing{
			Timestamp: nextTickTime,
			Duration:  0,
		},
	}, nil
}
