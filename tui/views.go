package tui

import (
	"fmt"
	"strings"
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
	var helpItems []string

	// Controller-specific keybindings
	if m.ActiveController != nil {
		controllerHelp := m.ActiveController.GetKeybindings()
		if controllerHelp != "" {
			helpItems = append(helpItems, controllerHelp)
		}
	}

	// Global keybindings
	globalHelp := []string{
		"[space/p] Play/Pause",
		"[q] Quit",
	}
	helpItems = append(helpItems, strings.Join(globalHelp, " • "))

	left.WriteString(BoxStyle.Render(HelpStyle.Render(strings.Join(helpItems, "\n"))))

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

