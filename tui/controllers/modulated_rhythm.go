package controllers

import (
	"fmt"

	"forbidden_sequencer/sequencer/adapters"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hypebeast/go-osc/osc"
)

// ModulatedRhythmController controls the mod_rhy pattern in sclang via OSC
type ModulatedRhythmController struct {
	sclangAdapter    *adapters.OSCAdapter
	kickSubdivision  int
	hihatSubdivision int
	kickCurve        float64
	kickEvents       int
	hihatCurve       float64
	hihatEvents      int
	kickQuantized    bool
	hihatQuantized   bool
	isPlaying        bool
}

// NewModulatedRhythmController creates a new modulated rhythm controller
func NewModulatedRhythmController(sclangAdapter *adapters.OSCAdapter) *ModulatedRhythmController {
	return &ModulatedRhythmController{
		sclangAdapter:    sclangAdapter,
		kickSubdivision:  1,
		hihatSubdivision: 1,
		kickCurve:        2.0,
		kickEvents:       8,
		hihatCurve:       1.5,
		hihatEvents:      6,
		kickQuantized:    true,
		hihatQuantized:   true,
		isPlaying:        false,
	}
}

// GetName returns the display name
func (c *ModulatedRhythmController) GetName() string {
	return "Ramp Time"
}

// GetKeybindings returns the controller-specific controls
func (c *ModulatedRhythmController) GetKeybindings() string {
	return "h/l: kick subdiv | H/L: hihat subdiv | c/C: kick curve | v/V: hihat curve | e/E: kick events | r/R: hihat events | z/Z: toggle quantize"
}

// GetStatus returns the current state
func (c *ModulatedRhythmController) GetStatus() string {
	return fmt.Sprintf("Kick: %dx (curve=%.1f, events=%d) | Hihat: %dx (curve=%.1f, events=%d)",
		c.kickSubdivision, c.kickCurve, c.kickEvents, c.hihatSubdivision, c.hihatCurve, c.hihatEvents)
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

	case "h":
		// Decrease kick subdivision
		if c.kickSubdivision > 1 {
			c.kickSubdivision--
			c.sendOSC("/pattern/mod_rhy/kick/subdivision", int32(c.kickSubdivision))
		}
		return true

	case "l":
		// Increase kick subdivision
		c.kickSubdivision++
		c.sendOSC("/pattern/mod_rhy/kick/subdivision", int32(c.kickSubdivision))
		return true

	case "H":
		// Decrease hihat subdivision
		if c.hihatSubdivision > 1 {
			c.hihatSubdivision--
			c.sendOSC("/pattern/mod_rhy/hihat/subdivision", int32(c.hihatSubdivision))
		}
		return true

	case "L":
		// Increase hihat subdivision
		c.hihatSubdivision++
		c.sendOSC("/pattern/mod_rhy/hihat/subdivision", int32(c.hihatSubdivision))
		return true

	case "c":
		// Decrease kick curve (less ritardando)
		if c.kickCurve > 0.5 {
			c.kickCurve -= 0.5
			c.sendOSC("/pattern/mod_rhy/kick/curve", float32(c.kickCurve))
		}
		return true

	case "C":
		// Increase kick curve (more ritardando)
		if c.kickCurve < 5.0 {
			c.kickCurve += 0.5
			c.sendOSC("/pattern/mod_rhy/kick/curve", float32(c.kickCurve))
		}
		return true

	case "v":
		// Decrease hihat curve (less ritardando)
		if c.hihatCurve > 0.5 {
			c.hihatCurve -= 0.5
			c.sendOSC("/pattern/mod_rhy/hihat/curve", float32(c.hihatCurve))
		}
		return true

	case "V":
		// Increase hihat curve (more ritardando)
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

	case "z":
		// Toggle kick quantization
		c.kickQuantized = !c.kickQuantized
		quantizedInt := int32(0)
		if c.kickQuantized {
			quantizedInt = 1
		}
		c.sendOSC("/pattern/mod_rhy/kick/quantized", quantizedInt)
		return true

	case "Z":
		// Toggle hihat quantization
		c.hihatQuantized = !c.hihatQuantized
		quantizedInt := int32(0)
		if c.hihatQuantized {
			quantizedInt = 1
		}
		c.sendOSC("/pattern/mod_rhy/hihat/quantized", quantizedInt)
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
