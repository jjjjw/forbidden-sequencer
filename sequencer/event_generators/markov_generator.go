package event_generators

import (
	"fmt"

	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

const markovSeed = 42

// MarkovGenerator implements EventGenerator using a Markov chain
// Maps state names to events for generation
type MarkovGenerator struct {
	// Name of the generator
	Name string
	// Markov chain for state transitions
	Chain *lib.MarkovChain
	// Event mapping from state names to events
	stateEvents map[string]events.Event
}

// NewMarkovGenerator creates a new Markov generator
func NewMarkovGenerator(name string) *MarkovGenerator {
	return &MarkovGenerator{
		Name:        name,
		Chain:       lib.NewMarkovChain(markovSeed),
		stateEvents: make(map[string]events.Event),
	}
}

// AddStateEvent maps a state name to an event
func (g *MarkovGenerator) AddStateEvent(stateName string, event events.Event) {
	g.stateEvents[stateName] = event
}

// GetNextEvent implements EventGenerator.GetNextEvent
func (g *MarkovGenerator) GetNextEvent() (events.Event, error) {
	// Get next state from chain
	state, err := g.Chain.Next()
	if err != nil {
		return events.Event{}, err
	}

	// Look up corresponding event
	event, exists := g.stateEvents[state]
	if !exists {
		return events.Event{}, fmt.Errorf("no event mapped to state %s", state)
	}

	return event, nil
}
