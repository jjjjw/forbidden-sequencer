package tui

import (
	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/sequencers"
)

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
	MIDIPort        int              `json:"midiPort"`
	ChannelMappings map[string]uint8 `json:"channelMappings"`
}

// Model is the main application state
type Model struct {
	Sequencer   *sequencers.Sequencer
	MidiAdapter *adapters.MIDIAdapter
	Settings    *Settings
	IsPlaying   bool
	Screen      Screen
	Err         error

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
	EventChan    chan events.ScheduledEvent
	ActiveEvents map[string]bool // tracks which events are currently "lit"
}

// ChannelMapping represents a MIDI channel assignment
type ChannelMapping struct {
	Name    string
	Channel uint8
}
