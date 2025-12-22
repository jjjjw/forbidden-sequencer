package sequencers

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/modules"
	seqlib "forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

// TechnoFactory creates techno modules
type TechnoFactory struct{}

// GetName returns the display name
func (f *TechnoFactory) GetName() string {
	return "Techno"
}

// Create creates a new techno config instance
func (f *TechnoFactory) Create(conductor *conductors.Conductor) ModuleConfig {
	return NewTechnoConfig(conductor)
}

// TechnoConfig wraps a techno module
type TechnoConfig struct {
	patterns []seqlib.Pattern
	bpm      float64
}

// NewTechnoConfig creates a new techno config
func NewTechnoConfig(conductor *conductors.Conductor) *TechnoConfig {
	// Create module patterns
	// BPM: 140, ticksPerBeat: 4 (16th notes)
	bpm := 140.0
	ticksPerBeat := 4

	// Calculate tick duration from BPM and set on conductor
	beatsPerSecond := bpm / 60.0
	ticksPerSecond := beatsPerSecond * float64(ticksPerBeat)
	tickDuration := time.Duration(float64(time.Second) / ticksPerSecond)
	conductor.SetTickDuration(tickDuration)

	patterns := modules.NewTechnoModule(conductor, ticksPerBeat)

	return &TechnoConfig{
		patterns: patterns,
		bpm:      bpm,
	}
}

// GetName returns the display name
func (c *TechnoConfig) GetName() string {
	return "Techno"
}

// GetKeybindings returns the module-specific controls
func (c *TechnoConfig) GetKeybindings() string {
	return ""
}

// GetStatus returns the current state
func (c *TechnoConfig) GetStatus() string {
	return fmt.Sprintf("BPM: %.1f", c.bpm)
}

// HandleInput processes module-specific input
func (c *TechnoConfig) HandleInput(msg tea.KeyMsg) bool {
	return false
}

// GetPatterns returns the patterns
func (c *TechnoConfig) GetPatterns() []seqlib.Pattern {
	return c.patterns
}

// Stop stops the patterns
func (c *TechnoConfig) Stop() {
	for _, p := range c.patterns {
		p.Stop()
	}
}

// Play resumes playback
func (c *TechnoConfig) Play() {
	for _, p := range c.patterns {
		p.Play()
	}
}
