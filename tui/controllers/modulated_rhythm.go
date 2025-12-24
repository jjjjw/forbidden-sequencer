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
	tickDur       float64
	isPlaying     bool
}

// NewModulatedRhythmController creates a new modulated rhythm controller
func NewModulatedRhythmController(sclangAdapter *adapters.OSCAdapter) *ModulatedRhythmController {
	return &ModulatedRhythmController{
		sclangAdapter: sclangAdapter,
		tickDur:       0.125,
		isPlaying:     false,
	}
}

// GetName returns the display name
func (c *ModulatedRhythmController) GetName() string {
	return "Ramp Time"
}

// GetKeybindings returns the controller-specific controls
func (c *ModulatedRhythmController) GetKeybindings() string {
	return "t/T: tick duration"
}

// GetStatus returns the current state
func (c *ModulatedRhythmController) GetStatus() string {
	return fmt.Sprintf("Tick: %.0fms", c.tickDur*1000)
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

	case "t":
		// Decrease tick duration
		if c.tickDur > 0.05 {
			c.tickDur -= 0.01
			c.sendOSC("/pattern/mod_rhy/tick_dur", float32(c.tickDur))
		}
		return true

	case "T":
		// Increase tick duration
		if c.tickDur < 0.5 {
			c.tickDur += 0.01
			c.sendOSC("/pattern/mod_rhy/tick_dur", float32(c.tickDur))
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
