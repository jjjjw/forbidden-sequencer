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
	settings    *Settings
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		fmt.Printf("Failed to load settings, using defaults: %v\n", err)
		settings = &Settings{
			MIDIPort:        0,
			ChannelMappings: make(map[string]uint8),
		}
	}
	a.settings = settings

	// Initialize MIDI adapter with saved port
	midiAdapter, err := adapters.NewMIDIAdapter(settings.MIDIPort)
	if err != nil {
		fmt.Printf("Failed to initialize MIDI adapter: %v\n", err)
		return
	}
	a.midiAdapter = midiAdapter

	// Restore channel mappings
	for eventName, channel := range settings.ChannelMappings {
		midiAdapter.SetChannelMapping(eventName, channel)
	}

	// Create techno sequencer (120 BPM, 8 ticks per beat = 16th notes)
	a.sequencer = sequencers.NewTechnoSequencer(120, 8, midiAdapter, false)

	// For now, start the sequencer on initialization
	a.StartTechno()
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

// GetMIDIPorts returns a list of available MIDI output ports
func (a *App) GetMIDIPorts() ([]adapters.MIDIPortInfo, error) {
	if a.midiAdapter == nil {
		return nil, fmt.Errorf("MIDI adapter not initialized")
	}
	return a.midiAdapter.ListAvailablePorts()
}

// GetCurrentMIDIPort returns the currently selected MIDI port index
func (a *App) GetCurrentMIDIPort() int {
	if a.midiAdapter == nil {
		return -1
	}
	return a.midiAdapter.GetCurrentPort()
}

// SetMIDIPort sets the MIDI output port
func (a *App) SetMIDIPort(portIndex int) error {
	if a.midiAdapter == nil {
		return fmt.Errorf("MIDI adapter not initialized")
	}
	if err := a.midiAdapter.SetPort(portIndex); err != nil {
		return err
	}

	// Save to settings
	a.settings.MIDIPort = portIndex
	if err := SaveSettings(a.settings); err != nil {
		fmt.Printf("Failed to save settings: %v\n", err)
	}

	return nil
}

// GetChannelMappings returns all channel mappings
func (a *App) GetChannelMappings() map[string]uint8 {
	if a.midiAdapter == nil {
		return make(map[string]uint8)
	}
	return a.midiAdapter.GetAllChannelMappings()
}

// SetChannelMapping sets the MIDI channel for a specific event name
func (a *App) SetChannelMapping(eventName string, channel int) error {
	if a.midiAdapter == nil {
		return fmt.Errorf("MIDI adapter not initialized")
	}
	if channel < 0 || channel > 15 {
		return fmt.Errorf("MIDI channel must be between 0 and 15")
	}
	a.midiAdapter.SetChannelMapping(eventName, uint8(channel))

	// Save to settings
	a.settings.ChannelMappings[eventName] = uint8(channel)
	if err := SaveSettings(a.settings); err != nil {
		fmt.Printf("Failed to save settings: %v\n", err)
	}

	return nil
}

// RemoveChannelMapping removes a channel mapping for a specific event name
func (a *App) RemoveChannelMapping(eventName string) error {
	if a.midiAdapter == nil {
		return fmt.Errorf("MIDI adapter not initialized")
	}

	// Remove from adapter (set to default channel 0)
	a.midiAdapter.SetChannelMapping(eventName, 0)

	// Remove from settings
	delete(a.settings.ChannelMappings, eventName)
	if err := SaveSettings(a.settings); err != nil {
		fmt.Printf("Failed to save settings: %v\n", err)
	}

	return nil
}
