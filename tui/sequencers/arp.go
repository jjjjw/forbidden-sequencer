package sequencers

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
	"forbidden_sequencer/sequencer/patterns/arp"
	seqlib "forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

// ArpFactory creates arpeggiator sequencers
type ArpFactory struct{}

// GetName returns the display name
func (f *ArpFactory) GetName() string {
	return "Arpeggiator"
}

// Create creates a new arp config instance
func (f *ArpFactory) Create(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) SequencerConfig {
	return NewArpConfig(adapter, eventChan)
}

// ArpConfig wraps an arpeggiator sequencer
type ArpConfig struct {
	sequencer   *seqlib.Sequencer
	pattern     *arp.ArpPattern
	rateChanges chan<- float64
	currentRate float64
}

// NewArpConfig creates a new arp config
func NewArpConfig(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) *ArpConfig {
	// Create arpeggiator with default settings
	// Example sequence: C major arpeggio with rests
	// Scale degrees: 0=C, 2=E, 4=G, 7=C (octave up), rest, 4=G, 2=E, 0=C
	sequence := []int{0, 2, 4, 7, arp.RestValue, 4, 2, 0}

	sequencer, pattern, conductor := seqlib.NewArpSequencer(
		150*time.Millisecond, // baseTickDuration
		sequence,
		lib.MajorScale,
		60,      // root note (middle C)
		adapter,
		eventChan,
	)

	return &ArpConfig{
		sequencer:   sequencer,
		pattern:     pattern,
		rateChanges: conductor.RateChanges(),
		currentRate: 1.0,
	}
}

// GetName returns the display name
func (c *ArpConfig) GetName() string {
	return "Arpeggiator"
}

// GetKeybindings returns the sequencer-specific controls
func (c *ArpConfig) GetKeybindings() string {
	return "j/k: adjust rate • up/down: root note ±semitone • [/]: octave down/up • ;/': fifth down/up"
}

// GetStatus returns the current state
func (c *ArpConfig) GetStatus() string {
	return fmt.Sprintf("Rate: %.2fx • Root: %d • Transpose: %+d",
		c.currentRate,
		c.pattern.GetRootNote(),
		c.pattern.GetTranspose())
}

// HandleInput processes sequencer-specific input
func (c *ArpConfig) HandleInput(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "k":
		// Increase rate
		c.currentRate *= 1.1
		select {
		case c.rateChanges <- c.currentRate:
		default:
		}
		return true

	case "j":
		// Decrease rate
		c.currentRate /= 1.1
		if c.currentRate < 0.1 {
			c.currentRate = 0.1
		}
		select {
		case c.rateChanges <- c.currentRate:
		default:
		}
		return true

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

// Start starts the sequencer
func (c *ArpConfig) Start() {
	c.sequencer.Start()
}

// Stop stops the sequencer
func (c *ArpConfig) Stop() {
	c.sequencer.Stop()
}

// Play resumes playback
func (c *ArpConfig) Play() {
	c.sequencer.Play()
}
