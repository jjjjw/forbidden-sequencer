package main

import (
	"fmt"
	"os"
	"strings"

	"forbidden_sequencer/sequencer/adapters"
	"forbidden_sequencer/sequencer/sequencers"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen represents the current view
type Screen int

const (
	ScreenMain Screen = iota
	ScreenSettings
	ScreenMIDIPorts
	ScreenChannelMapping
)

// Model is the main application state
type Model struct {
	sequencer   *sequencers.Sequencer
	midiAdapter *adapters.MIDIAdapter
	settings    *Settings
	isPlaying   bool
	screen      Screen
	err         error

	// MIDI port selection
	midiPorts    []adapters.MIDIPortInfo
	selectedPort int

	// Channel mapping
	channelMappings []channelMapping
	selectedMapping int
	editingChannel  bool
	channelInput    string

	// Settings menu
	settingsOptions  []string
	selectedSetting  int

	// Window size
	width  int
	height int
}

type channelMapping struct {
	name    string
	channel uint8
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	playingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("46"))

	stoppedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)
)

func initialModel() Model {
	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		fmt.Printf("Failed to load settings, using defaults: %v\n", err)
		settings = &Settings{
			MIDIPort:        0,
			ChannelMappings: make(map[string]uint8),
		}
	}

	m := Model{
		settings:        settings,
		screen:          ScreenMain,
		channelMappings: []channelMapping{},
		settingsOptions: []string{"MIDI Port", "Channel Mapping"},
	}

	// Initialize MIDI adapter
	midiAdapter, err := adapters.NewMIDIAdapter(settings.MIDIPort)
	if err != nil {
		m.err = fmt.Errorf("failed to initialize MIDI adapter: %w", err)
		return m
	}
	m.midiAdapter = midiAdapter

	// Restore channel mappings
	for eventName, channel := range settings.ChannelMappings {
		midiAdapter.SetChannelMapping(eventName, channel)
	}

	// Create techno sequencer (120 BPM, 8 ticks per beat = 16th notes)
	m.sequencer = sequencers.NewTechnoSequencer(120, 8, midiAdapter, false)

	// Load MIDI ports
	if ports, err := midiAdapter.ListAvailablePorts(); err == nil {
		m.midiPorts = ports
	}
	m.selectedPort = midiAdapter.GetCurrentPort()

	// Initialize channel mappings list
	m.updateChannelMappingsList()

	return m
}

func (m *Model) updateChannelMappingsList() {
	// Default event names that can be mapped
	eventNames := []string{"kick", "hihat"}

	m.channelMappings = make([]channelMapping, 0, len(eventNames))
	for _, name := range eventNames {
		channel := uint8(0)
		if m.midiAdapter != nil {
			channel = m.midiAdapter.GetChannelMapping(name)
		}
		m.channelMappings = append(m.channelMappings, channelMapping{
			name:    name,
			channel: channel,
		})
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			if m.midiAdapter != nil {
				m.midiAdapter.Close()
			}
			if m.sequencer != nil {
				m.sequencer.Stop()
			}
			return m, tea.Quit
		}

		// Screen-specific keys
		switch m.screen {
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
		if m.midiAdapter != nil {
			m.midiAdapter.Close()
		}
		if m.sequencer != nil {
			m.sequencer.Stop()
		}
		return m, tea.Quit

	case " ", "p":
		if m.sequencer != nil {
			if m.isPlaying {
				m.sequencer.Stop()
				m.isPlaying = false
			} else {
				m.sequencer.Start()
				m.isPlaying = true
			}
		}

	case "s":
		// Go to settings
		m.selectedSetting = 0
		m.screen = ScreenSettings
	}

	return m, nil
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.screen = ScreenMain

	case "up", "k":
		if m.selectedSetting > 0 {
			m.selectedSetting--
		}

	case "down", "j":
		if m.selectedSetting < len(m.settingsOptions)-1 {
			m.selectedSetting++
		}

	case "enter":
		switch m.selectedSetting {
		case 0: // MIDI Port
			if m.midiAdapter != nil {
				if ports, err := m.midiAdapter.ListAvailablePorts(); err == nil {
					m.midiPorts = ports
				}
			}
			m.screen = ScreenMIDIPorts
		case 1: // Channel Mapping
			m.updateChannelMappingsList()
			m.selectedMapping = 0
			m.screen = ScreenChannelMapping
		}
	}

	return m, nil
}

func (m Model) updateMIDIPorts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.screen = ScreenSettings

	case "up", "k":
		if m.selectedPort > 0 {
			m.selectedPort--
		}

	case "down", "j":
		if m.selectedPort < len(m.midiPorts)-1 {
			m.selectedPort++
		}

	case "enter":
		if m.midiAdapter != nil && m.selectedPort < len(m.midiPorts) {
			if err := m.midiAdapter.SetPort(m.selectedPort); err != nil {
				m.err = err
			} else {
				// Save to settings
				m.settings.MIDIPort = m.selectedPort
				if err := SaveSettings(m.settings); err != nil {
					m.err = fmt.Errorf("failed to save settings: %w", err)
				}
				m.err = nil
			}
		}
		m.screen = ScreenSettings
	}

	return m, nil
}

