package rand

import (
	"fmt"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// KickPattern fires kick events using a Markov chain to decide when to play
// States: "playing" and "silent"
// Silences after snare fires in the phrase
type KickPattern struct {
	conductor    *conductors.Conductor
	snarePattern *SnarePattern // reference to snare pattern for trigger state
	name         string        // event name
	frequency    float32       // frequency in Hz
	velocity     float64       // event velocity
	paused       bool
	chain        *lib.MarkovChain // Markov chain for play/silent decisions
}

// NewKickPattern creates a new kick pattern
func NewKickPattern(
	conductor *conductors.Conductor,
	snarePattern *SnarePattern,
	name string,
	frequency float32,
	velocity float64,
) *KickPattern {
	// Create Markov chain with two states: playing and silent
	chain := lib.NewMarkovChain(42)

	// Set transition probabilities (using string-based state transitions)
	// When playing: 50% keep playing, 50% go silent
	chain.SetTransitionProbability("playing", "playing", 0.5)
	chain.SetTransitionProbability("playing", "silent", 0.5)

	// When silent: 50% stay silent, 50% start playing
	chain.SetTransitionProbability("silent", "silent", 0.5)
	chain.SetTransitionProbability("silent", "playing", 0.5)

	return &KickPattern{
		conductor:    conductor,
		snarePattern: snarePattern,
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

// String returns a string representation of the pattern
func (k *KickPattern) String() string {
	return fmt.Sprintf("%s (markov: 50%% play, 50%% start)", k.name)
}

// GetEventsForTick implements the Pattern interface
func (k *KickPattern) GetEventsForTick(tick int64) []events.TickEvent {
	// When paused, return no events
	if k.paused {
		return nil
	}

	// Get tick position from snare pattern
	tickInPhrase := k.snarePattern.GetCurrentTickInPhrase()
	snareTriggerTick := k.snarePattern.GetSnareTriggerTick()

	// If snare will trigger and we're at or past the snare tick, stay silent
	if k.snarePattern.WillSnareTrigger() && tickInPhrase >= snareTriggerTick {
		return nil
	}

	// Use Markov chain to decide whether to play this tick
	state, err := k.chain.Next()
	if err != nil {
		return nil
	}

	// Only fire if we're in the "playing" state
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
