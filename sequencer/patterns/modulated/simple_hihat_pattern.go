package modulated

import (
	"fmt"
	"strings"

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
	quantized         bool    // whether to quantize events to ticks
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
		quantized:    true, // default to quantized mode
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

// SetQuantized toggles quantization mode
func (h *SimpleHihatPattern) SetQuantized(quantized bool) {
	h.quantized = quantized
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

	// Always use closed hihat (MIDI note 42)
	note := uint8(42)

	var tickEvents []events.TickEvent

	if h.quantized {
		// Quantized mode: use distribution to determine if we should fire
		if !h.distribution.ShouldFire(h.tickInPhrase, h.phraseLength) {
			return nil
		}

		// Generate events based on subdivision
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
	} else {
		// Non-quantized mode: generate events with sub-tick offsets
		// Get continuous positions from distribution
		if dist, ok := h.distribution.(interface{ GetContinuousPositions() []float64 }); ok {
			positions := dist.GetContinuousPositions()

			for _, pos := range positions {
				eventTick := int64(pos)
				offsetWithinTick := pos - float64(eventTick)

				// Only generate event if it falls on current tick
				if eventTick == int64(h.tickInPhrase) {
					for i := 0; i < h.subdivision; i++ {
						// Calculate offset combining sub-tick position and subdivision
						offsetPercent := offsetWithinTick + (float64(i) / float64(h.subdivision))

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
				}
			}
		}
	}

	return tickEvents
}

// Visualize returns a visual representation of the pattern
func (h *SimpleHihatPattern) Visualize() string {
	var sb strings.Builder
	for i := 0; i < h.phraseLength; i++ {
		if h.distribution.ShouldFire(i, h.phraseLength) {
			sb.WriteString("x")
		} else {
			sb.WriteString("-")
		}
	}
	return sb.String()
}

// GetPhraseLength returns the phrase length
func (h *SimpleHihatPattern) GetPhraseLength() int {
	return h.phraseLength
}

// GetTickInPhrase returns the current tick within the phrase
func (h *SimpleHihatPattern) GetTickInPhrase() int {
	return h.tickInPhrase
}

// GetPatternName returns the pattern name with metadata
func (h *SimpleHihatPattern) GetPatternName() string {
	actualEvents := h.events
	quantMode := "non-quantized"

	if h.quantized {
		quantMode = "quantized"
		// Try to get actual event count from distribution
		if dist, ok := h.distribution.(interface{ GetActualEvents() int }); ok {
			actualEvents = dist.GetActualEvents()
		}
	}

	return fmt.Sprintf("%s (curve=%.1f, events=%d/%d, %s)",
		h.name, h.curve, actualEvents, h.events, quantMode)
}
