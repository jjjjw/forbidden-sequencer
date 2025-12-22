package rand

import (
	"fmt"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// SnarePattern fires a snare event at 3/4 of the phrase, with Markov chain deciding whether to trigger
// Other patterns can check if snare will trigger to adjust their behavior
type SnarePattern struct {
	conductor      *conductors.Conductor
	name           string
	velocity       float64
	paused         bool
	phraseLength   int              // length of phrase in ticks
	tickInPhrase   int              // current tick within phrase
	lastTick       int64            // last tick we saw
	chain          *lib.MarkovChain // Markov chain for trigger decisions
	willTrigger    bool             // whether snare will trigger this phrase
	triggerChecked bool             // whether we've decided for this phrase
}

// NewSnarePattern creates a new snare pattern
func NewSnarePattern(
	conductor *conductors.Conductor,
	name string,
	velocity float64,
	phraseLength int,
) *SnarePattern {
	// Create Markov chain for trigger decision
	chain := lib.NewMarkovChain(55)

	// When triggered: 60% trigger again, 40% don't trigger
	chain.SetTransitionProbability("trigger", "trigger", 0.6)
	chain.SetTransitionProbability("trigger", "silent", 0.4)

	// When silent: 40% stay silent, 60% trigger
	chain.SetTransitionProbability("silent", "silent", 0.4)
	chain.SetTransitionProbability("silent", "trigger", 0.6)

	return &SnarePattern{
		conductor:      conductor,
		name:           name,
		velocity:       velocity,
		paused:         true,
		phraseLength:   phraseLength,
		tickInPhrase:   0,
		lastTick:       -1,
		chain:          chain,
		willTrigger:    false,
		triggerChecked: false,
	}
}

// Reset resets the pattern state
func (s *SnarePattern) Reset() {
	s.tickInPhrase = 0
	s.lastTick = -1
	s.chain.Reset()
	s.willTrigger = false
	s.triggerChecked = false
}

// Play resumes the pattern
func (s *SnarePattern) Play() {
	s.paused = false
}

// Stop pauses the pattern
func (s *SnarePattern) Stop() {
	s.paused = true
}

// String returns a string representation of the pattern
func (s *SnarePattern) String() string {
	return fmt.Sprintf("%s (conditional at 3/4)", s.name)
}

// updatePhrase tracks current position in phrase and decides trigger for new phrases
func (s *SnarePattern) updatePhrase(tick int64) {
	if tick != s.lastTick {
		s.tickInPhrase++
		if s.tickInPhrase >= s.phraseLength {
			s.tickInPhrase = 0
			s.triggerChecked = false // Reset for new phrase
		}
		s.lastTick = tick
	}

	// Decide whether to trigger at start of each phrase
	if !s.triggerChecked {
		state, _ := s.chain.Next()
		s.willTrigger = (state == "trigger")
		s.triggerChecked = true
	}
}

// GetSnareTriggerTick returns the tick in phrase where snare triggers (3/4 of phrase)
func (s *SnarePattern) GetSnareTriggerTick() int {
	return (s.phraseLength * 3) / 4
}

// WillSnareTrigger returns whether snare will trigger this phrase
func (s *SnarePattern) WillSnareTrigger() bool {
	return s.willTrigger
}

// GetCurrentTickInPhrase returns current tick position in phrase
func (s *SnarePattern) GetCurrentTickInPhrase() int {
	return s.tickInPhrase
}

// GetEventsForTick implements the Pattern interface
func (s *SnarePattern) GetEventsForTick(tick int64) []events.TickEvent {
	// Update phrase position and trigger decision
	s.updatePhrase(tick)

	// When paused, return no events
	if s.paused {
		return nil
	}

	snareTriggerTick := s.GetSnareTriggerTick()

	// Check if this is the snare trigger tick AND we decided to trigger
	if s.tickInPhrase == snareTriggerTick && s.willTrigger {
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

	// Don't trigger - return no events
	return nil
}
