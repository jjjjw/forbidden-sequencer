package techno

import (
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// TechnoPattern generates kick and hihat events on beats
type TechnoPattern struct {
	conductor    *conductors.Conductor
	paused       bool
	ticksPerBeat int   // number of ticks per beat
	tickInBeat   int   // current tick within beat
	lastTick     int64 // last tick we saw
}

// NewTechnoPattern creates a new techno pattern
func NewTechnoPattern(conductor *conductors.Conductor, ticksPerBeat int) *TechnoPattern {
	return &TechnoPattern{
		conductor:    conductor,
		paused:       true,
		ticksPerBeat: ticksPerBeat,
		tickInBeat:   0,
		lastTick:     -1,
	}
}

// Reset resets the pattern state
func (t *TechnoPattern) Reset() {
	t.tickInBeat = 0
	t.lastTick = -1
}

// Play resumes the pattern
func (t *TechnoPattern) Play() {
	t.paused = false
}

// Stop pauses the pattern
func (t *TechnoPattern) Stop() {
	t.paused = true
}

// updateBeat tracks current position in beat
func (t *TechnoPattern) updateBeat(tick int64) {
	if tick != t.lastTick {
		t.tickInBeat++
		if t.tickInBeat >= t.ticksPerBeat {
			t.tickInBeat = 0
		}
		t.lastTick = tick
	}
}

// GetEventsForTick implements the Pattern interface
func (t *TechnoPattern) GetEventsForTick(tick int64) []events.TickEvent {
	// Update beat position
	t.updateBeat(tick)

	// When paused, return no events
	if t.paused {
		return nil
	}

	// Only schedule on beat boundaries (tick 0 in beat)
	if t.tickInBeat != 0 {
		return nil
	}

	// Schedule both kick and hihat for this beat
	return []events.TickEvent{
		{
			// Kick on the beat
			Event: events.Event{
				Name: "kick",
				Type: events.EventTypeNote,
				Params: map[string]float32{
					"midi_note": 36,
					"amp":       0.8,
				},
			},
			TickTiming: events.TickTiming{
				Tick:          tick,
				OffsetPercent: 0.0, // On the beat
				DurationTicks: 0.1, // Short kick
			},
		},
		{
			// Hihat on the offbeat (halfway through beat)
			Event: events.Event{
				Name: "hihat",
				Type: events.EventTypeNote,
				Params: map[string]float32{
					"midi_note": 42,
					"amp":       0.6,
				},
			},
			TickTiming: events.TickTiming{
				Tick:          tick,
				OffsetPercent: float64(t.ticksPerBeat) / 2.0, // Halfway to next beat
				DurationTicks: 0.05,                          // Shorter hihat
			},
		},
	}
}
