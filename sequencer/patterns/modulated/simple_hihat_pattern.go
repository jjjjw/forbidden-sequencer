package modulated

import (
	"fmt"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// SimpleHihatPattern fires hihat events in the first half of the phrase (0-50%)
type SimpleHihatPattern struct {
	conductor         *conductors.Conductor
	name              string  // event name
	velocity          float64 // event velocity
	subdivision       int     // number of times to fire per tick
	paused            bool
	phraseLength      int   // length of phrase in ticks
	tickInPhrase      int   // current tick within phrase
	lastTick          int64 // last tick we saw
}

// NewSimpleHihatPattern creates a new simple hihat pattern
func NewSimpleHihatPattern(
	conductor *conductors.Conductor,
	name string,
	velocity float64,
	subdivision int,
	phraseLength int,
) *SimpleHihatPattern {
	return &SimpleHihatPattern{
		conductor:    conductor,
		name:         name,
		velocity:     velocity,
		subdivision:  subdivision,
		paused:       true,
		phraseLength: phraseLength,
		tickInPhrase: 0,
		lastTick:     -1,
	}
}

// Reset resets the pattern state
func (h *SimpleHihatPattern) Reset() {
	h.tickInPhrase = 0
	h.lastTick = -1
}

// Play resumes the pattern
func (h *SimpleHihatPattern) Play() {
	h.paused = false
}

// Stop pauses the pattern
func (h *SimpleHihatPattern) Stop() {
	h.paused = true
}

// SetSubdivision updates the subdivision value (how many times per tick to fire)
func (h *SimpleHihatPattern) SetSubdivision(subdivision int) {
	h.subdivision = subdivision
}

// String returns a string representation of the pattern
func (h *SimpleHihatPattern) String() string {
	return fmt.Sprintf("%s (0-50%% of phrase)", h.name)
}

// updatePhrase tracks current position in phrase
func (h *SimpleHihatPattern) updatePhrase(tick int64) {
	if tick != h.lastTick {
		h.tickInPhrase++
		if h.tickInPhrase >= h.phraseLength {
			h.tickInPhrase = 0
		}
		h.lastTick = tick
	}
}

// GetEventsForTick implements the Pattern interface
func (h *SimpleHihatPattern) GetEventsForTick(tick int64) []events.TickEvent {
	// Update phrase position
	h.updatePhrase(tick)

	// When paused, return no events
	if h.paused {
		return nil
	}

	// Fire if in first half of phrase (0-50%)
	if float64(h.tickInPhrase) < float64(h.phraseLength)*0.5 {
		// Always use closed hihat (MIDI note 42)
		note := uint8(42)

		// Generate events based on subdivision
		var tickEvents []events.TickEvent

		// Create an event for each subdivision
		for i := 0; i < h.subdivision; i++ {
			// Calculate offset as percentage of tick
			offsetPercent := float64(i) / float64(h.subdivision)

			// Duration for each note (75% of subdivision duration)
			durationTicks := 0.75 / float64(h.subdivision)

			tickEvents = append(tickEvents, events.TickEvent{
				Event: events.Event{
					Name: h.name,
					Type: events.EventTypeNote,
					Params: map[string]float32{
						"midi_note": float32(note),
						"amp":       float32(h.velocity),
					},
				},
				TickTiming: events.TickTiming{
					Tick:          tick,
				OffsetPercent: offsetPercent,
				DurationTicks: durationTicks,
				},
			})
		}

		return tickEvents
	}

	// Outside range - return no events
	return nil
}
