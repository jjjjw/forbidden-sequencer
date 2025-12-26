package controllers

import (
	"fmt"
	"strings"

	"forbidden_sequencer/adapter"

	tea "github.com/charmbracelet/bubbletea"
)

// CurveTimeController controls the curve_time pattern in sclang via OSC
type CurveTimeController struct {
	sclangAdapter *adapter.OSCAdapter
	baseEventDur  float64
	phraseEvents  int
	kickCurve     float64
	kickEvents    int
	kickOffset    int
	hihatCurve    float64
	hihatEvents   int
	hihatOffset   int
	debug         bool
	isPlaying     bool
	activeSynth   int // 0=phrase, 1=kick, 2=hihat
}

// NewCurveTimeController creates a new curve time controller
func NewCurveTimeController(sclangAdapter *adapter.OSCAdapter) *CurveTimeController {
	return &CurveTimeController{
		sclangAdapter: sclangAdapter,
		baseEventDur:  0.125,
		phraseEvents:  16,
		kickCurve:     1.5,
		kickEvents:    8,
		kickOffset:    0,
		hihatCurve:    1.5,
		hihatEvents:   8,
		hihatOffset:   0,
		debug:         false,
		isPlaying:     false,
		activeSynth:   1, // Start with kick selected
	}
}

// GetName returns the display name
func (c *CurveTimeController) GetName() string {
	return "Curve Time"
}

// GetKeybindings returns the controller-specific controls
func (c *CurveTimeController) GetKeybindings() string {
	return `p: play/stop
space: pause/resume
1-2: select synth
d/D: base event dur
r/R: phrase events
c/C: curve (active synth)
e/E: events (active synth)
o/O: offset (active synth)
x: debug`
}

// GetStatus returns the current state
func (c *CurveTimeController) GetStatus() string {
	var status strings.Builder

	// Phrase state
	phraseDur := float64(c.phraseEvents) * c.baseEventDur
	status.WriteString(fmt.Sprintf("Base: %.3fs, Events: %d, Phrase: %.2fs\n\n", c.baseEventDur, c.phraseEvents, phraseDur))

	// Synths section
	status.WriteString("Synths:\n")

	// Kick state (highlight if active)
	kickPrefix := "  "
	if c.activeSynth == 1 {
		kickPrefix = "> "
	}
	status.WriteString(fmt.Sprintf("%s1. Kick: curve=%.1f, events=%d, offset=%+d\n", kickPrefix, c.kickCurve, c.kickEvents, c.kickOffset))

	// Hihat state (highlight if active)
	hihatPrefix := "  "
	if c.activeSynth == 2 {
		hihatPrefix = "> "
	}
	status.WriteString(fmt.Sprintf("%s2. Hihat: curve=%.1f, events=%d, offset=%+d", hihatPrefix, c.hihatCurve, c.hihatEvents, c.hihatOffset))

	if c.debug {
		status.WriteString("\nDEBUG")
	}

	return status.String()
}

