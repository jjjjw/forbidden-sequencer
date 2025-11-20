package main

import (
	"fmt"
	"os"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/events"
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

	// Create techno sequencer (120 BPM, 8 ticks per beat = 16th notes)
	m.Sequencer = sequencers.NewTechnoSequencer(120, 8, midiAdapter, false)

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

	// Set up event callback to send messages to TUI
	if m.Sequencer != nil {
		m.Sequencer.OnEvent = func(event events.ScheduledEvent) {
			p.Send(tui.EventMsg(event))
		}
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
