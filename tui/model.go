package tui

import (
	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	seqlib "forbidden_sequencer/sequencer/sequencers"
	"forbidden_sequencer/tui/sequencers"
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
	SCAdapter  *adapters.SuperColliderAdapter
	Settings   *Settings
	IsPlaying  bool
	Screen     Screen
	Err        error
	Debug      bool // debug logging enabled

	// Global sequencer components (shared by all modules)
	Conductor *conductors.Conductor
	Sequencer *seqlib.Sequencer

	// Module management
	ModuleFactories     []sequencers.ModuleFactory // factory for each module type
	ActiveModule        sequencers.ModuleConfig    // currently active module instance
	ActiveModuleIndex   int                        // index of active factory
	ShowingModuleList   bool                       // true when module list overlay is visible
	SelectedModuleIndex int                        // for navigating module list

	// Window size
	Width  int
	Height int
}
