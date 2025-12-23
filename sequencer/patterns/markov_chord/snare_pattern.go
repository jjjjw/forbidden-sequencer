package markov_chord

import (
	"fmt"

	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// SnarePattern fires snare events using Markov chain during percussion sections
type SnarePattern struct {
	chordPattern *ChordPattern
	name         string
	velocity     float64
	paused       bool
	chain        *lib.MarkovChain
}

// NewSnarePattern creates a new snare pattern
func NewSnarePattern(
	chordPattern *ChordPattern,
	name string,
	velocity float64,
) *SnarePattern {
	// Create Markov chain
	chain := lib.NewMarkovChain(84)

	// When playing: 40% keep playing, 60% go silent
	chain.SetTransitionProbability("playing", "playing", 0.4)
	chain.SetTransitionProbability("playing", "silent", 0.6)

	// When silent: 70% stay silent, 30% start playing
	chain.SetTransitionProbability("silent", "silent", 0.7)
	chain.SetTransitionProbability("silent", "playing", 0.3)

	return &SnarePattern{
		chordPattern: chordPattern,
		name:         name,
		velocity:     velocity,
		paused:       true,
		chain:        chain,
	}
}

// Reset resets the pattern state
func (s *SnarePattern) Reset() {
	s.chain.Reset()
}

// Play resumes the pattern
func (s *SnarePattern) Play() {
	s.paused = false
}

// Stop pauses the pattern
func (s *SnarePattern) Stop() {
	s.paused = true
}

// String returns a string representation
func (s *SnarePattern) String() string {
	return fmt.Sprintf("%s (markov)", s.name)
}

// GetEventsForTick implements the Pattern interface
func (s *SnarePattern) GetEventsForTick(tick int64) []events.TickEvent {
	// When paused, return no events
	if s.paused {
		return nil
	}

	// Only play during percussion section
	if !s.chordPattern.IsPercussionSection() {
		return nil
	}

	// Use Markov chain to decide whether to play
	state, err := s.chain.Next()
	if err != nil {
		return nil
	}

	// Only fire if in "playing" state
	if state == "playing" {
		return []events.TickEvent{{
			Event: events.Event{
				Name: s.name,
				Type: events.EventTypeNote,
				Params: map[string]float32{
					"amp": float32(s.velocity),
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
