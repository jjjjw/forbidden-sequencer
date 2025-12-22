package sequencers

import (
	"fmt"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/modules"
	markovchord "forbidden_sequencer/sequencer/patterns/markov_chord"
	seqlib "forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

// MarkovChordFactory creates markov chord modules
type MarkovChordFactory struct{}

// GetName returns the display name
func (f *MarkovChordFactory) GetName() string {
	return "Markov Chord"
}

// Create creates a new markov chord config instance
func (f *MarkovChordFactory) Create(conductor *conductors.Conductor) ModuleConfig {
	return NewMarkovChordConfig(conductor)
}

// MarkovChordConfig wraps a markov chord module
type MarkovChordConfig struct {
	patterns          []seqlib.Pattern
	chordPattern      *markovchord.ChordPattern
	phrasesPerSection int
}

// NewMarkovChordConfig creates a new markov chord config
func NewMarkovChordConfig(conductor *conductors.Conductor) *MarkovChordConfig {
	// Create module patterns
	// phraseLength: 16 ticks, phrasesPerSection: 2
	phrasesPerSection := 2
	patterns, chordPattern := modules.NewMarkovChordModule(
		conductor,
		16,
		phrasesPerSection,
	)

	return &MarkovChordConfig{
		patterns:          patterns,
		chordPattern:      chordPattern,
		phrasesPerSection: phrasesPerSection,
	}
}

// GetName returns the display name
func (c *MarkovChordConfig) GetName() string {
	return "Markov Chord"
}

// GetKeybindings returns the module-specific controls
func (c *MarkovChordConfig) GetKeybindings() string {
	return ""
}

// GetStatus returns the current state
func (c *MarkovChordConfig) GetStatus() string {
	sectionName := "Chord"
	if c.chordPattern.IsPercussionSection() {
		sectionName = "Percussion"
	}
	return fmt.Sprintf("Section: %s | %d phrases/section", sectionName, c.phrasesPerSection)
}

// HandleInput processes module-specific input
func (c *MarkovChordConfig) HandleInput(msg tea.KeyMsg) bool {
	return false
}

// GetPatterns returns the patterns
func (c *MarkovChordConfig) GetPatterns() []seqlib.Pattern {
	return c.patterns
}

// Stop stops the patterns
func (c *MarkovChordConfig) Stop() {
	for _, p := range c.patterns {
		p.Stop()
	}
}

// Play resumes playback
func (c *MarkovChordConfig) Play() {
	for _, p := range c.patterns {
		p.Play()
	}
}
