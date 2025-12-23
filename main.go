package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/conductors"
	seqlib "forbidden_sequencer/sequencer/sequencers"
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
			SelectedSequencer: "Ramp Time", // default
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

	// Create global conductor with default tick duration (100ms)
	conductor := conductors.NewConductor(100 * time.Millisecond)

	// Create global sequencer (with empty patterns initially)
	sequencer := seqlib.NewSequencer(nil, conductor, scAdapter, nil, *debug)

	// Start the sequencer (this starts the runTickLoop)
	sequencer.Start()

	m := tui.Model{
		Settings:  settings,
		Screen:    tui.ScreenMain,
		SCAdapter: scAdapter,
		Debug:     *debug,
		Conductor: conductor,
		Sequencer: sequencer,
	}

	// Create module factories
	m.ModuleFactories = []sequencers.ModuleFactory{
		&sequencers.ModulatedRhythmFactory{},
		&sequencers.RandRhythmFactory{},
		&sequencers.ArpFactory{},
		&sequencers.TechnoFactory{},
		&sequencers.MarkovChordFactory{},
	}

	// Find and activate the saved module
	m.ActiveModuleIndex = 0
	for i, factory := range m.ModuleFactories {
		if factory.GetName() == settings.SelectedSequencer {
			m.ActiveModuleIndex = i
			break
		}
	}

	// Create and initialize active module (starts paused)
	if m.ActiveModuleIndex < len(m.ModuleFactories) {
		factory := m.ModuleFactories[m.ActiveModuleIndex]
		m.ActiveModule = factory.Create(conductor)

		// Load the module's patterns into the global sequencer
		sequencer.SetPatterns(m.ActiveModule.GetPatterns())
	}

	return m
}

func main() {
	flag.Parse()

	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
