package lib

import (
	"fmt"
	"math/rand"
)

// TODO: Support second-order or n-order Markov chains
// (where next state depends on last n states, not just last 1)

// MarkovChain represents a first-order Markov chain for state transitions
type MarkovChain struct {
	// Transition probabilities: transitions[fromState][toState] = probability
	transitions map[string]map[string]float32
	// Current state
	currentState string
	// Random number generator
	rng *rand.Rand
}

// NewMarkovChain creates a new Markov chain
func NewMarkovChain(seed int64) *MarkovChain {
	return &MarkovChain{
		transitions: make(map[string]map[string]float32),
		rng:         rand.New(rand.NewSource(seed)),
	}
}

// SetTransitionProbability sets the probability of transitioning from one state to another
// probability should be between 0.0 and 1.0
func (m *MarkovChain) SetTransitionProbability(fromState, toState string, probability float32) error {
	if probability < 0.0 || probability > 1.0 {
		return fmt.Errorf("probability must be between 0.0 and 1.0, got %f", probability)
	}

	if m.transitions[fromState] == nil {
		m.transitions[fromState] = make(map[string]float32)
	}

	m.transitions[fromState][toState] = probability
	return nil
}

// GetTransitionProbability returns the probability of transitioning from one state to another
func (m *MarkovChain) GetTransitionProbability(fromState, toState string) float32 {
	if m.transitions[fromState] == nil {
		return 0.0
	}
	return m.transitions[fromState][toState]
}

// NormalizeTransitions normalizes all transition probabilities for a given state
// so they sum to 1.0
func (m *MarkovChain) NormalizeTransitions(fromState string) {
	if m.transitions[fromState] == nil {
		return
	}

	var sum float32
	for _, prob := range m.transitions[fromState] {
		sum += prob
	}

	if sum > 0 {
		for toState := range m.transitions[fromState] {
			m.transitions[fromState][toState] /= sum
		}
	}
}

// Next generates the next state based on current state and transition probabilities
func (m *MarkovChain) Next() (string, error) {
	// If no current state, pick a random state to start
	if m.currentState == "" {
		if len(m.transitions) == 0 {
			return "", fmt.Errorf("no states in chain")
		}

		// Pick random starting state
		i := m.rng.Intn(len(m.transitions))
		for state := range m.transitions {
			if i == 0 {
				m.currentState = state
				break
			}
			i--
		}
	}

	// Get transitions from current state
	trans := m.transitions[m.currentState]
	if len(trans) == 0 {
		// No transitions defined, stay on current state
		return m.currentState, nil
	}

	// Select next state based on probabilities
	r := m.rng.Float32()
	var cumulative float32

	for toState, prob := range trans {
		cumulative += prob
		if r <= cumulative {
			m.currentState = toState
			return m.currentState, nil
		}
	}

	// Fallback to current state if probabilities don't sum to 1
	return m.currentState, nil
}

// SetCurrentState manually sets the current state of the chain
func (m *MarkovChain) SetCurrentState(stateName string) error {
	if _, exists := m.transitions[stateName]; !exists {
		return fmt.Errorf("state %s does not exist in chain", stateName)
	}
	m.currentState = stateName
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
