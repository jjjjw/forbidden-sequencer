package controllers

import tea "github.com/charmbracelet/bubbletea"

// Controller represents a pattern controller that manages interaction with sclang
type Controller interface {
	// GetName returns the display name
	GetName() string

	// GetKeybindings returns help text for controller-specific controls
	GetKeybindings() string

	// GetStatus returns current state as display string
	GetStatus() string

	// HandleInput processes controller-specific keyboard input
	// Returns true if the input was handled
	HandleInput(msg tea.KeyMsg) bool

	// Quit cleans up and stops the pattern
	Quit()
}
