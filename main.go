package main

import (
	"fmt"
	"os"
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/sequencers"
	"forbidden_sequencer/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func initialModel() tui.Model {
	// Load settings
	settings, err := tui.LoadSettings()
	if err != nil {
		fmt.Printf("Failed to load settings, using defaults: %v\n", err)
		settings = &tui.Settings{
			MIDIPort:        0,
			ChannelMappings: make(map[string]uint8),
		}
	}

	m := tui.Model{
		Settings:        settings,
		Screen:          tui.ScreenMain,
		ChannelMappings: []tui.ChannelMapping{},
		SettingsOptions: []string{"MIDI Port", "Channel Mapping"},
	}

	// Initialize MIDI adapter
	midiAdapter, err := adapters.NewMIDIAdapter(settings.MIDIPort)
	if err != nil {
		m.Err = fmt.Errorf("failed to initialize MIDI adapter: %w", err)
		return m
	}
	m.MidiAdapter = midiAdapter

	// Restore channel mappings
	for eventName, channel := range settings.ChannelMappings {
		midiAdapter.SetChannelMapping(eventName, channel)
	}

	// Create modulated rhythm sequencer
	// baseTickDuration: 100ms, phraseLength: 16 ticks
	sequencer, conductor := sequencers.NewModulatedRhythmSequencer(100*time.Millisecond, 16, midiAdapter, false)
	m.Sequencer = sequencer
	m.RateChanges = conductor.RateChanges()
	m.CurrentRate = 1.0

	// Initialize sequencer (starts paused)
	m.Sequencer.Start()

	// Load MIDI ports
	if ports, err := midiAdapter.ListAvailablePorts(); err == nil {
		m.MidiPorts = ports
	}
	m.SelectedPort = midiAdapter.GetCurrentPort()

	// Initialize channel mappings list
	m.UpdateChannelMappingsList()

	return m
}

func main() {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Set up goroutines to forward channel events to TUI
	if m.Sequencer != nil {
		// Forward events
		go func() {
			for event := range m.Sequencer.GetEventsChannel() {
				p.Send(tui.EventMsg(event))
			}
		}()
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