// HandleInput processes controller-specific input
func (c *CurveTimeController) HandleInput(msg tea.KeyMsg) bool {
	switch msg.String() {
	case " ":
		// Toggle pause/resume
		if c.isPlaying {
			c.sclangAdapter.Send("/pattern/curve_time/pause")
			c.isPlaying = false
		} else {
			c.sclangAdapter.Send("/pattern/curve_time/resume")
			c.isPlaying = true
		}
		return true

	case "p":
		// Toggle play/stop (reset position)
		if c.isPlaying {
			c.sclangAdapter.Send("/pattern/curve_time/stop")
			c.isPlaying = false
		} else {
			c.sclangAdapter.Send("/pattern/curve_time/play")
			c.isPlaying = true
		}
		return true

	case "1":
		// Select kick synth
		c.activeSynth = 1
		return true

	case "2":
		// Select hihat synth
		c.activeSynth = 2
		return true

	case "d":
		// Decrease base event duration
		if c.baseEventDur > 0.025 {
			c.baseEventDur -= 0.005
			c.sclangAdapter.Send("/pattern/curve_time/base_event_dur", float32(c.baseEventDur))
		}
		return true

	case "D":
		// Increase base event duration
		if c.baseEventDur < 1.0 {
			c.baseEventDur += 0.005
			c.sclangAdapter.Send("/pattern/curve_time/base_event_dur", float32(c.baseEventDur))
		}
		return true

	case "r":
		// Decrease phrase events
		if c.phraseEvents > 16 {
			c.phraseEvents--
			c.sclangAdapter.Send("/pattern/curve_time/phrase_events", int32(c.phraseEvents))
		}
		return true

	case "R":
		// Increase phrase events
		if c.phraseEvents < 32 {
			c.phraseEvents++
			c.sclangAdapter.Send("/pattern/curve_time/phrase_events", int32(c.phraseEvents))
		}
		return true

	case "c":
		// Decrease curve for active synth
		if c.activeSynth == 1 {
			if c.kickCurve > 0.5 {
				c.kickCurve -= 0.1
				c.sclangAdapter.Send("/pattern/curve_time/kick/curve", float32(c.kickCurve))
			}
		} else if c.activeSynth == 2 {
			if c.hihatCurve > 0.5 {
				c.hihatCurve -= 0.1
				c.sclangAdapter.Send("/pattern/curve_time/hihat/curve", float32(c.hihatCurve))
			}
		}
		return true

	case "C":
		// Increase curve for active synth
		if c.activeSynth == 1 {
			if c.kickCurve < 2.0 {
				c.kickCurve += 0.1
				c.sclangAdapter.Send("/pattern/curve_time/kick/curve", float32(c.kickCurve))
			}
		} else if c.activeSynth == 2 {
			if c.hihatCurve < 2.0 {
				c.hihatCurve += 0.1
				c.sclangAdapter.Send("/pattern/curve_time/hihat/curve", float32(c.hihatCurve))
			}
		}
		return true

	case "e":
		// Decrease events for active synth
		if c.activeSynth == 1 {
			if c.kickEvents > 1 {
				c.kickEvents--
				c.sclangAdapter.Send("/pattern/curve_time/kick/events", int32(c.kickEvents))
			}
		} else if c.activeSynth == 2 {
			if c.hihatEvents > 1 {
				c.hihatEvents--
				c.sclangAdapter.Send("/pattern/curve_time/hihat/events", int32(c.hihatEvents))
			}
		}
		return true

	case "E":
		// Increase events for active synth
		if c.activeSynth == 1 {
			if c.kickEvents < 16 {
				c.kickEvents++
				c.sclangAdapter.Send("/pattern/curve_time/kick/events", int32(c.kickEvents))
			}
		} else if c.activeSynth == 2 {
			if c.hihatEvents < 16 {
				c.hihatEvents++
				c.sclangAdapter.Send("/pattern/curve_time/hihat/events", int32(c.hihatEvents))
			}
		}
		return true

	case "o":
		// Decrease offset for active synth
		if c.activeSynth == 1 {
			if c.kickOffset > -(c.phraseEvents - 1) {
				c.kickOffset--
				c.sclangAdapter.Send("/pattern/curve_time/kick/offset", int32(c.kickOffset))
			}
		} else if c.activeSynth == 2 {
			if c.hihatOffset > -(c.phraseEvents - 1) {
				c.hihatOffset--
				c.sclangAdapter.Send("/pattern/curve_time/hihat/offset", int32(c.hihatOffset))
			}
		}
		return true

	case "O":
		// Increase offset for active synth
		if c.activeSynth == 1 {
			if c.kickOffset < (c.phraseEvents - 1) {
				c.kickOffset++
				c.sclangAdapter.Send("/pattern/curve_time/kick/offset", int32(c.kickOffset))
			}
		} else if c.activeSynth == 2 {
			if c.hihatOffset < (c.phraseEvents - 1) {
				c.hihatOffset++
				c.sclangAdapter.Send("/pattern/curve_time/hihat/offset", int32(c.hihatOffset))
			}
		}
		return true

	case "x":
		// Toggle debug
		c.debug = !c.debug
		debugInt := int32(0)
		if c.debug {
			debugInt = 1
		}
		c.sclangAdapter.Send("/pattern/curve_time/debug", debugInt)
		return true
	}

	return false
}

// Quit stops the pattern and resets to defaults
func (c *CurveTimeController) Quit() {
	c.sclangAdapter.Send("/pattern/curve_time/reset")
	c.isPlaying = false
}
