package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

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

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if m.ActiveController != nil {
				m.ActiveController.Quit()
			}
			// Save settings before quitting
			if m.Settings != nil {
				m.Settings.SelectedControllerIndex = m.ActiveControllerIndex
				SaveSettings(m.Settings)
			}
			return m, tea.Quit

		case "tab":
			// Show pattern selection screen
			if len(m.AvailableControllers) > 1 {
				m.SelectedPatternIndex = m.ActiveControllerIndex
				m.Screen = ScreenPatternSelect
			}
			return m, nil
		}

		// Screen-specific keys
		switch m.Screen {
		case ScreenMain:
			return m.updateMain(msg)
		case ScreenSettings:
			return m.updateSettings(msg)
		case ScreenPatternSelect:
			return m.updatePatternSelect(msg)
		}
	}

	return m, nil
}

func (m Model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Try controller-specific input first
	if m.ActiveController != nil {
		if m.ActiveController.HandleInput(msg) {
			return m, nil
		}
	}

	// Global keys (none currently, controller handles play/pause)

	return m, nil
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.Screen = ScreenMain
	}

	return m, nil
}

func (m Model) updatePatternSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel and return to main
		m.Screen = ScreenMain
		return m, nil

	case "up", "k":
		// Move selection up
		if m.SelectedPatternIndex > 0 {
			m.SelectedPatternIndex--
		}
		return m, nil

	case "down", "j":
		// Move selection down
		if m.SelectedPatternIndex < len(m.AvailableControllers)-1 {
			m.SelectedPatternIndex++
		}
		return m, nil

	case "enter":
		// Select pattern
		if m.SelectedPatternIndex != m.ActiveControllerIndex {
			// Quit the old controller
			if m.ActiveController != nil {
				m.ActiveController.Quit()
			}

			// Switch to new controller
			m.ActiveControllerIndex = m.SelectedPatternIndex
			m.ActiveController = m.AvailableControllers[m.ActiveControllerIndex]

			// Save settings immediately after switching
			if m.Settings != nil {
				m.Settings.SelectedControllerIndex = m.ActiveControllerIndex
				SaveSettings(m.Settings)
			}
		}

		// Return to main screen
		m.Screen = ScreenMain
		return m, nil
	}

	return m, nil
}

