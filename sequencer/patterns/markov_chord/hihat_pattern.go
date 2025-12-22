package markov_chord

import (
	"fmt"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// HihatPattern fires hihat events using Markov chain during percussion sections
type HihatPattern struct {
	conductor    *conductors.Conductor
	chordPattern *ChordPattern
	name         string
	velocity     float64
	paused       bool
	chain        *lib.MarkovChain
}

// NewHihatPattern creates a new hihat pattern
func NewHihatPattern(
	conductor *conductors.Conductor,
	chordPattern *ChordPattern,
	name string,
	velocity float64,
) *HihatPattern {
	// Create Markov chain
	chain := lib.NewMarkovChain(126)

	// When playing: 80% keep playing, 20% go silent (more consistent)
	chain.SetTransitionProbability("playing", "playing", 0.8)
	chain.SetTransitionProbability("playing", "silent", 0.2)

	// When silent: 30% stay silent, 70% start playing
	chain.SetTransitionProbability("silent", "silent", 0.3)
	chain.SetTransitionProbability("silent", "playing", 0.7)

	return &HihatPattern{
		conductor:    conductor,
		chordPattern: chordPattern,
		name:         name,
		velocity:     velocity,
		paused:       true,
		chain:        chain,
	}
}

// Reset resets the pattern state
func (h *HihatPattern) Reset() {
	h.chain.Reset()
}

// Play resumes the pattern
func (h *HihatPattern) Play() {
	h.paused = false
}

// Stop pauses the pattern
func (h *HihatPattern) Stop() {
	h.paused = true
}

// String returns a string representation
func (h *HihatPattern) String() string {
	return fmt.Sprintf("%s (markov)", h.name)
}

// GetEventsForTick implements the Pattern interface
func (h *HihatPattern) GetEventsForTick(tick int64) []events.TickEvent {
	// When paused, return no events
	if h.paused {
		return nil
	}

	// Only play during percussion section
	if !h.chordPattern.IsPercussionSection() {
		return nil
	}

	// Use Markov chain to decide whether to play
	state, err := h.chain.Next()
	if err != nil {
		return nil
	}

	// Only fire if in "playing" state
	if state == "playing" {
		return []events.TickEvent{{
			Event: events.Event{
				Name: h.name,
				Type: events.EventTypeNote,
				Params: map[string]float32{
					"amp": float32(h.velocity),
				},
			},
			TickTiming: events.TickTiming{
				Tick:          tick,
				OffsetPercent: 0.0,
				DurationTicks: 0.5,
			}, // Shorter hihat
		}}
	}

	return nil
}
