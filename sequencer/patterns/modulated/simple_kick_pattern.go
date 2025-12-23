package modulated

import (
	"fmt"

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

	// Use distribution to determine if we should fire
	if !k.distribution.ShouldFire(k.tickInPhrase, k.phraseLength) {
		return nil
	}

	// Generate events based on subdivision
	var tickEvents []events.TickEvent

	// Create an event for each subdivision
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

	return tickEvents
}
