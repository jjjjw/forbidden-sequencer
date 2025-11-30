package sequencers

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	seqlib "forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

// ModulatedRhythmFactory creates modulated rhythm sequencers
type ModulatedRhythmFactory struct{}

// GetName returns the display name
func (f *ModulatedRhythmFactory) GetName() string {
	return "Modulated Rhythm"
}

// Create creates a new modulated rhythm config instance
func (f *ModulatedRhythmFactory) Create(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) SequencerConfig {
	return NewModulatedRhythmConfig(adapter, eventChan)
}

// ModulatedRhythmConfig wraps a modulated rhythm sequencer
type ModulatedRhythmConfig struct {
	sequencer   *seqlib.Sequencer
	conductor   *conductors.ModulatedRhythmConductor
	rateChanges chan<- float64
	currentRate float64
}

// NewModulatedRhythmConfig creates a new modulated rhythm config
func NewModulatedRhythmConfig(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) *ModulatedRhythmConfig {
	// Create sequencer with default settings
	// baseTickDuration: 100ms, phraseLength: 16 ticks
	sequencer, conductor := seqlib.NewModulatedRhythmSequencer(
		100*time.Millisecond,
		16,
		adapter,
		eventChan,
		false, // debug
	)

	return &ModulatedRhythmConfig{
		sequencer:   sequencer,
		conductor:   conductor,
		rateChanges: conductor.RateChanges(),
		currentRate: 1.0,
	}
}

// GetName returns the display name
func (c *ModulatedRhythmConfig) GetName() string {
	return "Modulated Rhythm"
}

// GetKeybindings returns the sequencer-specific controls
func (c *ModulatedRhythmConfig) GetKeybindings() string {
	return "j/k: adjust rate"
}

// GetStatus returns the current state
func (c *ModulatedRhythmConfig) GetStatus() string {
	return fmt.Sprintf("Rate: %.2fx", c.currentRate)
}

// HandleInput processes sequencer-specific input
func (c *ModulatedRhythmConfig) HandleInput(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "up", "k":
		// Increase rate
		c.currentRate *= 1.1
		select {
		case c.rateChanges <- c.currentRate:
		default:
		}
		return true

	case "down", "j":
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
	}

	return false
}

// Start starts the sequencer
func (c *ModulatedRhythmConfig) Start() {
	c.sequencer.Start()
}

// Stop stops the sequencer
func (c *ModulatedRhythmConfig) Stop() {
	c.sequencer.Stop()
}

// Play resumes playback
func (c *ModulatedRhythmConfig) Play() {
	c.sequencer.Play()
}
