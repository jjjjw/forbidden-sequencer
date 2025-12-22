package markov_chord

import (
	"fmt"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// KickPattern fires kick events using Markov chain during percussion sections
type KickPattern struct {
	conductor    *conductors.Conductor
	chordPattern *ChordPattern // reference to chord pattern for section state
	name         string
	frequency    float32
	velocity     float64
	paused       bool
	chain        *lib.MarkovChain
}

// NewKickPattern creates a new kick pattern
func NewKickPattern(
	conductor *conductors.Conductor,
	chordPattern *ChordPattern,
	name string,
	frequency float32,
	velocity float64,
) *KickPattern {
	// Create Markov chain: playing vs silent
	chain := lib.NewMarkovChain(42)

	// When playing: 60% keep playing, 40% go silent
	chain.SetTransitionProbability("playing", "playing", 0.6)
	chain.SetTransitionProbability("playing", "silent", 0.4)

	// When silent: 40% stay silent, 60% start playing
	chain.SetTransitionProbability("silent", "silent", 0.4)
	chain.SetTransitionProbability("silent", "playing", 0.6)

	return &KickPattern{
		conductor:    conductor,
		chordPattern: chordPattern,
		name:         name,
		frequency:    frequency,
		velocity:     velocity,
		paused:       true,
		chain:        chain,
	}
}

// Reset resets the pattern state
func (k *KickPattern) Reset() {
	k.chain.Reset()
}

// Play resumes the pattern
func (k *KickPattern) Play() {
	k.paused = false
}

// Stop pauses the pattern
func (k *KickPattern) Stop() {
	k.paused = true
}

// String returns a string representation
func (k *KickPattern) String() string {
	return fmt.Sprintf("%s (markov)", k.name)
}

// GetEventsForTick implements the Pattern interface
func (k *KickPattern) GetEventsForTick(tick int64) []events.TickEvent {
	// When paused, return no events
	if k.paused {
		return nil
	}

	// Only play during percussion section
	if !k.chordPattern.IsPercussionSection() {
		return nil
	}

	// Use Markov chain to decide whether to play
	state, err := k.chain.Next()
	if err != nil {
		return nil
	}

	// Only fire if in "playing" state
	if state == "playing" {
		return []events.TickEvent{{
			Event: events.Event{
				Name: k.name,
				Type: events.EventTypeNote,
				Params: map[string]float32{
					"freq": k.frequency,
					"amp":  float32(k.velocity),
				},
			},
			TickTiming: events.TickTiming{
				Tick:          tick,
				OffsetPercent: 0.0,
				DurationTicks: 0.75,
			},
		}}
	}

	return nil
}
