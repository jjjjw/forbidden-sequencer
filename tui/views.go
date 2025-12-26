package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss/list"
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
	// Left panel - main content
	var left strings.Builder

	// Title
	left.WriteString(TitleStyle.Render("Forbidden Sequencer"))
	left.WriteString("\n\n")

	// Active controller name
	if m.ActiveController != nil {
		left.WriteString(StatusStyle.Render("Pattern: " + m.ActiveController.GetName()))
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
			// Parse the keybindings and render as a list
			lines := strings.Split(controllerHelp, "\n")
			var items []any
			for _, line := range lines {
				if line == "" {
					continue
				}
				// Split on ": " to separate key from description
				parts := strings.SplitN(line, ": ", 2)
				if len(parts) == 2 {
					items = append(items, KeyStyle.Render(parts[0])+" "+DescStyle.Render(parts[1]))
				} else {
					items = append(items, line)
				}
			}

			// Add global quit binding
			items = append(items, KeyStyle.Render("q")+" "+DescStyle.Render("Quit"))

			l := list.New(items...)
			left.WriteString(BoxStyle.Render(l.String()))
		}
	}

	return left.String()
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

