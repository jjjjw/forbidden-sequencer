package tui

import (
	"fmt"

	"forbidden_sequencer/sequencer/events"

	tea "github.com/charmbracelet/bubbletea"
)

// EventMsg is sent when an event is received from the sequencer
type EventMsg events.ScheduledEvent

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case EventMsg:
		// Add event to log (skip rests)
		event := events.ScheduledEvent(msg)
		if event.Event.Type == events.EventTypeRest {
			return m, nil
		}
		entry := EventLogEntry{
			Name:      event.Event.Name,
			Timestamp: event.Timing.Timestamp,
		}
		// Prepend to keep newest first
		m.EventLog = append([]EventLogEntry{entry}, m.EventLog...)
		// Keep only last 100 events
		if len(m.EventLog) > 100 {
			m.EventLog = m.EventLog[:100]
		}
		return m, nil

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			if m.ActiveSequencer != nil {
				m.ActiveSequencer.Stop()
			}
			return m, tea.Quit
		}

		// Screen-specific keys
		switch m.Screen {
		case ScreenMain:
			return m.updateMain(msg)
		case ScreenSettings:
			return m.updateSettings(msg)
		}
	}

	return m, nil
}

func (m Model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If showing sequencer list, handle list navigation
	if m.ShowingSequencerList {
		return m.updateSequencerList(msg)
	}

	// Try sequencer-specific input first
	if m.ActiveSequencer != nil {
		if m.ActiveSequencer.HandleInput(msg) {
			return m, nil
		}
	}

	// Global keys
	switch msg.String() {
	case "q", "esc":
		if m.ActiveSequencer != nil {
			m.ActiveSequencer.Stop()
		}
		return m, tea.Quit

	case " ", "p":
		if m.ActiveSequencer != nil {
			if m.IsPlaying {
				m.ActiveSequencer.Stop()
				m.IsPlaying = false
			} else {
				m.ActiveSequencer.Play()
				m.IsPlaying = true
			}
		}

	case "tab":
		// Toggle sequencer list
		m.ShowingSequencerList = true
		m.SelectedSequencerIndex = m.ActiveSequencerIndex

	case "s":
		// Go to settings
		m.Screen = ScreenSettings
	}

	return m, nil
}

func (m Model) updateSequencerList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		// Close sequencer list without switching
		m.ShowingSequencerList = false

	case "up", "k":
		if m.SelectedSequencerIndex > 0 {
			m.SelectedSequencerIndex--
		}

	case "down", "j":
		if m.SelectedSequencerIndex < len(m.SequencerFactories)-1 {
			m.SelectedSequencerIndex++
		}

	case "enter":
		// Switch to selected sequencer
		if m.SelectedSequencerIndex != m.ActiveSequencerIndex {
			// Stop and destroy current sequencer
			if m.ActiveSequencer != nil {
				m.ActiveSequencer.Stop()
				m.ActiveSequencer = nil
			}

			// Create new sequencer from factory
			m.ActiveSequencerIndex = m.SelectedSequencerIndex
			if m.ActiveSequencerIndex < len(m.SequencerFactories) && m.OSCAdapter != nil {
				factory := m.SequencerFactories[m.ActiveSequencerIndex]
				m.ActiveSequencer = factory.Create(m.OSCAdapter, m.EventChan)
				m.ActiveSequencer.Start()
				m.IsPlaying = false

				// Save to settings
				m.Settings.SelectedSequencer = factory.GetName()
				if err := SaveSettings(m.Settings); err != nil {
					m.Err = fmt.Errorf("failed to save settings: %w", err)
				}
			}
		}
		m.ShowingSequencerList = false
	}

	return m, nil
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.Screen = ScreenMain
	}

	return m, nil
}

