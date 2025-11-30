package sequencers

import (
	tea "github.com/charmbracelet/bubbletea"
)

// SequencerConfig wraps a sequencer instance and provides UI metadata
type SequencerConfig interface {
	// GetName returns the display name for the sequencer
	GetName() string

	// GetKeybindings returns a description of sequencer-specific controls
	GetKeybindings() string

	// GetStatus returns the current state/info of the sequencer
	GetStatus() string

	// HandleInput processes sequencer-specific key input
	// Returns true if the input was handled, false otherwise
	HandleInput(msg tea.KeyMsg) bool

	// Start starts the sequencer
	Start()

	// Stop stops the sequencer
	Stop()

	// Play resumes playback
	Play()
}
