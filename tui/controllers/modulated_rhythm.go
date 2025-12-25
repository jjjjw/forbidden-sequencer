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
	kickCurve     float64
	kickEvents    int
	hihatCurve    float64
	hihatEvents   int
	isPlaying     bool
}

// NewModulatedRhythmController creates a new modulated rhythm controller
func NewModulatedRhythmController(sclangAdapter *adapters.OSCAdapter) *ModulatedRhythmController {
	return &ModulatedRhythmController{
		sclangAdapter: sclangAdapter,
		phraseDur:     2.0,
		kickCurve:     2.0,
		kickEvents:    8,
		hihatCurve:    1.5,
		hihatEvents:   8,
		isPlaying:     false,
	}
}

// GetName returns the display name
func (c *ModulatedRhythmController) GetName() string {
	return "Ramp Time"
}

// GetKeybindings returns the controller-specific controls
func (c *ModulatedRhythmController) GetKeybindings() string {
	return "d/D: phrase duration | c/C: kick curve | v/V: hihat curve | e/E: kick events | r/R: hihat events"
}

// GetStatus returns the current state
func (c *ModulatedRhythmController) GetStatus() string {
	return fmt.Sprintf("Phrase: %.1fs | Kick: curve=%.1f, events=%d | Hihat: curve=%.1f, events=%d",
		c.phraseDur, c.kickCurve, c.kickEvents, c.hihatCurve, c.hihatEvents)
}

// HandleInput processes controller-specific input
func (c *ModulatedRhythmController) HandleInput(msg tea.KeyMsg) bool {
	switch msg.String() {
	case " ", "p":
		// Toggle play/pause
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
			c.kickCurve -= 0.5
			c.sendOSC("/pattern/mod_rhy/kick/curve", float32(c.kickCurve))
		}
		return true

	case "C":
		// Increase kick curve
		if c.kickCurve < 5.0 {
			c.kickCurve += 0.5
			c.sendOSC("/pattern/mod_rhy/kick/curve", float32(c.kickCurve))
		}
		return true

	case "v":
		// Decrease hihat curve
		if c.hihatCurve > 0.5 {
			c.hihatCurve -= 0.5
			c.sendOSC("/pattern/mod_rhy/hihat/curve", float32(c.hihatCurve))
		}
		return true

	case "V":
		// Increase hihat curve
		if c.hihatCurve < 5.0 {
			c.hihatCurve += 0.5
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
	}

	return false
}

// Quit stops the pattern and cleans up
func (c *ModulatedRhythmController) Quit() {
	c.sendOSC("/pattern/mod_rhy/stop")
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
