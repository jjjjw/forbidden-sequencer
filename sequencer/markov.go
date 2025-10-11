package sequencer

import (
	"fmt"
	"math/rand"
)

// TODO: Support second-order or n-order Markov chains
// (where next event depends on last n events, not just last 1)

// MarkovChain represents a first-order Markov chain for generating events
type MarkovChain struct {
	// Transition probabilities: transitions[fromEvent][toEvent] = probability
	transitions map[string]map[string]float32
	// Event pool for the chain
	events map[string]Event
	// Current state (last generated event name)
	currentState string
	// Random number generator
	rng *rand.Rand
}

// NewMarkovChain creates a new Markov chain
func NewMarkovChain(seed int64) *MarkovChain {
	return &MarkovChain{
		transitions: make(map[string]map[string]float32),
		events:      make(map[string]Event),
		rng:         rand.New(rand.NewSource(seed)),
	}
}

// AddEvent adds an event to the chain's event pool
func (m *MarkovChain) AddEvent(event Event) {
	m.events[event.Name] = event
	// Initialize transitions map for this event if it doesn't exist
	if m.transitions[event.Name] == nil {
		m.transitions[event.Name] = make(map[string]float32)
	}
}

// SetTransitionProbability sets the probability of transitioning from one event to another
// probability should be between 0.0 and 1.0
func (m *MarkovChain) SetTransitionProbability(fromEvent, toEvent string, probability float32) error {
	if probability < 0.0 || probability > 1.0 {
		return fmt.Errorf("probability must be between 0.0 and 1.0, got %f", probability)
	}

	if m.transitions[fromEvent] == nil {
		m.transitions[fromEvent] = make(map[string]float32)
	}

	m.transitions[fromEvent][toEvent] = probability
	return nil
}

// GetTransitionProbability returns the probability of transitioning from one event to another
func (m *MarkovChain) GetTransitionProbability(fromEvent, toEvent string) float32 {
	if m.transitions[fromEvent] == nil {
		return 0.0
	}
	return m.transitions[fromEvent][toEvent]
}

// NormalizeTransitions normalizes all transition probabilities for a given event
// so they sum to 1.0
func (m *MarkovChain) NormalizeTransitions(fromEvent string) {
	if m.transitions[fromEvent] == nil {
		return
	}

	var sum float32
	for _, prob := range m.transitions[fromEvent] {
		sum += prob
	}

	if sum > 0 {
		for toEvent := range m.transitions[fromEvent] {
			m.transitions[fromEvent][toEvent] /= sum
		}
	}
}

// Generate generates the next event based on current state and transition probabilities
func (m *MarkovChain) Generate() (Event, error) {
	// If no current state, pick a random event to start
	if m.currentState == "" {
		if len(m.events) == 0 {
			return Event{}, fmt.Errorf("no events in chain")
		}

		// Pick random starting event
		i := m.rng.Intn(len(m.events))
		for name := range m.events {
			if i == 0 {
				m.currentState = name
				break
			}
			i--
		}
	}

	// Get transitions from current state
	trans := m.transitions[m.currentState]
	if len(trans) == 0 {
		// No transitions defined, stay on current event
		return m.events[m.currentState], nil
	}

	// Select next event based on probabilities
	r := m.rng.Float32()
	var cumulative float32

	for toEvent, prob := range trans {
		cumulative += prob
		if r <= cumulative {
			m.currentState = toEvent
			return m.events[toEvent], nil
		}
	}

	// Fallback to current event if probabilities don't sum to 1
	return m.events[m.currentState], nil
}

// SetCurrentState manually sets the current state of the chain
func (m *MarkovChain) SetCurrentState(eventName string) error {
	if _, exists := m.events[eventName]; !exists {
		return fmt.Errorf("event %s does not exist in chain", eventName)
	}
	m.currentState = eventName
	return nil
}

// GetCurrentState returns the current state of the chain
func (m *MarkovChain) GetCurrentState() string {
	return m.currentState
}

// Reset resets the chain to no current state
func (m *MarkovChain) Reset() {
	m.currentState = ""
}
