package main

import (
	"flag"
	"fmt"
	"os"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/tui"
	"forbidden_sequencer/tui/sequencers"

	tea "github.com/charmbracelet/bubbletea"
)

var debug = flag.Bool("debug", false, "Enable debug logging")

func initialModel() tui.Model {
	// Load settings
	settings, err := tui.LoadSettings()
	if err != nil {
		fmt.Printf("Failed to load settings, using defaults: %v\n", err)
		settings = &tui.Settings{
			SelectedSequencer: "Modulated Rhythm", // default
		}
	}

	// Initialize SuperCollider adapter
	scAdapter, err := adapters.SetupSuperColliderAdapter(*debug)
	if err != nil {
		return tui.Model{
			Settings: settings,
			Screen:   tui.ScreenMain,
			Err:      err,
		}
	}

	m := tui.Model{
		Settings:     settings,
		Screen:       tui.ScreenMain,
		SCAdapter:    scAdapter,
		Debug:        *debug,
	}

	// Create event channel (owned by model)
	m.EventChan = make(chan events.ScheduledEvent, 100)

	// Create sequencer factories
	m.SequencerFactories = []sequencers.SequencerFactory{
		&sequencers.ModulatedRhythmFactory{},
		&sequencers.RandRhythmFactory{},
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
	// Use SuperCollider adapter for server-side scheduling
	if m.ActiveSequencerIndex < len(m.SequencerFactories) {
		factory := m.SequencerFactories[m.ActiveSequencerIndex]
		m.ActiveSequencer = factory.Create(scAdapter, m.EventChan)
		m.ActiveSequencer.Start()
	}

	return m
}

func main() {
	flag.Parse()

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
