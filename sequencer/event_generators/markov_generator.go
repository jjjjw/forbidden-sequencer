package event_generators

import (
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

const markovSeed = 42

// MarkovGenerator implements EventGenerator using a Markov chain
type MarkovGenerator struct {
	// Name of the generator
	Name string
	// Markov chain for event generation
	Chain *lib.MarkovChain
}

// NewMarkovGenerator creates a new Markov generator
func NewMarkovGenerator(name string) *MarkovGenerator {
	return &MarkovGenerator{
		Name:  name,
		Chain: lib.NewMarkovChain(markovSeed),
	}
}

// GetNextEvent implements EventGenerator.GetNextEvent
func (g *MarkovGenerator) GetNextEvent() (events.Event, error) {
	return g.Chain.Generate()
}
