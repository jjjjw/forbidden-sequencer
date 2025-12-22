package sequencers

import (
	seqlib "forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

// ModuleConfig wraps a module's patterns and provides UI metadata
type ModuleConfig interface {
	// GetName returns the display name for the module
	GetName() string

	// GetKeybindings returns a description of module-specific controls
	GetKeybindings() string

	// GetStatus returns the current state/info of the module
	GetStatus() string

	// HandleInput processes module-specific key input
	// Returns true if the input was handled, false otherwise
	HandleInput(msg tea.KeyMsg) bool

	// GetPatterns returns the patterns for this module
	GetPatterns() []seqlib.Pattern

	// Stop stops all patterns
	Stop()

	// Play resumes all patterns
	Play()
}
