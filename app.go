package main

import (
	"context"
	"fmt"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/sequencers"
)

// App struct
type App struct {
	ctx         context.Context
	sequencer   *sequencers.Sequencer
	midiAdapter *adapters.MIDIAdapter
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize MIDI adapter (using default port)
	midiAdapter, err := adapters.NewMIDIAdapter(-1)
	if err != nil {
		fmt.Printf("Failed to initialize MIDI adapter: %v\n", err)
		return
	}
	a.midiAdapter = midiAdapter

	// Create techno sequencer (120 BPM, 4 ticks per beat = 16th notes)
	a.sequencer = sequencers.NewTechnoSequencer(120, 4, midiAdapter, false)
}

// StartTechno starts the techno sequencer
func (a *App) StartTechno() string {
	if a.sequencer == nil {
		return "Sequencer not initialized"
	}
	a.sequencer.Start()
	return "Techno sequencer started!"
}

// Stop stops the sequencer
func (a *App) Stop() string {
	if a.sequencer == nil {
		return "Sequencer not initialized"
	}
	// Sequencer doesn't have a Stop method yet, but we can pause
	a.sequencer.Pause()
	return "Sequencer stopped"
}

// Pause pauses the sequencer
func (a *App) Pause() string {
	if a.sequencer == nil {
		return "Sequencer not initialized"
	}
	a.sequencer.Pause()
	return "Sequencer paused"
}

// Resume resumes the sequencer
func (a *App) Resume() string {
	if a.sequencer == nil {
		return "Sequencer not initialized"
	}
	a.sequencer.Resume()
	return "Sequencer resumed"
}
