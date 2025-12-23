package tui

import (
	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/tui/controllers"
)

// Screen represents the current view
type Screen int

const (
	ScreenMain Screen = iota
	ScreenSettings
)

// Settings represents persisted application settings
type Settings struct {
	SelectedSequencer string `json:"selectedSequencer"` // name of the selected sequencer
}

// Model is the main application state
type Model struct {
	SClangAdapter *adapters.OSCAdapter // OSC client for sclang (port 57120)
	Settings      *Settings
	IsPlaying     bool
	Screen        Screen
	Err           error
	Debug         bool // debug logging enabled

	// Active pattern controller
	ActiveController controllers.Controller // currently active controller instance

	// Window size
	Width  int
	Height int
}
