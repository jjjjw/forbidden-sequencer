package techno

import (
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// HihatPattern generates hihat events on every off-beat (half-beat)
type HihatPattern struct {
	conductor    *conductors.CommonTimeConductor
	lastFireTick int64
	initialized  bool
}

// NewHihatPattern creates a new hihat pattern
func NewHihatPattern(c *conductors.CommonTimeConductor) *HihatPattern {
	return &HihatPattern{
		conductor:    c,
		lastFireTick: 0,
		initialized:  false,
	}
}

// Reset resets the pattern state
func (h *HihatPattern) Reset() {
	h.initialized = false
}

// GetNextScheduledEvent implements the Pattern interface
func (h *HihatPattern) GetNextScheduledEvent() (events.ScheduledEvent, error) {
	ticksPerBeat := h.conductor.GetTicksPerBeat()

	// Initialize on first call, offset by half a beat to maintain off-beat rhythm
	if !h.initialized {
		h.lastFireTick = h.conductor.GetCurrentTick() + int64(ticksPerBeat/2)
		h.initialized = true
	}

	// Calculate next fire tick (next off-beat)
	nextFireTick := h.lastFireTick + int64(ticksPerBeat)

	// Calculate absolute wall-clock time for this tick (drift-free)
	nextFireTime := h.conductor.GetAbsoluteTimeForTick(nextFireTick)
	timeDelta := time.Until(nextFireTime)

	// Update last fire tick
	h.lastFireTick = nextFireTick

	// Create hihat event (MIDI note 42 = closed hihat)
	return events.ScheduledEvent{
		Event: events.Event{
			Name: "hihat",
			Type: events.EventTypeNote,
			A:    42.0, // MIDI note number for closed hihat
			B:    0.6,  // Fixed velocity for now
		},
		Timing: events.Timing{
			Delta:    timeDelta,
			Duration: 50 * time.Millisecond,
		},
	}, nil
}
