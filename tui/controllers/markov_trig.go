package controllers

import (
	"fmt"
	"strings"

	"forbidden_sequencer/adapter"

	tea "github.com/charmbracelet/bubbletea"
)

// MarkovTrigController controls the markov_trig pattern in sclang via OSC
type MarkovTrigController struct {
	sclangAdapter *adapter.OSCAdapter
	baseEventDur  float64
	phraseLength  int
	kickProb      float64
	snareProb     float64
	hihatProb     float64
	fm1Prob       float64
	fm2Prob       float64
	debug         bool
	isPlaying     bool
	activeSynth   int // 0=kick, 1=snare, 2=hihat, 3=fm1, 4=fm2
}

// NewMarkovTrigController creates a new markov triggers controller
func NewMarkovTrigController(sclangAdapter *adapter.OSCAdapter) *MarkovTrigController {
	return &MarkovTrigController{
		sclangAdapter: sclangAdapter,
		baseEventDur:  0.125,
		phraseLength:  16,
		kickProb:      0.5,
		snareProb:     0.5,
		hihatProb:     0.5,
		fm1Prob:       0.3,
		fm2Prob:       0.3,
		debug:         false,
		isPlaying:     false,
		activeSynth:   0, // Start with kick selected
	}
}

// GetName returns the display name
func (c *MarkovTrigController) GetName() string {
	return "Markov Triggers"
}

// GetKeybindings returns the controller-specific controls
func (c *MarkovTrigController) GetKeybindings() string {
	return `p: play/stop
space: pause/resume
1-5: select synth
d/D: base event dur
r/R: phrase length
e/E: probability (active synth)
x: debug`
}

// GetStatus returns the current state
func (c *MarkovTrigController) GetStatus() string {
	var status strings.Builder

	// Phrase state
	phraseDur := float64(c.phraseLength) * c.baseEventDur
	status.WriteString(fmt.Sprintf("Base: %.3fs, Length: %d, Phrase: %.2fs\n\n", c.baseEventDur, c.phraseLength, phraseDur))

	// Synths section
	status.WriteString("Synths:\n")

	// Kick
	kickPrefix := "  "
	if c.activeSynth == 0 {
		kickPrefix = "> "
	}
	status.WriteString(fmt.Sprintf("%s1. Kick: %.0f%%\n", kickPrefix, c.kickProb*100))

	// Snare
	snarePrefix := "  "
	if c.activeSynth == 1 {
		snarePrefix = "> "
	}
	status.WriteString(fmt.Sprintf("%s2. Snare: %.0f%%\n", snarePrefix, c.snareProb*100))

	// Hihat
	hihatPrefix := "  "
	if c.activeSynth == 2 {
		hihatPrefix = "> "
	}
	status.WriteString(fmt.Sprintf("%s3. Hihat: %.0f%%\n", hihatPrefix, c.hihatProb*100))

	// FM1
	fm1Prefix := "  "
	if c.activeSynth == 3 {
		fm1Prefix = "> "
	}
	status.WriteString(fmt.Sprintf("%s4. FM1: %.0f%%\n", fm1Prefix, c.fm1Prob*100))

	// FM2
	fm2Prefix := "  "
	if c.activeSynth == 4 {
		fm2Prefix = "> "
	}
	status.WriteString(fmt.Sprintf("%s5. FM2: %.0f%%", fm2Prefix, c.fm2Prob*100))

	if c.debug {
		status.WriteString("\nDEBUG")
	}

	return status.String()
}

