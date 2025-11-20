package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View returns the current screen view
func (m Model) View() string {
	switch m.Screen {
	case ScreenMain:
		return m.viewMain()
	case ScreenSettings:
		return m.viewSettings()
	case ScreenMIDIPorts:
		return m.viewMIDIPorts()
	case ScreenChannelMapping:
		return m.viewChannelMapping()
	}
	return ""
}

func (m Model) viewMain() string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Forbidden Sequencer"))
	b.WriteString("\n\n")

	// Error display
	if m.Err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", m.Err)))
		b.WriteString("\n\n")
	}

	// Status
	status := StoppedStyle.Render("STOPPED")
	if m.IsPlaying {
		status = PlayingStyle.Render("PLAYING")
	}
	b.WriteString(StatusStyle.Render("Status: "))
	b.WriteString(status)
	b.WriteString("\n\n")

	// Current MIDI port
	if m.MidiAdapter != nil {
		portName := "Unknown"
		ports, err := m.MidiAdapter.ListAvailablePorts()
		if err == nil {
			for _, p := range ports {
				if p.Index == m.MidiAdapter.GetCurrentPort() {
					portName = p.Name
					break
				}
			}
		}
		b.WriteString(StatusStyle.Render(fmt.Sprintf("MIDI Port: %s", portName)))
		b.WriteString("\n\n")
	}

	// Event log
	b.WriteString(m.viewEventLog())
	b.WriteString("\n\n")

	// Help
	help := []string{
		"[space/p] Play/Stop",
		"[s] Settings",
		"[q] Quit",
	}
	b.WriteString(BoxStyle.Render(HelpStyle.Render(strings.Join(help, "  "))))

	return b.String()
}

func (m Model) viewEventLog() string {
	// All event types to display
	eventTypes := []string{"kick", "hihat"}

	var indicators []string
	for _, eventType := range eventTypes {
		active := m.ActiveEvents != nil && m.ActiveEvents[eventType]
		if active {
			indicators = append(indicators, EventActiveStyle.Render(eventType))
		} else {
			indicators = append(indicators, EventInactiveStyle.Render(eventType))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, indicators...)
}

func (m Model) viewSettings() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Settings"))
	b.WriteString("\n\n")

	for i, option := range m.SettingsOptions {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.SelectedSetting {
			cursor = "> "
			style = SelectedStyle
		}

		// Show current value
		var value string
		switch i {
		case 0: // MIDI Port
			if m.MidiAdapter != nil {
				portName := "Unknown"
				ports, err := m.MidiAdapter.ListAvailablePorts()
				if err == nil {
					for _, p := range ports {
						if p.Index == m.MidiAdapter.GetCurrentPort() {
							portName = p.Name
							break
						}
					}
				}
				value = portName
			} else {
				value = "Not connected"
			}
		case 1: // Channel Mapping
			value = fmt.Sprintf("%d mappings", len(m.Settings.ChannelMappings))
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s: %s", cursor, option, value)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Help
	help := "[j/k] Navigate  [enter] Select  [esc] Back"
	b.WriteString(HelpStyle.Render(help))

	return b.String()
}

func (m Model) viewMIDIPorts() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Select MIDI Port"))
	b.WriteString("\n\n")

	if len(m.MidiPorts) == 0 {
		b.WriteString(ErrorStyle.Render("No MIDI ports available"))
		b.WriteString("\n\n")
	} else {
		for i, port := range m.MidiPorts {
			cursor := "  "
			style := lipgloss.NewStyle()
			if i == m.SelectedPort {
				cursor = "> "
				style = SelectedStyle
			}
			current := ""
			if m.MidiAdapter != nil && i == m.MidiAdapter.GetCurrentPort() {
				current = " (current)"
			}
			b.WriteString(style.Render(fmt.Sprintf("%s%s%s", cursor, port.Name, current)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Help
	help := "[j/k] Navigate  [enter] Select  [esc] Back"
	b.WriteString(HelpStyle.Render(help))

	return b.String()
}

func (m Model) viewChannelMapping() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Channel Mapping"))
	b.WriteString("\n\n")

	if m.EditingChannel {
		mapping := m.ChannelMappings[m.SelectedMapping]
		b.WriteString(fmt.Sprintf("Edit channel for '%s': %s_", mapping.Name, m.ChannelInput))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("[enter] Save  [esc] Cancel"))
	} else {
		for i, mapping := range m.ChannelMappings {
			cursor := "  "
			style := lipgloss.NewStyle()
			if i == m.SelectedMapping {
				cursor = "> "
				style = SelectedStyle
			}
			b.WriteString(style.Render(fmt.Sprintf("%s%s: channel %d", cursor, mapping.Name, mapping.Channel)))
			b.WriteString("\n")
		}
		b.WriteString("\n")

		// Help
		help := "[j/k] Navigate  [enter/e] Edit  [esc] Back"
		b.WriteString(HelpStyle.Render(help))
	}

	return b.String()
}
