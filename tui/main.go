package main

import (
	"flag"
	"fmt"
	"os"

	"forbidden_sequencer/adapter"
	"forbidden_sequencer/controllers"
	tui "forbidden_sequencer/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var debug = flag.Bool("debug", false, "Enable debug logging")

func initialModel() tui.Model {
	// Load settings
	settings, err := tui.LoadSettings()
	if err != nil {
		fmt.Printf("Failed to load settings, using defaults: %v\n", err)
		settings = &tui.Settings{
			SelectedControllerIndex: 0, // default to first controller
		}
	}

	// Initialize sclang OSC adapter (for pattern control)
	sclangAdapter, err := adapter.SetupSClangAdapter()
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

	// Create all available controllers
	m.AvailableControllers = []controllers.Controller{
		controllers.NewCurveTimeController(sclangAdapter),
		controllers.NewMarkovTrigController(sclangAdapter),
		controllers.NewMarkovChordController(sclangAdapter),
	}

	// Set initial controller from settings (with bounds checking)
	m.ActiveControllerIndex = settings.SelectedControllerIndex
	if m.ActiveControllerIndex < 0 || m.ActiveControllerIndex >= len(m.AvailableControllers) {
		m.ActiveControllerIndex = 0 // fall back to first controller
		settings.SelectedControllerIndex = 0
	}
	m.ActiveController = m.AvailableControllers[m.ActiveControllerIndex]

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
