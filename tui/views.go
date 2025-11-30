package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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
	// If showing sequencer list, show overlay
	if m.ShowingSequencerList {
		return m.viewSequencerList()
	}

	// Left panel - main content
	var left strings.Builder

	// Title
	left.WriteString(TitleStyle.Render("Forbidden Sequencer"))
	left.WriteString("\n\n")

	// Active sequencer name
	if m.ActiveSequencer != nil {
		left.WriteString(StatusStyle.Render("Sequencer: " + m.ActiveSequencer.GetName()))
		left.WriteString("\n\n")

		// Sequencer status
		left.WriteString(StatusStyle.Render(m.ActiveSequencer.GetStatus()))
		left.WriteString("\n\n")
	}

	// Error display
	if m.Err != nil {
		left.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", m.Err)))
		left.WriteString("\n\n")
	}

	// Status
	status := StoppedStyle.Render("MUTED")
	if m.IsPlaying {
		status = PlayingStyle.Render("PLAYING")
	}
	left.WriteString(StatusStyle.Render("Status: "))
	left.WriteString(status)
	left.WriteString("\n\n")

	// Help - layered keybindings
	var helpItems []string

	// Sequencer-specific keybindings
	if m.ActiveSequencer != nil {
		sequencerHelp := m.ActiveSequencer.GetKeybindings()
		if sequencerHelp != "" {
			helpItems = append(helpItems, sequencerHelp)
		}
	}

	// Global keybindings
	globalHelp := []string{
		"[space/p] Play/Mute",
		"[tab] Switch Sequencer",
		"[s] Settings",
		"[q] Quit",
	}
	helpItems = append(helpItems, strings.Join(globalHelp, " • "))

	left.WriteString(BoxStyle.Render(HelpStyle.Render(strings.Join(helpItems, "\n"))))

	// Right panel - event log
	right := m.viewEventLog()

	// Join left and right panels
	return lipgloss.JoinHorizontal(lipgloss.Top, left.String(), "  ", right)
}

func (m Model) viewSequencerList() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Select Sequencer"))
	b.WriteString("\n\n")

	for i, factory := range m.SequencerFactories {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.SelectedSequencerIndex {
			cursor = "> "
			style = SelectedStyle
		}
		current := ""
		if i == m.ActiveSequencerIndex {
			current = " (active)"
		}
		b.WriteString(style.Render(fmt.Sprintf("%s%s%s", cursor, factory.GetName(), current)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Help
	help := "[j/k] Navigate  [enter] Select  [esc] Cancel"
	b.WriteString(HelpStyle.Render(help))

	return b.String()
}

func (m Model) viewEventLog() string {
	// Build rows
	var rows [][]string
	limit := 30
	if len(m.EventLog) < limit {
		limit = len(m.EventLog)
	}

	for i := 0; i < limit; i++ {
		entry := m.EventLog[i]
		timestamp := entry.Timestamp.Format("15:04:05.000")
		rows = append(rows, []string{timestamp, entry.Name})
	}

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("241"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				style = style.Bold(true)
			}
			// Set min widths
			switch col {
			case 0: // Time
				style = style.Width(14)
			case 1: // Event
				style = style.Width(10)
			}
			return style
		}).
		Headers("Time", "Event").
		Rows(rows...)

	return t.String()
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
