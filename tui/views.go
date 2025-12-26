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
	case ScreenPatternSelect:
		return m.viewPatternSelect()
	}
	return ""
}

func (m Model) viewMain() string {
	// Left panel - main content
	var left strings.Builder

	// Title
	left.WriteString(TitleStyle.Render("Forbidden Sequencer"))
	left.WriteString("\n\n")

	// Active controller name with index
	if m.ActiveController != nil {
		controllerInfo := fmt.Sprintf("Pattern: %s (%d/%d)",
			m.ActiveController.GetName(),
			m.ActiveControllerIndex+1,
			len(m.AvailableControllers))
		left.WriteString(StatusStyle.Render(controllerInfo))
		left.WriteString("\n\n")

		// Controller status
		status := m.ActiveController.GetStatus()
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

	// Help - keybindings
	if m.ActiveController != nil {
		// Controller-specific keybindings
		controllerHelp := m.ActiveController.GetKeybindings()
		if controllerHelp != "" {
			// Parse the keybindings and render as a table
			lines := strings.Split(controllerHelp, "\n")
			var rows [][]string
			for _, line := range lines {
				if line == "" {
					continue
				}
				// Split on ": " to separate key from description
				parts := strings.SplitN(line, ": ", 2)
				if len(parts) == 2 {
					rows = append(rows, []string{parts[0], parts[1]})
				}
			}

			// Add global keybindings
			rows = append(rows, []string{"tab", "Select pattern"})
			rows = append(rows, []string{"q", "Quit"})

			// Create table with blue border
			t := table.New().
				Border(lipgloss.RoundedBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("62"))).
				StyleFunc(func(row, col int) lipgloss.Style {
					if col == 0 {
						// Key column - bold and pink
						return KeyStyle
					}
					// Description column - gray
					return DescStyle
				}).
				Rows(rows...)

			left.WriteString(t.String())
		}
	}

	return left.String()
}

func (m Model) viewSettings() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Settings"))
	b.WriteString("\n\n")

	// Display current controller
	if m.ActiveController != nil {
		b.WriteString(fmt.Sprintf("Selected Pattern: %s (index %d)\n", m.ActiveController.GetName(), m.ActiveControllerIndex))
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
	if m.SClangAdapter != nil {
		b.WriteString(StatusStyle.Render("SuperCollider Configuration:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Host: %s\n", m.SClangAdapter.GetHost()))
		b.WriteString(fmt.Sprintf("  Port: %d (sclang)\n", m.SClangAdapter.GetPort()))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Help
	help := "[esc] Back"
	b.WriteString(HelpStyle.Render(help))

	return b.String()
}

func (m Model) viewPatternSelect() string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Select Pattern"))
	b.WriteString("\n\n")

	// Build the entire list with inline styling
	for i, controller := range m.AvailableControllers {
		prefix := "  "
		if i == m.SelectedPatternIndex {
			prefix = "> "
		}

		// Show indicator if this is the currently active pattern
		activeIndicator := ""
		if i == m.ActiveControllerIndex {
			activeIndicator = " (active)"
		}

		line := fmt.Sprintf("%s%d. %s%s", prefix, i+1, controller.GetName(), activeIndicator)

		// Apply color inline using lipgloss without Render
		if i == m.SelectedPatternIndex {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Inline(true).Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Help
	help := "[↑/↓] Navigate • [enter] Select • [esc] Cancel"
	b.WriteString(HelpStyle.Render(help))

	return b.String()
}

