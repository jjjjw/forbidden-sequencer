package tui

import (
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	seqlib "forbidden_sequencer/sequencer/sequencers"
	"forbidden_sequencer/tui/sequencers"
)

// EventLogEntry represents a single event in the event log
type EventLogEntry struct {
	Name         string
	ReceivedTime time.Time // when the event was logged
	Timestamp    time.Time // when the event is scheduled to play
}

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

	// Event display
	EventChan chan events.ScheduledEvent
	EventLog  []EventLogEntry // stores last 100 events, newest first
}
