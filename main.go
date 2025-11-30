package main

import (
	"fmt"
	"os"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/tui"
	"forbidden_sequencer/tui/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

func initialModel() tui.Model {
	// Load settings
	settings, err := tui.LoadSettings()
	if err != nil {
		fmt.Printf("Failed to load settings, using defaults: %v\n", err)
		settings = &tui.Settings{
			MIDIPort:          0,
			ChannelMappings:   make(map[string]uint8),
			SelectedSequencer: "Modulated Rhythm", // default
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

	// Restore channel mappings (will override defaults if user has saved preferences)
	for eventName, channel := range settings.ChannelMappings {
		midiAdapter.SetChannelMapping(eventName, channel)
	}

	// Create event channel (owned by model)
	m.EventChan = make(chan events.ScheduledEvent, 100)

	// Create sequencer factories
	m.SequencerFactories = []sequencers.SequencerFactory{
		&sequencers.ModulatedRhythmFactory{},
		&sequencers.ArpFactory{},
		&sequencers.TechnoFactory{},
	}

	// Find and activate the saved sequencer
	m.ActiveSequencerIndex = 0
	for i, factory := range m.SequencerFactories {
		if factory.GetName() == settings.SelectedSequencer {
			m.ActiveSequencerIndex = i
			break
		}
	}

	// Create and initialize active sequencer (starts paused)
	if m.ActiveSequencerIndex < len(m.SequencerFactories) {
		factory := m.SequencerFactories[m.ActiveSequencerIndex]
		m.ActiveSequencer = factory.Create(midiAdapter, m.EventChan)
		m.ActiveSequencer.Start()
	}

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

	// Single goroutine to forward events from the channel to TUI
	go func() {
		for event := range m.EventChan {
			p.Send(tui.EventMsg(event))
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
