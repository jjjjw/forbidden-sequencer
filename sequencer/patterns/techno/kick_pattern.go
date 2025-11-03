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
	initialized  bool
}

// NewKickPattern creates a new kick pattern
func NewKickPattern(c *conductors.CommonTimeConductor) *KickPattern {
	return &KickPattern{
		conductor:   c,
		initialized: false,
	}
}

// Reset resets the pattern state
func (k *KickPattern) Reset() {
	k.initialized = false
}

// GetNextScheduledEvent implements the Pattern interface
func (k *KickPattern) GetNextScheduledEvent() (events.ScheduledEvent, error) {
	ticksPerBeat := k.conductor.GetTicksPerBeat()

	// Initialize on first call
	if !k.initialized {
		k.lastFireTick = k.conductor.GetCurrentTick()
		k.initialized = true
	}

	// Calculate next fire tick (next beat boundary)
	nextFireTick := k.lastFireTick + int64(ticksPerBeat)

	// Calculate absolute wall-clock time for this tick (drift-free)
	nextFireTime := k.conductor.GetAbsoluteTimeForTick(nextFireTick)
	timeDelta := time.Until(nextFireTime)

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
