package techno

import (
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// KickPattern generates kick drum events on every beat
type KickPattern struct {
	conductor    *conductors.CommonTimeConductor
	lastFireTick int64
}

// NewKickPattern creates a new kick pattern
func NewKickPattern(c *conductors.CommonTimeConductor) *KickPattern {
	return &KickPattern{
		conductor:    c,
		lastFireTick: int64(-c.GetTicksPerBeat()), // Start negative so first fire is at tick 0
	}
}

// GetNextScheduledEvent implements the Pattern interface
func (k *KickPattern) GetNextScheduledEvent() (events.ScheduledEvent, error) {
	currentTick := k.conductor.GetCurrentTick()
	ticksPerBeat := k.conductor.GetTicksPerBeat()

	// Calculate next fire tick (next beat boundary)
	nextFireTick := k.lastFireTick + int64(ticksPerBeat)

	// Convert tick delta to time delta
	tickDelta := nextFireTick - currentTick
	timeDelta := k.conductor.GetTickDuration() * time.Duration(tickDelta)

	// Update last fire tick
	k.lastFireTick = nextFireTick

	// Create kick event (MIDI note 36 = bass drum)
	return events.ScheduledEvent{
		Event: events.Event{
			Name: "kick",
			Type: events.EventTypeNote,
			A:    36.0, // MIDI note number for kick
			B:    0.8,  // Fixed velocity for now
		},
		Timing: events.Timing{
			Delta:    timeDelta,
			Duration: 100 * time.Millisecond,
		},
	}, nil
}
