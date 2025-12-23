package rand

import (
	"fmt"

	"forbidden_sequencer/sequencer/events"
)

// GatedPattern fires events only when within a specified tick range of the phrase
type GatedPattern struct {
	name         string  // event name (e.g., "kick", "hihat")
	note         uint8   // MIDI note number
	velocity     float64 // event velocity
	startTick    int     // first tick in phrase when pattern fires (inclusive)
	endTick      int     // last tick in phrase when pattern fires (exclusive)
	paused       bool
	phraseLength int   // length of phrase in ticks
	tickInPhrase int   // current tick within phrase
	lastTick     int64 // last tick we saw
}

// NewGatedPattern creates a new gated pattern
// startTick: first tick to fire (inclusive)
// endTick: last tick to fire (exclusive)
func NewGatedPattern(
	name string,
	note uint8,
	velocity float64,
	startTick int,
	endTick int,
	phraseLength int,
) *GatedPattern {
	return &GatedPattern{
		name:         name,
		note:         note,
		velocity:     velocity,
		startTick:    startTick,
		endTick:      endTick,
		paused:       true,
		phraseLength: phraseLength,
		tickInPhrase: 0,
		lastTick:     -1,
	}
}

// Reset resets the pattern state
func (g *GatedPattern) Reset() {
	g.tickInPhrase = 0
	g.lastTick = -1
}

// Play resumes the pattern
func (g *GatedPattern) Play() {
	g.paused = false
}

// Stop pauses the pattern
func (g *GatedPattern) Stop() {
	g.paused = true
}

// String returns a string representation of the pattern
func (g *GatedPattern) String() string {
	return fmt.Sprintf("%s (ticks %d-%d)", g.name, g.startTick, g.endTick)
}

// updatePhrase tracks current position in phrase
func (g *GatedPattern) updatePhrase(tick int64) {
	if tick != g.lastTick {
		g.tickInPhrase++
		if g.tickInPhrase >= g.phraseLength {
			g.tickInPhrase = 0
		}
		g.lastTick = tick
	}
}

// GetEventsForTick implements the Pattern interface
func (g *GatedPattern) GetEventsForTick(tick int64) []events.TickEvent {
	// Update phrase position
	g.updatePhrase(tick)

	// When paused, return no events
	if g.paused {
		return nil
	}

	// Check if tick is in the active range
	inRange := g.tickInPhrase >= g.startTick && g.tickInPhrase < g.endTick

	if inRange {
		return []events.TickEvent{{
			Event: events.Event{
				Name: g.name,
				Type: events.EventTypeNote,
				Params: map[string]float32{
					"midi_note": float32(g.note),
					"amp":       float32(g.velocity),
				},
			},
			TickTiming: events.TickTiming{
				Tick:          tick,
				OffsetPercent: 0.0,
				DurationTicks: 0.75,
			},
		}}
	}

	// Outside range - return no events
	return nil
}
