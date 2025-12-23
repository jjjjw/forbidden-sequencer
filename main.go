package main

import (
	"flag"
	"fmt"
	"os"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/tui"
	"forbidden_sequencer/tui/controllers"

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

	// Initialize sclang OSC adapter (for pattern control)
	sclangAdapter, err := adapters.SetupSClangAdapter()
	if err != nil {
		return tui.Model{
			Settings: settings,
			Screen:   tui.ScreenMain,
			Err:      err,
		}
	}

	m := tui.Model{
		Settings:      settings,
		Screen:        tui.ScreenMain,
		SClangAdapter: sclangAdapter,
		Debug:         *debug,
	}

	// Create controller for modulated rhythm pattern
	m.ActiveController = controllers.NewModulatedRhythmController(sclangAdapter)

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
