package controllers

import (
	"fmt"
	"strings"

	"forbidden_sequencer/sequencer/adapters"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hypebeast/go-osc/osc"
)

// ModulatedRhythmController controls the mod_rhy pattern in sclang via OSC
type ModulatedRhythmController struct {
	sclangAdapter *adapters.OSCAdapter
	phraseDur     float64
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

// NewModulatedRhythmController creates a new modulated rhythm controller
func NewModulatedRhythmController(sclangAdapter *adapters.OSCAdapter) *ModulatedRhythmController {
	return &ModulatedRhythmController{
		sclangAdapter: sclangAdapter,
		phraseDur:     2.0,
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
func (c *ModulatedRhythmController) GetName() string {
	return "Ramp Time"
}

// GetKeybindings returns the controller-specific controls
func (c *ModulatedRhythmController) GetKeybindings() string {
	return `p: play/stop
space: pause/resume
1: select kick
2: select hihat
d/D: phrase dur
r/R: phrase events
c/C: curve (active synth)
e/E: events (active synth)
o/O: offset (active synth)
x: debug`
}

// GetStatus returns the current state
func (c *ModulatedRhythmController) GetStatus() string {
	var status strings.Builder

	// Phrase state
	status.WriteString(fmt.Sprintf("Phrase: %.1fs, %d events\n\n", c.phraseDur, c.phraseEvents))

	// Synths section
	status.WriteString("Synths:\n")

	// Kick state (highlight if active)
	kickPrefix := "  "
	if c.activeSynth == 1 {
		kickPrefix = "> "
	}
	status.WriteString(fmt.Sprintf("%sKick: curve=%.1f, events=%d, offset=%+d\n", kickPrefix, c.kickCurve, c.kickEvents, c.kickOffset))

	// Hihat state (highlight if active)
	hihatPrefix := "  "
	if c.activeSynth == 2 {
		hihatPrefix = "> "
	}
	status.WriteString(fmt.Sprintf("%sHihat: curve=%.1f, events=%d, offset=%+d", hihatPrefix, c.hihatCurve, c.hihatEvents, c.hihatOffset))

	if c.debug {
		status.WriteString("\nDEBUG")
	}

	return status.String()
}

// HandleInput processes controller-specific input
func (c *ModulatedRhythmController) HandleInput(msg tea.KeyMsg) bool {
	switch msg.String() {
	case " ":
		// Toggle pause/resume
		if c.isPlaying {
			c.sendOSC("/pattern/mod_rhy/pause")
			c.isPlaying = false
		} else {
			c.sendOSC("/pattern/mod_rhy/resume")
			c.isPlaying = true
		}
		return true

	case "p":
		// Toggle play/stop (reset position)
		if c.isPlaying {
			c.sendOSC("/pattern/mod_rhy/stop")
			c.isPlaying = false
		} else {
			c.sendOSC("/pattern/mod_rhy/play")
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
		// Decrease phrase duration
		if c.phraseDur > 0.5 {
			c.phraseDur -= 0.1
			c.sendOSC("/pattern/mod_rhy/phrase_dur", float32(c.phraseDur))
		}
		return true

	case "D":
		// Increase phrase duration
		if c.phraseDur < 10.0 {
			c.phraseDur += 0.1
			c.sendOSC("/pattern/mod_rhy/phrase_dur", float32(c.phraseDur))
		}
		return true

	case "r":
		// Decrease phrase events
		if c.phraseEvents > 16 {
			c.phraseEvents--
			c.sendOSC("/pattern/mod_rhy/phrase_events", int32(c.phraseEvents))
		}
		return true

	case "R":
		// Increase phrase events
		if c.phraseEvents < 32 {
			c.phraseEvents++
			c.sendOSC("/pattern/mod_rhy/phrase_events", int32(c.phraseEvents))
		}
		return true

	case "c":
		// Decrease curve for active synth
		if c.activeSynth == 1 {
			if c.kickCurve > 0.5 {
				c.kickCurve -= 0.1
				c.sendOSC("/pattern/mod_rhy/kick/curve", float32(c.kickCurve))
			}
		} else if c.activeSynth == 2 {
			if c.hihatCurve > 0.5 {
				c.hihatCurve -= 0.1
				c.sendOSC("/pattern/mod_rhy/hihat/curve", float32(c.hihatCurve))
			}
		}
		return true

	case "C":
		// Increase curve for active synth
		if c.activeSynth == 1 {
			if c.kickCurve < 2.0 {
				c.kickCurve += 0.1
				c.sendOSC("/pattern/mod_rhy/kick/curve", float32(c.kickCurve))
			}
		} else if c.activeSynth == 2 {
			if c.hihatCurve < 2.0 {
				c.hihatCurve += 0.1
				c.sendOSC("/pattern/mod_rhy/hihat/curve", float32(c.hihatCurve))
			}
		}
		return true

	case "e":
		// Decrease events for active synth
		if c.activeSynth == 1 {
			if c.kickEvents > 1 {
				c.kickEvents--
				c.sendOSC("/pattern/mod_rhy/kick/events", int32(c.kickEvents))
			}
		} else if c.activeSynth == 2 {
			if c.hihatEvents > 1 {
				c.hihatEvents--
				c.sendOSC("/pattern/mod_rhy/hihat/events", int32(c.hihatEvents))
			}
		}
		return true

	case "E":
		// Increase events for active synth
		if c.activeSynth == 1 {
			if c.kickEvents < 16 {
				c.kickEvents++
				c.sendOSC("/pattern/mod_rhy/kick/events", int32(c.kickEvents))
			}
		} else if c.activeSynth == 2 {
			if c.hihatEvents < 16 {
				c.hihatEvents++
				c.sendOSC("/pattern/mod_rhy/hihat/events", int32(c.hihatEvents))
			}
		}
		return true

	case "o":
		// Decrease offset for active synth
		if c.activeSynth == 1 {
			if c.kickOffset > -(c.phraseEvents - 1) {
				c.kickOffset--
				c.sendOSC("/pattern/mod_rhy/kick/offset", int32(c.kickOffset))
			}
		} else if c.activeSynth == 2 {
			if c.hihatOffset > -(c.phraseEvents - 1) {
				c.hihatOffset--
				c.sendOSC("/pattern/mod_rhy/hihat/offset", int32(c.hihatOffset))
			}
		}
		return true

	case "O":
		// Increase offset for active synth
		if c.activeSynth == 1 {
			if c.kickOffset < (c.phraseEvents - 1) {
				c.kickOffset++
				c.sendOSC("/pattern/mod_rhy/kick/offset", int32(c.kickOffset))
			}
		} else if c.activeSynth == 2 {
			if c.hihatOffset < (c.phraseEvents - 1) {
				c.hihatOffset++
				c.sendOSC("/pattern/mod_rhy/hihat/offset", int32(c.hihatOffset))
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
		c.sendOSC("/pattern/mod_rhy/debug", debugInt)
		return true
	}

	return false
}

// Quit stops the pattern and resets to defaults
func (c *ModulatedRhythmController) Quit() {
	c.sendOSC("/pattern/mod_rhy/reset")
	c.isPlaying = false
}

// sendOSC sends an OSC message to sclang
func (c *ModulatedRhythmController) sendOSC(address string, args ...interface{}) {
	msg := osc.NewMessage(address)
	for _, arg := range args {
		msg.Append(arg)
	}

	client := osc.NewClient(c.sclangAdapter.GetHost(), c.sclangAdapter.GetPort())
	if err := client.Send(msg); err != nil {
		fmt.Printf("OSC send error: %v\n", err)
	}
}
