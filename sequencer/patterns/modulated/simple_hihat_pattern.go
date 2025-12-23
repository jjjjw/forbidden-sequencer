package modulated

import (
	"fmt"

	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// SimpleHihatPattern fires hihat events using a distribution pattern
type SimpleHihatPattern struct {
	distribution      lib.Distribution
	name              string  // event name
	velocity          float64 // event velocity
	subdivision       int     // number of times to fire per tick
	paused            bool
	phraseLength      int     // length of phrase in ticks
	tickInPhrase      int     // current tick within phrase
	lastTick          int64   // last tick we saw
	events            int     // number of events in distribution
	curve             float64 // curve parameter for distribution
}

// NewSimpleHihatPattern creates a new simple hihat pattern with ritardando distribution
func NewSimpleHihatPattern(
	name string,
	velocity float64,
	subdivision int,
	phraseLength int,
	events int,
	curve float64,
) *SimpleHihatPattern {
	// Create initial distribution
	distribution := lib.NewRitardandoDistribution(events, phraseLength, curve)

	return &SimpleHihatPattern{
		distribution: distribution,
		name:         name,
		velocity:     velocity,
		subdivision:  subdivision,
		paused:       true,
		phraseLength: phraseLength,
		tickInPhrase: 0,
		lastTick:     -1,
		events:       events,
		curve:        curve,
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

// SetCurve updates the ritardando curve and recreates the distribution
func (h *SimpleHihatPattern) SetCurve(curve float64) {
	h.curve = curve
	h.distribution = lib.NewRitardandoDistribution(h.events, h.phraseLength, curve)
}

// SetEvents updates the number of events and recreates the distribution
func (h *SimpleHihatPattern) SetEvents(events int) {
	h.events = events
	h.distribution = lib.NewRitardandoDistribution(events, h.phraseLength, h.curve)
}

// String returns a string representation of the pattern
func (h *SimpleHihatPattern) String() string {
	return fmt.Sprintf("%s (ritardando: %d events, curve=%.1f)", h.name, h.events, h.curve)
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

	// Use distribution to determine if we should fire
	if !h.distribution.ShouldFire(h.tickInPhrase, h.phraseLength) {
		return nil
	}

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