// HandleInput processes controller-specific input
func (c *MarkovTrigController) HandleInput(msg tea.KeyMsg) bool {
	switch msg.String() {
	case " ":
		// Toggle pause/resume
		if c.isPlaying {
			c.sclangAdapter.Send("/pattern/markov_trig/pause")
			c.isPlaying = false
		} else {
			c.sclangAdapter.Send("/pattern/markov_trig/resume")
			c.isPlaying = true
		}
		return true

	case "p":
		// Toggle play/stop (reset position)
		if c.isPlaying {
			c.sclangAdapter.Send("/pattern/markov_trig/stop")
			c.isPlaying = false
		} else {
			c.sclangAdapter.Send("/pattern/markov_trig/play")
			c.isPlaying = true
		}
		return true

	case "1":
		// Select kick synth
		c.activeSynth = 0
		return true

	case "2":
		// Select snare synth
		c.activeSynth = 1
		return true

	case "3":
		// Select hihat synth
		c.activeSynth = 2
		return true

	case "4":
		// Select fm1 synth
		c.activeSynth = 3
		return true

	case "5":
		// Select fm2 synth
		c.activeSynth = 4
		return true

	case "d":
		// Decrease base event duration
		if c.baseEventDur > 0.025 {
			c.baseEventDur -= 0.005
			c.sclangAdapter.Send("/pattern/markov_trig/base_event_dur", float32(c.baseEventDur))
		}
		return true

	case "D":
		// Increase base event duration
		if c.baseEventDur < 1.0 {
			c.baseEventDur += 0.005
			c.sclangAdapter.Send("/pattern/markov_trig/base_event_dur", float32(c.baseEventDur))
		}
		return true

	case "r":
		// Decrease phrase length
		if c.phraseLength > 4 {
			c.phraseLength--
			c.sclangAdapter.Send("/pattern/markov_trig/phrase_length", int32(c.phraseLength))
		}
		return true

	case "R":
		// Increase phrase length
		if c.phraseLength < 64 {
			c.phraseLength++
			c.sclangAdapter.Send("/pattern/markov_trig/phrase_length", int32(c.phraseLength))
		}
		return true

	case "e":
		// Decrease probability for active synth
		switch c.activeSynth {
		case 0: // kick
			if c.kickProb > 0.0 {
				c.kickProb -= 0.1
				if c.kickProb < 0.0 {
					c.kickProb = 0.0
				}
				c.sclangAdapter.Send("/pattern/markov_trig/kick/prob", float32(c.kickProb))
			}
		case 1: // snare
			if c.snareProb > 0.0 {
				c.snareProb -= 0.1
				if c.snareProb < 0.0 {
					c.snareProb = 0.0
				}
				c.sclangAdapter.Send("/pattern/markov_trig/snare/prob", float32(c.snareProb))
			}
		case 2: // hihat
			if c.hihatProb > 0.0 {
				c.hihatProb -= 0.1
				if c.hihatProb < 0.0 {
					c.hihatProb = 0.0
				}
				c.sclangAdapter.Send("/pattern/markov_trig/hihat/prob", float32(c.hihatProb))
			}
		case 3: // fm1
			if c.fm1Prob > 0.0 {
				c.fm1Prob -= 0.1
				if c.fm1Prob < 0.0 {
					c.fm1Prob = 0.0
				}
				c.sclangAdapter.Send("/pattern/markov_trig/fm1/prob", float32(c.fm1Prob))
			}
		case 4: // fm2
			if c.fm2Prob > 0.0 {
				c.fm2Prob -= 0.1
				if c.fm2Prob < 0.0 {
					c.fm2Prob = 0.0
				}
				c.sclangAdapter.Send("/pattern/markov_trig/fm2/prob", float32(c.fm2Prob))
			}
		}
		return true

	case "E":
		// Increase probability for active synth
		switch c.activeSynth {
		case 0: // kick
			if c.kickProb < 1.0 {
				c.kickProb += 0.1
				if c.kickProb > 1.0 {
					c.kickProb = 1.0
				}
				c.sclangAdapter.Send("/pattern/markov_trig/kick/prob", float32(c.kickProb))
			}
		case 1: // snare
			if c.snareProb < 1.0 {
				c.snareProb += 0.1
				if c.snareProb > 1.0 {
					c.snareProb = 1.0
				}
				c.sclangAdapter.Send("/pattern/markov_trig/snare/prob", float32(c.snareProb))
			}
		case 2: // hihat
			if c.hihatProb < 1.0 {
				c.hihatProb += 0.1
				if c.hihatProb > 1.0 {
					c.hihatProb = 1.0
				}
				c.sclangAdapter.Send("/pattern/markov_trig/hihat/prob", float32(c.hihatProb))
			}
		case 3: // fm1
			if c.fm1Prob < 1.0 {
				c.fm1Prob += 0.1
				if c.fm1Prob > 1.0 {
					c.fm1Prob = 1.0
				}
				c.sclangAdapter.Send("/pattern/markov_trig/fm1/prob", float32(c.fm1Prob))
			}
		case 4: // fm2
			if c.fm2Prob < 1.0 {
				c.fm2Prob += 0.1
				if c.fm2Prob > 1.0 {
					c.fm2Prob = 1.0
				}
				c.sclangAdapter.Send("/pattern/markov_trig/fm2/prob", float32(c.fm2Prob))
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
		c.sclangAdapter.Send("/pattern/markov_trig/debug", debugInt)
		return true
	}

	return false
}

// Quit stops the pattern and resets to defaults
func (c *MarkovTrigController) Quit() {
	c.sclangAdapter.Send("/pattern/markov_trig/reset")
	c.isPlaying = false
}
