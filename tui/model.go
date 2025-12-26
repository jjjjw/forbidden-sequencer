package tui

import (
	"forbidden_sequencer/adapter"
	"forbidden_sequencer/tui/controllers"
)

// Screen represents the current view
type Screen int

const (
	ScreenMain Screen = iota
	ScreenSettings
	ScreenPatternSelect
)

// Settings represents persisted application settings
type Settings struct {
	SelectedControllerIndex int `json:"selectedControllerIndex"` // index of the selected controller
}

// Model is the main application state
type Model struct {
	SClangAdapter *adapter.OSCAdapter // OSC client for sclang (port 57120)
	Settings      *Settings
	IsPlaying     bool
	Screen        Screen
	Err           error
	Debug         bool // debug logging enabled

	// Pattern controllers
	AvailableControllers  []controllers.Controller // all available controllers
	ActiveController      controllers.Controller   // currently active controller instance
	ActiveControllerIndex int                      // index of active controller
	SelectedPatternIndex  int                      // temporary selection for pattern screen

	// Window size
	Width  int
	Height int
}
