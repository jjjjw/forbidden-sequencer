package modulated

import (
	"fmt"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// SimpleKickPattern fires kick events in the first half of the phrase (0-50%)
type SimpleKickPattern struct {
	conductor         *conductors.Conductor
	name              string  // event name
	note              uint8   // MIDI note number
	velocity          float64 // event velocity
	subdivision       int     // number of times to fire per tick
	paused            bool
	phraseLength      int   // length of phrase in ticks
	tickInPhrase      int   // current tick within phrase
	lastTick          int64 // last tick we saw
}

// NewSimpleKickPattern creates a new simple kick pattern
func NewSimpleKickPattern(
	conductor *conductors.Conductor,
	name string,
	note uint8,
	velocity float64,
	subdivision int,
	phraseLength int,
) *SimpleKickPattern {
	return &SimpleKickPattern{
		conductor:    conductor,
		name:         name,
		note:         note,
		velocity:     velocity,
		subdivision:  subdivision,
		paused:       true,
		phraseLength: phraseLength,
		tickInPhrase: 0,
		lastTick:     -1,
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

// String returns a string representation of the pattern
func (k *SimpleKickPattern) String() string {
	return fmt.Sprintf("%s (0-50%% of phrase)", k.name)
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

	// Fire if in first half of phrase (0-50%)
	if float64(k.tickInPhrase) < float64(k.phraseLength)*0.5 {
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
				Tick:          tick,
				OffsetPercent: offsetPercent,
				DurationTicks: durationTicks,
			})
		}

		return tickEvents
	}

	// Outside range - return no events
	return nil
}
