package test

import (
	"forbidden_sequencer/sequencer/events"
)

// SimpleKickPattern fires a kick on every 4th tick
type SimpleKickPattern struct {
	paused bool
}

// NewSimpleKickPattern creates a simple test kick pattern
func NewSimpleKickPattern() *SimpleKickPattern {
	return &SimpleKickPattern{
		paused: true,
	}
}

// GetEventsForTick implements the Pattern interface
func (s *SimpleKickPattern) GetEventsForTick(tick int64) []events.TickEvent {
	if s.paused {
		return nil
	}

	// Fire kick on every 4th tick
	if tick%4 == 0 {
		return []events.TickEvent{{
			Event: events.Event{
				Name: "kick",
				Type: events.EventTypeNote,
				Params: map[string]float32{
					"freq": 50.0,
					"amp":  0.8,
				},
			},
			TickTiming: events.TickTiming{
				Tick:          tick,
				OffsetPercent: 0.0,  // On the beat
				DurationTicks: 0.75, // 75% of tick duration
			},
		}}
	}

	return nil
}

// Reset resets the pattern
func (s *SimpleKickPattern) Reset() {}

// Play starts the pattern
func (s *SimpleKickPattern) Play() {
	s.paused = false
}

// Stop pauses the pattern
func (s *SimpleKickPattern) Stop() {
	s.paused = true
}
