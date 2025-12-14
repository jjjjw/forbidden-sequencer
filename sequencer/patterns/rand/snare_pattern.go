package modulated

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// SnarePattern fires a snare event at 3/4 of the phrase, but only if conductor decides to trigger
type SnarePattern struct {
	conductor *conductors.ModulatedRhythmConductor
	name      string  // event name
	velocity  float64 // event velocity
	paused    bool
}

// NewSnarePattern creates a new snare pattern
func NewSnarePattern(
	conductor *conductors.ModulatedRhythmConductor,
	name string,
	velocity float64,
) *SnarePattern {
	return &SnarePattern{
		conductor: conductor,
		name:      name,
		velocity:  velocity,
		paused:    true,
	}
}

// Reset resets the pattern state
func (s *SnarePattern) Reset() {
	// No internal state to reset
}

// Play resumes the pattern
func (s *SnarePattern) Play() {
	s.paused = false
}

// Stop pauses the pattern
func (s *SnarePattern) Stop() {
	s.paused = true
}

// String returns a string representation of the pattern
func (s *SnarePattern) String() string {
	return fmt.Sprintf("%s (conditional at 3/4)", s.name)
}

// GetScheduledEventsForTick implements the Pattern interface
func (s *SnarePattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if s.paused {
		return nil
	}

	// Get tick position from conductor
	nextTickInPhrase := s.conductor.GetNextTickInPhrase()
	snareTriggerTick := s.conductor.GetSnareTriggerTick()
	willTrigger := s.conductor.WillSnareTrigger()

	// Check if this is the snare trigger tick AND conductor decided to trigger
	if nextTickInPhrase == snareTriggerTick && willTrigger {
		// Fire event with duration = 75% of tick
		noteDuration := time.Duration(float64(tickDuration) * 0.75)
		return []events.ScheduledEvent{{
			Event: events.Event{
				Name: s.name,
				Type: events.EventTypeNote,
				A:    0, // snare doesn't use pitch
				B:    float32(s.velocity),
			},
			Timing: events.Timing{
				Timestamp: nextTickTime,
				Duration:  noteDuration,
			},
		}}
	}

	// Don't trigger - return no events
	return nil
}
