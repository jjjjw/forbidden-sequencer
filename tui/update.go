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
			if m.MidiAdapter != nil {
				m.MidiAdapter.Close()
			}
			if m.Sequencer != nil {
				m.Sequencer.Stop()
			}
			return m, tea.Quit
		}

		// Screen-specific keys
		switch m.Screen {
		case ScreenMain:
			return m.updateMain(msg)
		case ScreenSettings:
			return m.updateSettings(msg)
		case ScreenMIDIPorts:
			return m.updateMIDIPorts(msg)
		case ScreenChannelMapping:
			return m.updateChannelMapping(msg)
		}
	}

	return m, nil
}

func (m Model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		if m.MidiAdapter != nil {
			m.MidiAdapter.Close()
		}
		if m.Sequencer != nil {
			m.Sequencer.Stop()
		}
		return m, tea.Quit

	case " ", "p":
		if m.Sequencer != nil {
			if m.IsPlaying {
				m.Sequencer.Stop()
				m.IsPlaying = false
			} else {
				m.Sequencer.Play()
				m.IsPlaying = true
			}
		}

	case "r":
		// Reset to beginning
		if m.Sequencer != nil {
			m.Sequencer.Reset()
		}

	case "s":
		// Go to settings
		m.SelectedSetting = 0
		m.Screen = ScreenSettings

	case "up", "k":
		// Increase rate
		if m.RateChanges != nil {
			m.CurrentRate *= 1.1
			select {
			case m.RateChanges <- m.CurrentRate:
			default:
			}
		}

	case "down", "j":
		// Decrease rate
		if m.RateChanges != nil {
			m.CurrentRate /= 1.1
			if m.CurrentRate < 0.1 {
				m.CurrentRate = 0.1
			}
			select {
			case m.RateChanges <- m.CurrentRate:
			default:
			}
		}
	}

	return m, nil
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.Screen = ScreenMain

	case "up", "k":
		if m.SelectedSetting > 0 {
			m.SelectedSetting--
		}

	case "down", "j":
		if m.SelectedSetting < len(m.SettingsOptions)-1 {
			m.SelectedSetting++
		}

	case "enter":
		switch m.SelectedSetting {
		case 0: // MIDI Port
			if m.MidiAdapter != nil {
				if ports, err := m.MidiAdapter.ListAvailablePorts(); err == nil {
					m.MidiPorts = ports
				}
			}
			m.Screen = ScreenMIDIPorts
		case 1: // Channel Mapping
			m.UpdateChannelMappingsList()
			m.SelectedMapping = 0
			m.Screen = ScreenChannelMapping
		}
	}

	return m, nil
}

func (m Model) updateMIDIPorts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.Screen = ScreenSettings

	case "up", "k":
		if m.SelectedPort > 0 {
			m.SelectedPort--
		}

	case "down", "j":
		if m.SelectedPort < len(m.MidiPorts)-1 {
			m.SelectedPort++
		}

	case "enter":
		if m.MidiAdapter != nil && m.SelectedPort < len(m.MidiPorts) {
			if err := m.MidiAdapter.SetPort(m.SelectedPort); err != nil {
				m.Err = err
			} else {
				// Save to settings
				m.Settings.MIDIPort = m.SelectedPort
				if err := SaveSettings(m.Settings); err != nil {
					m.Err = fmt.Errorf("failed to save settings: %w", err)
				}
				m.Err = nil
			}
		}
		m.Screen = ScreenSettings
	}

	return m, nil
}

func (m Model) updateChannelMapping(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.EditingChannel {
		switch msg.String() {
		case "esc":
			m.EditingChannel = false
			m.ChannelInput = ""

		case "enter":
			// Parse and save channel
			var channel int
			if _, err := fmt.Sscanf(m.ChannelInput, "%d", &channel); err == nil {
				if channel >= 0 && channel <= 15 && m.SelectedMapping < len(m.ChannelMappings) {
					mapping := &m.ChannelMappings[m.SelectedMapping]
					mapping.Channel = uint8(channel)
					if m.MidiAdapter != nil {
						m.MidiAdapter.SetChannelMapping(mapping.Name, uint8(channel))
					}
					// Save to settings
					m.Settings.ChannelMappings[mapping.Name] = uint8(channel)
					if err := SaveSettings(m.Settings); err != nil {
						m.Err = fmt.Errorf("failed to save settings: %w", err)
					}
				}
			}
			m.EditingChannel = false
			m.ChannelInput = ""

		case "backspace":
			if len(m.ChannelInput) > 0 {
				m.ChannelInput = m.ChannelInput[:len(m.ChannelInput)-1]
			}

		default:
			// Only accept digits
			if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
				if len(m.ChannelInput) < 2 {
					m.ChannelInput += msg.String()
				}
			}
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "esc":
		m.Screen = ScreenSettings

	case "up", "k":
		if m.SelectedMapping > 0 {
			m.SelectedMapping--
		}

	case "down", "j":
		if m.SelectedMapping < len(m.ChannelMappings)-1 {
			m.SelectedMapping++
		}

	case "enter", "e":
		m.EditingChannel = true
		m.ChannelInput = fmt.Sprintf("%d", m.ChannelMappings[m.SelectedMapping].Channel)
	}

	return m, nil
}

// UpdateChannelMappingsList refreshes the channel mappings from the adapter
func (m *Model) UpdateChannelMappingsList() {
	// Default event names that can be mapped
	eventNames := []string{"kick", "hihat"}

	m.ChannelMappings = make([]ChannelMapping, 0, len(eventNames))
	for _, name := range eventNames {
		channel := uint8(0)
		if m.MidiAdapter != nil {
			channel = m.MidiAdapter.GetChannelMapping(name)
		}
		m.ChannelMappings = append(m.ChannelMappings, ChannelMapping{
			Name:    name,
			Channel: channel,
		})
	}
}
