package sequencers

import (
	"fmt"
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/patterns/modulated"
	seqlib "forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

// ModulatedRhythmFactory creates modulated rhythm sequencers
type ModulatedRhythmFactory struct{}

// GetName returns the display name
func (f *ModulatedRhythmFactory) GetName() string {
	return "Ramp Time"
}

// Create creates a new modulated rhythm config instance
func (f *ModulatedRhythmFactory) Create(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) SequencerConfig {
	return NewModulatedRhythmConfig(adapter, eventChan)
}

// ModulatedRhythmConfig wraps a modulated rhythm sequencer
type ModulatedRhythmConfig struct {
	sequencer        *seqlib.Sequencer
	conductor        *conductors.PhraseConductor
	kickPattern      *modulated.SimpleKickPattern
	hihatPattern     *modulated.SimpleHihatPattern
	rateChanges      chan<- float64
	currentRate      float64
	kickSubdivision  int
	hihatSubdivision int
}

// NewModulatedRhythmConfig creates a new modulated rhythm config
func NewModulatedRhythmConfig(adapter adapters.EventAdapter, eventChan chan<- events.ScheduledEvent) *ModulatedRhythmConfig {
	// Create sequencer with default settings
	// baseTickDuration: 100ms, phraseLength: 16 ticks
	sequencer, conductor, kickPattern, hihatPattern := seqlib.NewModulatedRhythmSequencer(
		100*time.Millisecond,
		16,
		adapter,
		eventChan,
	)

	return &ModulatedRhythmConfig{
		sequencer:        sequencer,
		conductor:        conductor,
		kickPattern:      kickPattern,
		hihatPattern:     hihatPattern,
		rateChanges:      conductor.RateChanges(),
		currentRate:      1.0,
		kickSubdivision:  1,
		hihatSubdivision: 1,
	}
}

// GetName returns the display name
func (c *ModulatedRhythmConfig) GetName() string {
	return "Ramp Time"
}

// GetKeybindings returns the sequencer-specific controls
func (c *ModulatedRhythmConfig) GetKeybindings() string {
	return "j/k: rate | h/l: kick subdiv | H/L: hihat subdiv"
}

// GetStatus returns the current state
func (c *ModulatedRhythmConfig) GetStatus() string {
	return fmt.Sprintf("Rate: %.2fx | Kick: %dx | Hihat: %dx", c.currentRate, c.kickSubdivision, c.hihatSubdivision)
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
