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
	}
	return ""
}

func (m Model) viewMain() string {
	// If showing sequencer list, show overlay
	if m.ShowingModuleList {
		return m.viewSequencerList()
	}

	// Left panel - main content
	var left strings.Builder

	// Title
	left.WriteString(TitleStyle.Render("Forbidden Sequencer"))
	left.WriteString("\n\n")

	// Active module name
	if m.ActiveModule != nil {
		left.WriteString(StatusStyle.Render("Module: " + m.ActiveModule.GetName()))
		left.WriteString("\n\n")

		// Global tempo
		if m.Conductor != nil {
			tickDuration := m.Conductor.GetTickDuration()
			left.WriteString(StatusStyle.Render(fmt.Sprintf("Tick Duration: %.0fms", float64(tickDuration.Milliseconds()))))
			left.WriteString("\n\n")
		}

		// Module status
		status := m.ActiveModule.GetStatus()
		if status != "" {
			left.WriteString(StatusStyle.Render(status))
			left.WriteString("\n\n")
		}
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
	if m.ActiveModule != nil {
		sequencerHelp := m.ActiveModule.GetKeybindings()
		if sequencerHelp != "" {
			helpItems = append(helpItems, sequencerHelp)
		}
	}

	// Global keybindings
	globalHelp := []string{
		"[j/k] Tick Duration",
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

	for i, factory := range m.ModuleFactories {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.SelectedModuleIndex {
			cursor = "> "
			style = SelectedStyle
		}
		current := ""
		if i == m.ActiveModuleIndex {
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
		scheduledTime := entry.Timestamp.Format("15:04:05.000")
		rows = append(rows, []string{entry.Name, scheduledTime})
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
			case 0: // Event
				style = style.Width(10)
			case 1: // Time
				style = style.Width(14)
			}
			return style
		}).
		Headers("Event", "Time").
		Rows(rows...)

	return t.String()
}

func (m Model) viewSettings() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Settings"))
	b.WriteString("\n\n")

	// Display current sequencer
	if m.Settings != nil && m.Settings.SelectedSequencer != "" {
		b.WriteString(fmt.Sprintf("Selected Sequencer: %s\n", m.Settings.SelectedSequencer))
	}
	b.WriteString("\n")

	// Display debug status
	debugStatus := "disabled"
	if m.Debug {
		debugStatus = "enabled"
	}
	b.WriteString(fmt.Sprintf("Debug Logging: %s\n", debugStatus))
	b.WriteString("\n")

	// Display SuperCollider adapter configuration
	if m.SCAdapter != nil {
		b.WriteString(StatusStyle.Render("SuperCollider Configuration:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Host: %s\n", m.SCAdapter.GetHost()))
		b.WriteString(fmt.Sprintf("  Port: %d (scsynth)\n", m.SCAdapter.GetPort()))
		b.WriteString("\n")

		// Display SynthDef mappings
		synthDefs := m.SCAdapter.GetAllSynthDefMappings()
		if len(synthDefs) > 0 {
			b.WriteString(StatusStyle.Render("SynthDef Mappings:"))
			b.WriteString("\n")
			for eventName, synthDef := range synthDefs {
				b.WriteString(fmt.Sprintf("  %s → %s\n", eventName, synthDef))
			}
			b.WriteString("\n")
		}

		// Display Bus mappings
		buses := m.SCAdapter.GetAllBusMappings()
		if len(buses) > 0 {
			b.WriteString(StatusStyle.Render("Bus Routing:"))
			b.WriteString("\n")
			for eventName, busID := range buses {
				busName := "master out"
				if busID != 0 {
					busName = fmt.Sprintf("bus %d", busID)
				}
				b.WriteString(fmt.Sprintf("  %s → %s\n", eventName, busName))
			}
			b.WriteString("\n")
		}

	}

	b.WriteString("\n")

	// Help
	help := "[esc] Back"
	b.WriteString(HelpStyle.Render(help))

	return b.String()
}

