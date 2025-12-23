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
	kickCurve        float64
	kickEvents       int
	hihatCurve       float64
	hihatEvents      int
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
		kickCurve:        2.0, // moderate ritardando
		kickEvents:       8,   // 8 events across phrase
		hihatCurve:       1.5, // lighter ritardando
		hihatEvents:      6,   // 6 events across phrase
	}
}

// GetName returns the display name
func (c *ModulatedRhythmConfig) GetName() string {
	return "Ramp Time"
}

// GetKeybindings returns the module-specific controls
func (c *ModulatedRhythmConfig) GetKeybindings() string {
	return "h/l: kick subdiv | H/L: hihat subdiv | c/C: kick curve | v/V: hihat curve | e/E: kick events | r/R: hihat events"
}

// GetStatus returns the current state
func (c *ModulatedRhythmConfig) GetStatus() string {
	return fmt.Sprintf("Kick: %dx (curve=%.1f, events=%d) | Hihat: %dx (curve=%.1f, events=%d)",
		c.kickSubdivision, c.kickCurve, c.kickEvents, c.hihatSubdivision, c.hihatCurve, c.hihatEvents)
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

	case "c":
		// Decrease kick curve (less ritardando)
		if c.kickCurve > 0.5 {
			c.kickCurve -= 0.5
			c.kickPattern.SetCurve(c.kickCurve)
		}
		return true

	case "C":
		// Increase kick curve (more ritardando)
		if c.kickCurve < 5.0 {
			c.kickCurve += 0.5
			c.kickPattern.SetCurve(c.kickCurve)
		}
		return true

	case "v":
		// Decrease hihat curve (less ritardando)
		if c.hihatCurve > 0.5 {
			c.hihatCurve -= 0.5
			c.hihatPattern.SetCurve(c.hihatCurve)
		}
		return true

	case "V":
		// Increase hihat curve (more ritardando)
		if c.hihatCurve < 5.0 {
			c.hihatCurve += 0.5
			c.hihatPattern.SetCurve(c.hihatCurve)
		}
		return true

	case "e":
		// Decrease kick events
		if c.kickEvents > 1 {
			c.kickEvents--
			c.kickPattern.SetEvents(c.kickEvents)
		}
		return true

	case "E":
		// Increase kick events
		if c.kickEvents < 16 {
			c.kickEvents++
			c.kickPattern.SetEvents(c.kickEvents)
		}
		return true

	case "r":
		// Decrease hihat events
		if c.hihatEvents > 1 {
			c.hihatEvents--
			c.hihatPattern.SetEvents(c.hihatEvents)
		}
		return true

	case "R":
		// Increase hihat events
		if c.hihatEvents < 16 {
			c.hihatEvents++
			c.hihatPattern.SetEvents(c.hihatEvents)
		}
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