func (m Model) updateChannelMapping(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingChannel {
		switch msg.String() {
		case "esc":
			m.editingChannel = false
			m.channelInput = ""

		case "enter":
			// Parse and save channel
			var channel int
			if _, err := fmt.Sscanf(m.channelInput, "%d", &channel); err == nil {
				if channel >= 0 && channel <= 15 && m.selectedMapping < len(m.channelMappings) {
					mapping := &m.channelMappings[m.selectedMapping]
					mapping.channel = uint8(channel)
					if m.midiAdapter != nil {
						m.midiAdapter.SetChannelMapping(mapping.name, uint8(channel))
					}
					// Save to settings
					m.settings.ChannelMappings[mapping.name] = uint8(channel)
					if err := SaveSettings(m.settings); err != nil {
						m.err = fmt.Errorf("failed to save settings: %w", err)
					}
				}
			}
			m.editingChannel = false
			m.channelInput = ""

		case "backspace":
			if len(m.channelInput) > 0 {
				m.channelInput = m.channelInput[:len(m.channelInput)-1]
			}

		default:
			// Only accept digits
			if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
				if len(m.channelInput) < 2 {
					m.channelInput += msg.String()
				}
			}
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "esc":
		m.screen = ScreenSettings

	case "up", "k":
		if m.selectedMapping > 0 {
			m.selectedMapping--
		}

	case "down", "j":
		if m.selectedMapping < len(m.channelMappings)-1 {
			m.selectedMapping++
		}

	case "enter", "e":
		m.editingChannel = true
		m.channelInput = fmt.Sprintf("%d", m.channelMappings[m.selectedMapping].channel)
	}

	return m, nil
}

func (m Model) View() string {
	switch m.screen {
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
	b.WriteString(titleStyle.Render("Forbidden Sequencer"))
	b.WriteString("\n\n")

	// Error display
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	// Status
	status := stoppedStyle.Render("STOPPED")
	if m.isPlaying {
		status = playingStyle.Render("PLAYING")
	}
	b.WriteString(statusStyle.Render("Status: "))
	b.WriteString(status)
	b.WriteString("\n\n")

	// Current MIDI port
	if m.midiAdapter != nil {
		portName := "Unknown"
		ports, err := m.midiAdapter.ListAvailablePorts()
		if err == nil {
			for _, p := range ports {
				if p.Index == m.midiAdapter.GetCurrentPort() {
					portName = p.Name
					break
				}
			}
		}
		b.WriteString(statusStyle.Render(fmt.Sprintf("MIDI Port: %s", portName)))
		b.WriteString("\n\n")
	}

	// Help
	help := []string{
		"[space/p] Play/Stop",
		"[s] Settings",
		"[q] Quit",
	}
	b.WriteString(boxStyle.Render(helpStyle.Render(strings.Join(help, "  "))))

	return b.String()
}

func (m Model) viewSettings() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Settings"))
	b.WriteString("\n\n")

	for i, option := range m.settingsOptions {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.selectedSetting {
			cursor = "> "
			style = selectedStyle
		}

		// Show current value
		var value string
		switch i {
		case 0: // MIDI Port
			if m.midiAdapter != nil {
				portName := "Unknown"
				ports, err := m.midiAdapter.ListAvailablePorts()
				if err == nil {
					for _, p := range ports {
						if p.Index == m.midiAdapter.GetCurrentPort() {
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
			value = fmt.Sprintf("%d mappings", len(m.settings.ChannelMappings))
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s: %s", cursor, option, value)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Help
	help := "[j/k] Navigate  [enter] Select  [esc] Back"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func (m Model) viewMIDIPorts() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Select MIDI Port"))
	b.WriteString("\n\n")

	if len(m.midiPorts) == 0 {
		b.WriteString(errorStyle.Render("No MIDI ports available"))
		b.WriteString("\n\n")
	} else {
		for i, port := range m.midiPorts {
			cursor := "  "
			style := lipgloss.NewStyle()
			if i == m.selectedPort {
				cursor = "> "
				style = selectedStyle
			}
			current := ""
			if m.midiAdapter != nil && i == m.midiAdapter.GetCurrentPort() {
				current = " (current)"
			}
			b.WriteString(style.Render(fmt.Sprintf("%s%s%s", cursor, port.Name, current)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Help
	help := "[j/k] Navigate  [enter] Select  [esc] Back"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func (m Model) viewChannelMapping() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Channel Mapping"))
	b.WriteString("\n\n")

	if m.editingChannel {
		mapping := m.channelMappings[m.selectedMapping]
		b.WriteString(fmt.Sprintf("Edit channel for '%s': %s_", mapping.name, m.channelInput))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("[enter] Save  [esc] Cancel"))
	} else {
		for i, mapping := range m.channelMappings {
			cursor := "  "
			style := lipgloss.NewStyle()
			if i == m.selectedMapping {
				cursor = "> "
				style = selectedStyle
			}
			b.WriteString(style.Render(fmt.Sprintf("%s%s: channel %d", cursor, mapping.name, mapping.channel)))
			b.WriteString("\n")
		}
		b.WriteString("\n")

		// Help
		help := "[j/k] Navigate  [enter/e] Edit  [esc] Back"
		b.WriteString(helpStyle.Render(help))
	}

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
