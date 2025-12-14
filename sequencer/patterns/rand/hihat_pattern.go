package modulated

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// HihatPattern fires hihat events using a Markov chain
// States: "playing" and "silent"
// Transition probabilities:
//   - playing → playing: 0.3 (30% keep playing)
//   - playing → silent: 0.7 (70% go silent)
//   - silent → silent: 0.5 (50% stay silent)
//   - silent → playing: 0.5 (50% start playing)
// Uses MIDI note 42 (closed hihat)
// Silences after snare fires in the phrase
type HihatPattern struct {
	conductor *conductors.ModulatedRhythmConductor
	name      string           // event name
	velocity  float64          // event velocity
	paused    bool
	chain     *lib.MarkovChain // Markov chain for play/silent decisions
}

// NewHihatPattern creates a new hihat pattern
func NewHihatPattern(
	conductor *conductors.ModulatedRhythmConductor,
	name string,
	velocity float64,
) *HihatPattern {
	// Create Markov chain with two states: playing and silent
	chain := lib.NewMarkovChain(43) // Different seed from kick

	// Set transition probabilities (less busy than kick)
	// When playing: 30% keep playing, 70% go silent
	chain.SetTransitionProbability("playing", "playing", 0.3)
	chain.SetTransitionProbability("playing", "silent", 0.7)

	// When silent: 50% stay silent, 50% start playing
	chain.SetTransitionProbability("silent", "silent", 0.5)
	chain.SetTransitionProbability("silent", "playing", 0.5)

	return &HihatPattern{
		conductor: conductor,
		name:      name,
		velocity:  velocity,
		paused:    true,
		chain:     chain,
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

// String returns a string representation of the pattern
func (h *HihatPattern) String() string {
	return fmt.Sprintf("%s (markov: 30%% play, 50%% start)", h.name)
}

// GetScheduledEventsForTick implements the Pattern interface
func (h *HihatPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if h.paused {
		return nil
	}

	// Get tick position from conductor
	nextTickInPhrase := h.conductor.GetNextTickInPhrase()
	snareTriggerTick := h.conductor.GetSnareTriggerTick()

	// If snare will trigger and we're at or past the snare tick, stay silent
	if h.conductor.WillSnareTrigger() && nextTickInPhrase >= snareTriggerTick {
		return nil
	}

	// Use Markov chain to decide whether to play this tick
	state, err := h.chain.Next()
	if err != nil {
		return nil
	}

	// Only fire if we're in the "playing" state
	if state == "playing" {
		// Fire event with duration = 75% of tick
		noteDuration := time.Duration(float64(tickDuration) * 0.75)

		return []events.ScheduledEvent{{
			Event: events.Event{
				Name: h.name,
				Type: events.EventTypeNote,
				A:    0, // hihat doesn't use pitch
				B:    float32(h.velocity),
			},
			Timing: events.Timing{
				Timestamp: nextTickTime,
				Duration:  noteDuration,
			},
		}}
	}

	// Silent state - return no events
	return nil
}
