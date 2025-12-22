package sequencers

import (
	"fmt"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/modules"
	"forbidden_sequencer/sequencer/patterns/modulated"
	seqlib "forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

// ModulatedRhythmFactory creates modulated rhythm modules
type ModulatedRhythmFactory struct{}

// GetName returns the display name
func (f *ModulatedRhythmFactory) GetName() string {
	return "Ramp Time"
}

// Create creates a new modulated rhythm config instance
func (f *ModulatedRhythmFactory) Create(conductor *conductors.Conductor) ModuleConfig {
	return NewModulatedRhythmConfig(conductor)
}

// ModulatedRhythmConfig wraps a modulated rhythm module
type ModulatedRhythmConfig struct {
	patterns         []seqlib.Pattern
	kickPattern      *modulated.SimpleKickPattern
	hihatPattern     *modulated.SimpleHihatPattern
	kickSubdivision  int
	hihatSubdivision int
}

// NewModulatedRhythmConfig creates a new modulated rhythm config
func NewModulatedRhythmConfig(conductor *conductors.Conductor) *ModulatedRhythmConfig {
	// Create module patterns
	// phraseLength: 16 ticks
	patterns, kickPattern, hihatPattern := modules.NewModulatedRhythmModule(conductor, 16)

	return &ModulatedRhythmConfig{
		patterns:         patterns,
		kickPattern:      kickPattern,
		hihatPattern:     hihatPattern,
		kickSubdivision:  1,
		hihatSubdivision: 1,
	}
}

// GetName returns the display name
func (c *ModulatedRhythmConfig) GetName() string {
	return "Ramp Time"
}

// GetKeybindings returns the module-specific controls
func (c *ModulatedRhythmConfig) GetKeybindings() string {
	return "h/l: kick subdiv | H/L: hihat subdiv"
}

// GetStatus returns the current state
func (c *ModulatedRhythmConfig) GetStatus() string {
	return fmt.Sprintf("Kick: %dx | Hihat: %dx", c.kickSubdivision, c.hihatSubdivision)
}

// HandleInput processes module-specific input
func (c *ModulatedRhythmConfig) HandleInput(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "h":
		// Decrease kick subdivision
		if c.kickSubdivision > 1 {
			c.kickSubdivision--
			c.kickPattern.SetSubdivision(c.kickSubdivision)
		}
		return true

	case "l":
		// Increase kick subdivision
		c.kickSubdivision++
		c.kickPattern.SetSubdivision(c.kickSubdivision)
		return true

	case "H":
		// Decrease hihat subdivision
		if c.hihatSubdivision > 1 {
			c.hihatSubdivision--
			c.hihatPattern.SetSubdivision(c.hihatSubdivision)
		}
		return true

	case "L":
		// Increase hihat subdivision
		c.hihatSubdivision++
		c.hihatPattern.SetSubdivision(c.hihatSubdivision)
		return true
	}

	return false
}

// GetPatterns returns the patterns
func (c *ModulatedRhythmConfig) GetPatterns() []seqlib.Pattern {
	return c.patterns
}

// Stop stops the patterns
func (c *ModulatedRhythmConfig) Stop() {
	for _, p := range c.patterns {
		p.Stop()
	}
}

// Play resumes playback
func (c *ModulatedRhythmConfig) Play() {
	for _, p := range c.patterns {
		p.Play()
	}
}
