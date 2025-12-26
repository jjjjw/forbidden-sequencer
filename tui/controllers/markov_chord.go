package controllers

import (
	"fmt"
	"strings"

	"forbidden_sequencer/adapter"

	tea "github.com/charmbracelet/bubbletea"
)

// MarkovChordController controls the markov_chord pattern in sclang via OSC
type MarkovChordController struct {
	sclangAdapter     *adapter.OSCAdapter
	baseEventDur      float64
	phraseLength      int
	phrasesPerSection int
	rootNote          int
	debug             bool
	isPlaying         bool
	currentSection    string // "Chord" or "Percussion" (for display)
}

// NewMarkovChordController creates a new markov chord controller
func NewMarkovChordController(sclangAdapter *adapter.OSCAdapter) *MarkovChordController {
	return &MarkovChordController{
		sclangAdapter:     sclangAdapter,
		baseEventDur:      0.125,
		phraseLength:      16,
		phrasesPerSection: 2,
		rootNote:          53, // F3
		debug:             false,
		isPlaying:         false,
		currentSection:    "Chord",
	}
}

// GetName returns the display name
func (c *MarkovChordController) GetName() string {
	return "Markov Chord"
}

// GetKeybindings returns the controller-specific controls
func (c *MarkovChordController) GetKeybindings() string {
	return `p: play/stop
space: pause/resume
d/D: base event dur
r/R: phrase length
n/N: root note
s/S: phrases per section
x: debug`
}

// GetStatus returns the current state
func (c *MarkovChordController) GetStatus() string {
	var status strings.Builder

	// Pattern state
	phraseDur := float64(c.phraseLength) * c.baseEventDur
	status.WriteString(fmt.Sprintf("Base: %.3fs, Length: %d, Phrase: %.2fs\n", c.baseEventDur, c.phraseLength, phraseDur))
	status.WriteString(fmt.Sprintf("Section: %s (%d phrases)\n", c.currentSection, c.phrasesPerSection))
	status.WriteString(fmt.Sprintf("Root: %d (MIDI)", c.rootNote))

	if c.debug {
		status.WriteString("\nDEBUG")
	}

	return status.String()
}

// HandleInput processes controller-specific input
func (c *MarkovChordController) HandleInput(msg tea.KeyMsg) bool {
	switch msg.String() {
	case " ":
		// Toggle pause/resume
		if c.isPlaying {
			c.sclangAdapter.Send("/pattern/markov_chord/pause")
			c.isPlaying = false
		} else {
			c.sclangAdapter.Send("/pattern/markov_chord/resume")
			c.isPlaying = true
		}
		return true

	case "p":
		// Toggle play/stop (reset position)
		if c.isPlaying {
			c.sclangAdapter.Send("/pattern/markov_chord/stop")
			c.isPlaying = false
		} else {
			c.sclangAdapter.Send("/pattern/markov_chord/play")
			c.isPlaying = true
			c.currentSection = "Chord" // Reset to chord section on play
		}
		return true

	case "d":
		// Decrease base event duration
		if c.baseEventDur > 0.025 {
			c.baseEventDur -= 0.005
			c.sclangAdapter.Send("/pattern/markov_chord/base_event_dur", float32(c.baseEventDur))
		}
		return true

	case "D":
		// Increase base event duration
		if c.baseEventDur < 1.0 {
			c.baseEventDur += 0.005
			c.sclangAdapter.Send("/pattern/markov_chord/base_event_dur", float32(c.baseEventDur))
		}
		return true

	case "r":
		// Decrease phrase length
		if c.phraseLength > 4 {
			c.phraseLength--
			c.sclangAdapter.Send("/pattern/markov_chord/phrase_length", int32(c.phraseLength))
		}
		return true

	case "R":
		// Increase phrase length
		if c.phraseLength < 64 {
			c.phraseLength++
			c.sclangAdapter.Send("/pattern/markov_chord/phrase_length", int32(c.phraseLength))
		}
		return true

	case "n":
		// Decrease root note
		if c.rootNote > 0 {
			c.rootNote--
			c.sclangAdapter.Send("/pattern/markov_chord/root_note", int32(c.rootNote))
		}
		return true

	case "N":
		// Increase root note
		if c.rootNote < 127 {
			c.rootNote++
			c.sclangAdapter.Send("/pattern/markov_chord/root_note", int32(c.rootNote))
		}
		return true

	case "s":
		// Decrease phrases per section
		if c.phrasesPerSection > 1 {
			c.phrasesPerSection--
			c.sclangAdapter.Send("/pattern/markov_chord/phrases_per_section", int32(c.phrasesPerSection))
		}
		return true

	case "S":
		// Increase phrases per section
		if c.phrasesPerSection < 16 {
			c.phrasesPerSection++
			c.sclangAdapter.Send("/pattern/markov_chord/phrases_per_section", int32(c.phrasesPerSection))
		}
		return true

	case "x":
		// Toggle debug
		c.debug = !c.debug
		debugInt := int32(0)
		if c.debug {
			debugInt = 1
		}
		c.sclangAdapter.Send("/pattern/markov_chord/debug", debugInt)
		return true
	}

	return false
}

// Quit stops the pattern and resets to defaults
func (c *MarkovChordController) Quit() {
	c.sclangAdapter.Send("/pattern/markov_chord/reset")
	c.isPlaying = false
}
