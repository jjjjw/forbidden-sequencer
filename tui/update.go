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

