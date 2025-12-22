package sequencers

import (
	"fmt"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/lib"
	"forbidden_sequencer/sequencer/modules"
	"forbidden_sequencer/sequencer/patterns/arp"
	seqlib "forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

// ArpFactory creates arpeggiator modules
type ArpFactory struct{}

// GetName returns the display name
func (f *ArpFactory) GetName() string {
	return "Arpeggiator"
}

// Create creates a new arp config instance
func (f *ArpFactory) Create(conductor *conductors.Conductor) ModuleConfig {
	return NewArpConfig(conductor)
}

// ArpConfig wraps an arpeggiator module
type ArpConfig struct {
	patterns []seqlib.Pattern
	pattern  *arp.ArpPattern
}

// NewArpConfig creates a new arp config
func NewArpConfig(conductor *conductors.Conductor) *ArpConfig {
	// Create arpeggiator with default settings
	// Example sequence: C major arpeggio with rests
	// Scale degrees: 0=C, 2=E, 4=G, 7=C (octave up), rest, 4=G, 2=E, 0=C
	sequence := []int{0, 2, 4, 7, arp.RestValue, 4, 2, 0}

	patterns, pattern := modules.NewArpModule(
		sequence,
		lib.MajorScale,
		60, // root note (middle C)
	)

	return &ArpConfig{
		patterns: patterns,
		pattern:  pattern,
	}
}

// GetName returns the display name
func (c *ArpConfig) GetName() string {
	return "Arpeggiator"
}

// GetKeybindings returns the module-specific controls
func (c *ArpConfig) GetKeybindings() string {
	return "up/down: root note ±semitone • [/]: octave down/up • ;/': fifth down/up"
}

// GetStatus returns the current state
func (c *ArpConfig) GetStatus() string {
	return fmt.Sprintf("Root: %d • Transpose: %+d",
		c.pattern.GetRootNote(),
		c.pattern.GetTranspose())
}

// HandleInput processes module-specific input
func (c *ArpConfig) HandleInput(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "up":
		// Root note up one semitone
		currentRoot := c.pattern.GetRootNote()
		if currentRoot < 127 {
			c.pattern.SetRootNote(currentRoot + 1)
		}
		return true

	case "down":
		// Root note down one semitone
		currentRoot := c.pattern.GetRootNote()
		if currentRoot > 0 {
			c.pattern.SetRootNote(currentRoot - 1)
		}
		return true

	case "[":
		// Octave down
		c.pattern.ShiftOctaveDown()
		return true

	case "]":
		// Octave up
		c.pattern.ShiftOctaveUp()
		return true

	case ";":
		// Fifth down
		c.pattern.ShiftFifthDown()
		return true

	case "'":
		// Fifth up
		c.pattern.ShiftFifthUp()
		return true
	}

	return false
}

// GetPatterns returns the patterns
func (c *ArpConfig) GetPatterns() []seqlib.Pattern {
	return c.patterns
}

// Stop stops the patterns
func (c *ArpConfig) Stop() {
	for _, p := range c.patterns {
		p.Stop()
	}
}

// Play resumes playback
func (c *ArpConfig) Play() {
	for _, p := range c.patterns {
		p.Play()
	}
}
