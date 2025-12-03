package tui

import (
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/events"
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
	OSCAdapter *adapters.OSCAdapter
	Settings   *Settings
	IsPlaying  bool
	Screen     Screen
	Err        error

	// Sequencer management
	SequencerFactories     []sequencers.SequencerFactory // factory for each sequencer type
	ActiveSequencer        sequencers.SequencerConfig    // currently active sequencer instance
	ActiveSequencerIndex   int                           // index of active factory
	ShowingSequencerList   bool                          // true when sequencer list overlay is visible
	SelectedSequencerIndex int                           // for navigating sequencer list

	// Window size
	Width  int
	Height int

	// Event display
	EventChan chan events.ScheduledEvent
	EventLog  []EventLogEntry // stores last 100 events, newest first
}
