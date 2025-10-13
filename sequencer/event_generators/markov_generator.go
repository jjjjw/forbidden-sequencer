package event_generators

import "forbidden_sequencer/sequencer"

const markovSeed = 42

// MarkovGenerator implements EventGenerator using a Markov chain
type MarkovGenerator struct {
	// Name of the generator
	Name string
	// Markov chain for event generation
	Chain *sequencer.MarkovChain
}

// NewMarkovGenerator creates a new Markov generator
func NewMarkovGenerator(name string) *MarkovGenerator {
	return &MarkovGenerator{
		Name:  name,
		Chain: sequencer.NewMarkovChain(markovSeed),
	}
}

// GetNextEvent implements EventGenerator.GetNextEvent
func (g *MarkovGenerator) GetNextEvent() (sequencer.Event, error) {
	return g.Chain.Generate()
}
