package lanes

import "forbidden_sequencer/sequencer"

const markovSeed = 42

// MarkovLane implements Lane using a Markov chain
type MarkovLane struct {
	// Name of the lane
	Name string
	// Markov chain for event generation
	Chain *sequencer.MarkovChain
}

// NewMarkovLane creates a new Markov lane
func NewMarkovLane(name string) *MarkovLane {
	return &MarkovLane{
		Name:  name,
		Chain: sequencer.NewMarkovChain(markovSeed),
	}
}

// GetNextEvent implements Lane.GetNextEvent
func (l *MarkovLane) GetNextEvent() (sequencer.Event, error) {
	return l.Chain.Generate()
}
