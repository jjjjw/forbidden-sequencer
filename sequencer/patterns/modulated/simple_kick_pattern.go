package modulated

import (
	"fmt"
	"strings"

	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// SimpleKickPattern fires kick events using a distribution pattern
type SimpleKickPattern struct {
	distribution      lib.Distribution
	name              string  // event name
	note              uint8   // MIDI note number
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

// NewSimpleKickPattern creates a new simple kick pattern with ritardando distribution
func NewSimpleKickPattern(
	name string,
	note uint8,
	velocity float64,
	subdivision int,
	phraseLength int,
	events int,
	curve float64,
) *SimpleKickPattern {
	// Create initial distribution
	distribution := lib.NewRitardandoDistribution(events, phraseLength, curve)

	return &SimpleKickPattern{
		distribution: distribution,
		name:         name,
		note:         note,
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
func (k *SimpleKickPattern) Reset() {
	k.tickInPhrase = 0
	k.lastTick = -1
}

// Play resumes the pattern
func (k *SimpleKickPattern) Play() {
	k.paused = false
}

// Stop pauses the pattern
func (k *SimpleKickPattern) Stop() {
	k.paused = true
}

// SetSubdivision updates the subdivision value (how many times per tick to fire)
func (k *SimpleKickPattern) SetSubdivision(subdivision int) {
	k.subdivision = subdivision
}

// SetCurve updates the ritardando curve and recreates the distribution
func (k *SimpleKickPattern) SetCurve(curve float64) {
	k.curve = curve
	k.distribution = lib.NewRitardandoDistribution(k.events, k.phraseLength, curve)
}

// SetEvents updates the number of events and recreates the distribution
func (k *SimpleKickPattern) SetEvents(events int) {
	k.events = events
	k.distribution = lib.NewRitardandoDistribution(events, k.phraseLength, k.curve)
}

// SetQuantized toggles quantization mode
func (k *SimpleKickPattern) SetQuantized(quantized bool) {
	k.quantized = quantized
}

// String returns a string representation of the pattern
func (k *SimpleKickPattern) String() string {
	return fmt.Sprintf("%s (ritardando: %d events, curve=%.1f)", k.name, k.events, k.curve)
}

// updatePhrase tracks current position in phrase
func (k *SimpleKickPattern) updatePhrase(tick int64) {
	if tick != k.lastTick {
		k.tickInPhrase++
		if k.tickInPhrase >= k.phraseLength {
			k.tickInPhrase = 0
		}
		k.lastTick = tick
	}
}

// GetEventsForTick implements the Pattern interface
func (k *SimpleKickPattern) GetEventsForTick(tick int64) []events.TickEvent {
	// Update phrase position
	k.updatePhrase(tick)

	// When paused, return no events
	if k.paused {
		return nil
	}

	var tickEvents []events.TickEvent

	if k.quantized {
		// Quantized mode: use distribution to determine if we should fire
		if !k.distribution.ShouldFire(k.tickInPhrase, k.phraseLength) {
			return nil
		}

		// Generate events based on subdivision
		for i := 0; i < k.subdivision; i++ {
			// Calculate offset as percentage of tick
			offsetPercent := float64(i) / float64(k.subdivision)

			// Duration for each note (75% of subdivision duration)
			durationTicks := 0.75 / float64(k.subdivision)

			tickEvents = append(tickEvents, events.TickEvent{
				Event: events.Event{
					Name: k.name,
					Type: events.EventTypeNote,
					Params: map[string]float32{
						"midi_note": float32(k.note),
						"amp":       float32(k.velocity),
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
		if dist, ok := k.distribution.(interface{ GetContinuousPositions() []float64 }); ok {
			positions := dist.GetContinuousPositions()

			for _, pos := range positions {
				eventTick := int64(pos)
				offsetWithinTick := pos - float64(eventTick)

				// Only generate event if it falls on current tick
				if eventTick == int64(k.tickInPhrase) {
					for i := 0; i < k.subdivision; i++ {
						// Calculate offset combining sub-tick position and subdivision
						offsetPercent := offsetWithinTick + (float64(i) / float64(k.subdivision))

						// Duration for each note (75% of subdivision duration)
						durationTicks := 0.75 / float64(k.subdivision)

						tickEvents = append(tickEvents, events.TickEvent{
							Event: events.Event{
								Name: k.name,
								Type: events.EventTypeNote,
								Params: map[string]float32{
									"midi_note": float32(k.note),
									"amp":       float32(k.velocity),
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
func (k *SimpleKickPattern) Visualize() string {
	var sb strings.Builder
	for i := 0; i < k.phraseLength; i++ {
		if k.distribution.ShouldFire(i, k.phraseLength) {
			sb.WriteString("x")
		} else {
			sb.WriteString("-")
		}
	}
	return sb.String()
}

// GetPhraseLength returns the phrase length
func (k *SimpleKickPattern) GetPhraseLength() int {
	return k.phraseLength
}

// GetTickInPhrase returns the current tick within the phrase
func (k *SimpleKickPattern) GetTickInPhrase() int {
	return k.tickInPhrase
}

// GetPatternName returns the pattern name with metadata
func (k *SimpleKickPattern) GetPatternName() string {
	actualEvents := k.events
	quantMode := "non-quantized"

	if k.quantized {
		quantMode = "quantized"
		// Try to get actual event count from distribution
		if dist, ok := k.distribution.(interface{ GetActualEvents() int }); ok {
			actualEvents = dist.GetActualEvents()
		}
	}

	return fmt.Sprintf("%s (curve=%.1f, events=%d/%d, %s)",
		k.name, k.curve, actualEvents, k.events, quantMode)
}
