package sequencers

import (
	"fmt"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/events"
	seqlib "forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

// TechnoFactory creates techno sequencers
type TechnoFactory struct{}

// GetName returns the display name
func (f *TechnoFactory) GetName() string {
	return "Techno"
}

// Create creates a new techno config instance
func (f *TechnoFactory) Create(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) SequencerConfig {
	return NewTechnoConfig(adapter, eventChan)
}

// TechnoConfig wraps a techno sequencer
type TechnoConfig struct {
	sequencer *seqlib.Sequencer
	bpm       float64
}

// NewTechnoConfig creates a new techno config
func NewTechnoConfig(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) *TechnoConfig {
	// Create sequencer with default settings
	// BPM: 140
	bpm := 140.0
	sequencer := seqlib.NewTechnoSequencer(bpm, adapter, eventChan)

	return &TechnoConfig{
		sequencer: sequencer,
		bpm:       bpm,
	}
}

// GetName returns the display name
func (c *TechnoConfig) GetName() string {
	return "Techno"
}

// GetKeybindings returns the sequencer-specific controls
func (c *TechnoConfig) GetKeybindings() string {
	return "" // No sequencer-specific controls for now
}

// GetStatus returns the current state
func (c *TechnoConfig) GetStatus() string {
	return fmt.Sprintf("BPM: %.1f", c.bpm)
}

// HandleInput processes sequencer-specific input
func (c *TechnoConfig) HandleInput(msg tea.KeyMsg) bool {
	// No sequencer-specific input for now
	return false
}

// Start starts the sequencer
func (c *TechnoConfig) Start() {
	c.sequencer.Start()
}

// Stop stops the sequencer
func (c *TechnoConfig) Stop() {
	c.sequencer.Stop()
}

// Play resumes playback
func (c *TechnoConfig) Play() {
	c.sequencer.Play()
}
