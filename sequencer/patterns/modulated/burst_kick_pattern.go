package modulated

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// BurstKickPattern fires kick events using a Markov chain to decide when to play
// States: "playing" and "silent"
// Transition probabilities:
//   - playing → playing: 0.5 (50% keep playing)
//   - playing → silent: 0.5 (50% go silent)
//   - silent → silent: 0.4 (40% stay silent)
//   - silent → playing: 0.6 (60% start playing)
//
// Silences after snare fires in the phrase
type BurstKickPattern struct {
	conductor       *conductors.PhraseConductor
	rhythmConductor *conductors.ModulatedRhythmConductor
	name            string  // event name
	note            uint8   // MIDI note number
	velocity        float64 // event velocity
	paused          bool
	chain           *lib.MarkovChain // Markov chain for play/silent decisions
}

// NewBurstKickPattern creates a new burst kick pattern
func NewBurstKickPattern(
	conductor *conductors.PhraseConductor,
	rhythmConductor *conductors.ModulatedRhythmConductor,
	name string,
	note uint8,
	velocity float64,
) *BurstKickPattern {
	// Create Markov chain with two states: playing and silent
	chain := lib.NewMarkovChain(42)

	// Set transition probabilities (using string-based state transitions)
	// When playing: 50% keep playing, 50% go silent
	chain.SetTransitionProbability("playing", "playing", 0.5)
	chain.SetTransitionProbability("playing", "silent", 0.5)

	// When silent: 40% stay silent, 60% start playing
	chain.SetTransitionProbability("silent", "silent", 0.5)
	chain.SetTransitionProbability("silent", "playing", 0.5)

	return &BurstKickPattern{
		conductor:       conductor,
		rhythmConductor: rhythmConductor,
		name:            name,
		note:            note,
		velocity:        velocity,
		paused:          true,
		chain:           chain,
	}
}

// Reset resets the pattern state
func (b *BurstKickPattern) Reset() {
	b.chain.Reset()
}

// Play resumes the pattern
func (b *BurstKickPattern) Play() {
	b.paused = false
}

// Stop pauses the pattern
func (b *BurstKickPattern) Stop() {
	b.paused = true
}

// String returns a string representation of the pattern
func (b *BurstKickPattern) String() string {
	return fmt.Sprintf("%s (markov: 50%% play, 60%% start)", b.name)
}

// GetScheduledEventsForTick implements the Pattern interface
func (b *BurstKickPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if b.paused {
		return nil
	}

	// Get tick position from conductor
	nextTickInPhrase := b.conductor.GetNextTickInPhrase()
	snareTriggerTick := b.rhythmConductor.GetSnareTriggerTick()

	// If snare will trigger and we're at or past the snare tick, stay silent
	if b.rhythmConductor.WillSnareTrigger() && nextTickInPhrase >= snareTriggerTick {
		return nil
	}

	// Use Markov chain to decide whether to play this tick
	state, err := b.chain.Next()
	if err != nil {
		return nil
	}

	// Only fire if we're in the "playing" state
	if state == "playing" {
		// Fire event with duration = 75% of tick
		noteDuration := time.Duration(float64(tickDuration) * 0.75)
		return []events.ScheduledEvent{{
			Event: events.Event{
				Name: b.name,
				Type: events.EventTypeNote,
				A:    float32(b.note),
				B:    float32(b.velocity),
			},
			Timing: events.Timing{
				Timestamp: nextTickTime,
				Duration:  noteDuration,
			},
		}}
	}

	return nil
}
