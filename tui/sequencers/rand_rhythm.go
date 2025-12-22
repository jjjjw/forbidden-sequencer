package sequencers

import (
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/modules"
	seqlib "forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

// RandRhythmFactory creates randomized rhythm modules using Markov chains
type RandRhythmFactory struct{}

// GetName returns the display name
func (f *RandRhythmFactory) GetName() string {
	return "Markov"
}

// Create creates a new rand rhythm config instance
func (f *RandRhythmFactory) Create(conductor *conductors.Conductor) ModuleConfig {
	return NewRandRhythmConfig(conductor)
}

// RandRhythmConfig wraps randomized rhythm patterns
type RandRhythmConfig struct {
	patterns []seqlib.Pattern
}

// NewRandRhythmConfig creates a new rand rhythm config
func NewRandRhythmConfig(conductor *conductors.Conductor) *RandRhythmConfig {
	// Create module patterns
	// phraseLength: 16 ticks
	patterns := modules.NewRandRhythmModule(conductor, 16)

	return &RandRhythmConfig{
		patterns: patterns,
	}
}

// GetName returns the display name
func (c *RandRhythmConfig) GetName() string {
	return "Markov"
}

// GetKeybindings returns the module-specific controls
func (c *RandRhythmConfig) GetKeybindings() string {
	return ""
}

// GetStatus returns the current state
func (c *RandRhythmConfig) GetStatus() string {
	return ""
}

// HandleInput processes module-specific input
func (c *RandRhythmConfig) HandleInput(msg tea.KeyMsg) bool {
	return false
}

// GetPatterns returns the patterns
func (c *RandRhythmConfig) GetPatterns() []seqlib.Pattern {
	return c.patterns
}

// Stop stops the patterns
func (c *RandRhythmConfig) Stop() {
	for _, p := range c.patterns {
		p.Stop()
	}
}

// Play resumes playback
func (c *RandRhythmConfig) Play() {
	for _, p := range c.patterns {
		p.Play()
	}
}
