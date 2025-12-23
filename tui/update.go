package tui

import (
	"fmt"
	"time"

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
		case "ctrl+c":
			if m.ActiveModule != nil {
				m.ActiveModule.Stop()
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
	if m.ShowingModuleList {
		return m.updateSequencerList(msg)
	}

	// Try sequencer-specific input first
	if m.ActiveModule != nil {
		if m.ActiveModule.HandleInput(msg) {
			return m, nil
		}
	}

	// Global keys
	switch msg.String() {
	case "q", "esc":
		if m.ActiveModule != nil {
			m.ActiveModule.Stop()
		}
		return m, tea.Quit

	case " ", "p":
		if m.ActiveModule != nil {
			if m.IsPlaying {
				m.ActiveModule.Stop()
				m.IsPlaying = false
			} else {
				m.ActiveModule.Play()
				m.IsPlaying = true
			}
		}

	case "j", "down":
		// Global: increase tick duration (slow down)
		if m.Conductor != nil {
			currentDuration := m.Conductor.GetTickDuration()
			newDuration := time.Duration(float64(currentDuration) * 1.1)
			m.Conductor.SetTickDuration(newDuration)
		}

	case "k", "up":
		// Global: decrease tick duration (speed up)
		if m.Conductor != nil {
			currentDuration := m.Conductor.GetTickDuration()
			newDuration := time.Duration(float64(currentDuration) / 1.1)
			m.Conductor.SetTickDuration(newDuration)
		}

	case "tab":
		// Toggle sequencer list
		m.ShowingModuleList = true
		m.SelectedModuleIndex = m.ActiveModuleIndex

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
		m.ShowingModuleList = false

	case "up", "k":
		if m.SelectedModuleIndex > 0 {
			m.SelectedModuleIndex--
		}

	case "down", "j":
		if m.SelectedModuleIndex < len(m.ModuleFactories)-1 {
			m.SelectedModuleIndex++
		}

	case "enter":
		// Switch to selected module
		if m.SelectedModuleIndex != m.ActiveModuleIndex {
			// Stop current module patterns
			wasPlaying := m.IsPlaying
			if m.ActiveModule != nil {
				m.ActiveModule.Stop()
			}

			// Create new module from factory
			m.ActiveModuleIndex = m.SelectedModuleIndex
			if m.ActiveModuleIndex < len(m.ModuleFactories) && m.Conductor != nil {
				factory := m.ModuleFactories[m.ActiveModuleIndex]
				m.ActiveModule = factory.Create(m.Conductor)

				// Load new patterns into global sequencer
				m.Sequencer.SetPatterns(m.ActiveModule.GetPatterns())

				// Resume playing if was playing before
				if wasPlaying {
					m.ActiveModule.Play()
				} else {
					m.IsPlaying = false
				}

				// Save to settings
				m.Settings.SelectedSequencer = factory.GetName()
				if err := SaveSettings(m.Settings); err != nil {
					m.Err = fmt.Errorf("failed to save settings: %w", err)
				}
			}
		}
		m.ShowingModuleList = false
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

