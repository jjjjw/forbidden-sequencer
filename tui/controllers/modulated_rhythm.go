package controllers

import (
	"fmt"

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
d/D: phrase dur
1/!: phrase events
c/C: kick curve
v/V: hihat curve
e/E: kick events
r/R: hihat events
[/]: kick offset
o/O: hihat offset
x: debug`
}

// GetStatus returns the current state
func (c *ModulatedRhythmController) GetStatus() string {
	debugStr := ""
	if c.debug {
		debugStr = " | DEBUG"
	}
	return fmt.Sprintf("Phrase: %.1fs, %d events | Kick: curve=%.1f, events=%d, offset=%+d | Hihat: curve=%.1f, events=%d, offset=%+d%s",
		c.phraseDur, c.phraseEvents, c.kickCurve, c.kickEvents, c.kickOffset, c.hihatCurve, c.hihatEvents, c.hihatOffset, debugStr)
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

	case "c":
		// Decrease kick curve
		if c.kickCurve > 0.5 {
			c.kickCurve -= 0.1
			c.sendOSC("/pattern/mod_rhy/kick/curve", float32(c.kickCurve))
		}
		return true

	case "C":
		// Increase kick curve
		if c.kickCurve < 2.0 {
			c.kickCurve += 0.1
			c.sendOSC("/pattern/mod_rhy/kick/curve", float32(c.kickCurve))
		}
		return true

	case "v":
		// Decrease hihat curve
		if c.hihatCurve > 0.5 {
			c.hihatCurve -= 0.1
			c.sendOSC("/pattern/mod_rhy/hihat/curve", float32(c.hihatCurve))
		}
		return true

	case "V":
		// Increase hihat curve
		if c.hihatCurve < 2.0 {
			c.hihatCurve += 0.1
			c.sendOSC("/pattern/mod_rhy/hihat/curve", float32(c.hihatCurve))
		}
		return true

	case "e":
		// Decrease kick events
		if c.kickEvents > 1 {
			c.kickEvents--
			c.sendOSC("/pattern/mod_rhy/kick/events", int32(c.kickEvents))
		}
		return true

	case "E":
		// Increase kick events
		if c.kickEvents < 16 {
			c.kickEvents++
			c.sendOSC("/pattern/mod_rhy/kick/events", int32(c.kickEvents))
		}
		return true

	case "r":
		// Decrease hihat events
		if c.hihatEvents > 1 {
			c.hihatEvents--
			c.sendOSC("/pattern/mod_rhy/hihat/events", int32(c.hihatEvents))
		}
		return true

	case "R":
		// Increase hihat events
		if c.hihatEvents < 16 {
			c.hihatEvents++
			c.sendOSC("/pattern/mod_rhy/hihat/events", int32(c.hihatEvents))
		}
		return true

	case "1":
		// Decrease phrase events
		if c.phraseEvents > 16 {
			c.phraseEvents--
			c.sendOSC("/pattern/mod_rhy/phrase_events", int32(c.phraseEvents))
		}
		return true

	case "!":
		// Increase phrase events
		if c.phraseEvents < 32 {
			c.phraseEvents++
			c.sendOSC("/pattern/mod_rhy/phrase_events", int32(c.phraseEvents))
		}
		return true

	case "[":
		// Decrease kick offset
		if c.kickOffset > -(c.phraseEvents - 1) {
			c.kickOffset--
			c.sendOSC("/pattern/mod_rhy/kick/offset", int32(c.kickOffset))
		}
		return true

	case "]":
		// Increase kick offset
		if c.kickOffset < (c.phraseEvents - 1) {
			c.kickOffset++
			c.sendOSC("/pattern/mod_rhy/kick/offset", int32(c.kickOffset))
		}
		return true

	case "o":
		// Decrease hihat offset
		if c.hihatOffset > -(c.phraseEvents - 1) {
			c.hihatOffset--
			c.sendOSC("/pattern/mod_rhy/hihat/offset", int32(c.hihatOffset))
		}
		return true

	case "O":
		// Increase hihat offset
		if c.hihatOffset < (c.phraseEvents - 1) {
			c.hihatOffset++
			c.sendOSC("/pattern/mod_rhy/hihat/offset", int32(c.hihatOffset))
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
