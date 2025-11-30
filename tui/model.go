package tui

import (
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/tui/sequencers"
)

// EventLogEntry represents a single event in the event log
type EventLogEntry struct {
	Name      string
	Timestamp time.Time
}

// Screen represents the current view
type Screen int

const (
	ScreenMain Screen = iota
	ScreenSettings
	ScreenMIDIPorts
	ScreenChannelMapping
)

// Settings represents persisted application settings
type Settings struct {
	MIDIPort          int              `json:"midiPort"`
	ChannelMappings   map[string]uint8 `json:"channelMappings"`
	SelectedSequencer string           `json:"selectedSequencer"` // name of the selected sequencer
}

// Model is the main application state
type Model struct {
	MidiAdapter *adapters.MIDIAdapter
	Settings    *Settings
	IsPlaying   bool
	Screen      Screen
	Err         error

	// Sequencer management
	SequencerFactories     []sequencers.SequencerFactory // factory for each sequencer type
	ActiveSequencer        sequencers.SequencerConfig    // currently active sequencer instance
	ActiveSequencerIndex   int                           // index of active factory
	ShowingSequencerList   bool                          // true when sequencer list overlay is visible
	SelectedSequencerIndex int                           // for navigating sequencer list

	// MIDI port selection
	MidiPorts    []adapters.MIDIPortInfo
	SelectedPort int

	// Channel mapping
	ChannelMappings []ChannelMapping
	SelectedMapping int
	EditingChannel  bool
	ChannelInput    string

	// Settings menu
	SettingsOptions []string
	SelectedSetting int

	// Window size
	Width  int
	Height int

	// Event display
	EventChan chan events.ScheduledEvent
	EventLog  []EventLogEntry // stores last 100 events, newest first
}

// ChannelMapping represents a MIDI channel assignment
type ChannelMapping struct {
	Name    string
	Channel uint8
}
