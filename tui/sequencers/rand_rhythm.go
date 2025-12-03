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

// RandRhythmFactory creates randomized rhythm sequencers using Markov chains
type RandRhythmFactory struct{}

// GetName returns the display name
func (f *RandRhythmFactory) GetName() string {
	return "Rand Rhythm"
}

// Create creates a new rand rhythm config instance
func (f *RandRhythmFactory) Create(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) SequencerConfig {
	return NewRandRhythmConfig(adapter, eventChan)
}

// RandRhythmConfig wraps a randomized rhythm sequencer
type RandRhythmConfig struct {
	sequencer   *seqlib.Sequencer
	conductor   *conductors.ModulatedRhythmConductor
	rateChanges chan<- float64
	currentRate float64
}

// NewRandRhythmConfig creates a new rand rhythm config
func NewRandRhythmConfig(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) *RandRhythmConfig {
	// Create sequencer with default settings
	// baseTickDuration: 100ms, phraseLength: 16 ticks
	sequencer, conductor := seqlib.NewRandRhythmSequencer(
		100*time.Millisecond,
		16,
		adapter,
		eventChan,
		false, // debug
	)

	return &RandRhythmConfig{
		sequencer:   sequencer,
		conductor:   conductor,
		rateChanges: conductor.RateChanges(),
		currentRate: 1.0,
	}
}

// GetName returns the display name
func (c *RandRhythmConfig) GetName() string {
	return "Rand Rhythm"
}

// GetKeybindings returns the sequencer-specific controls
func (c *RandRhythmConfig) GetKeybindings() string {
	return "j/k: adjust rate"
}

// GetStatus returns the current state
func (c *RandRhythmConfig) GetStatus() string {
	return fmt.Sprintf("Rate: %.2fx", c.currentRate)
}

// HandleInput processes sequencer-specific input
func (c *RandRhythmConfig) HandleInput(msg tea.KeyMsg) bool {
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
func (c *RandRhythmConfig) Start() {
	c.sequencer.Start()
}

// Stop stops the sequencer
func (c *RandRhythmConfig) Stop() {
	c.sequencer.Stop()
}

// Play resumes playback
func (c *RandRhythmConfig) Play() {
	c.sequencer.Play()
}
