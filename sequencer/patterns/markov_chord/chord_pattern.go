package markov_chord

import (

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// SectionType represents the current section
type SectionType int

const (
	SectionChord SectionType = iota
	SectionPercussion
)

// ChordPattern plays 4-note chords and manages section state
type ChordPattern struct {
	conductor         *conductors.Conductor
	rootNote          uint8     // root MIDI note
	scale             lib.Scale // scale for chord notes
	velocity          float64   // chord velocity
	paused            bool
	currentSection    SectionType // current section (chord or percussion)
	phraseCounter     int         // counts phrases in current section
	phrasesPerSection int         // how many phrases per section
	phraseLength      int         // length of phrase in ticks
	tickInPhrase      int         // current tick within phrase
	lastTick          int64       // last tick we saw
	chordPlayed       bool        // tracks if chord was played this section
}

// NewChordPattern creates a new chord pattern
func NewChordPattern(
	conductor *conductors.Conductor,
	rootNote uint8,
	scale lib.Scale,
	velocity float64,
	phraseLength int,
	phrasesPerSection int,
) *ChordPattern {
	return &ChordPattern{
		conductor:         conductor,
		rootNote:          rootNote,
		scale:             scale,
		velocity:          velocity,
		paused:            true,
		currentSection:    SectionChord, // Start with chord
		phraseCounter:     0,
		phrasesPerSection: phrasesPerSection,
		phraseLength:      phraseLength,
		tickInPhrase:      0,
		lastTick:          -1,
		chordPlayed:       false,
	}
}

// Reset resets the pattern state
func (c *ChordPattern) Reset() {
	c.currentSection = SectionChord
	c.phraseCounter = 0
	c.tickInPhrase = 0
	c.lastTick = -1
	c.chordPlayed = false
}

// Play resumes the pattern
func (c *ChordPattern) Play() {
	c.paused = false
}

// Stop pauses the pattern
func (c *ChordPattern) Stop() {
	c.paused = true
}

// String returns a string representation
func (c *ChordPattern) String() string {
	return "chord (4-voice)"
}

// GetCurrentSection returns the current section type
func (c *ChordPattern) GetCurrentSection() SectionType {
	return c.currentSection
}

// IsChordSection returns true if in chord section
func (c *ChordPattern) IsChordSection() bool {
	return c.currentSection == SectionChord
}

// IsPercussionSection returns true if in percussion section
func (c *ChordPattern) IsPercussionSection() bool {
	return c.currentSection == SectionPercussion
}

// updateSection checks if we should switch sections based on tick
func (c *ChordPattern) updateSection(tick int64) {
	// Detect if we've moved to a new tick
	if tick != c.lastTick {
		c.tickInPhrase++

		// Detect phrase boundary
		if c.tickInPhrase >= c.phraseLength {
			c.tickInPhrase = 0
			c.phraseCounter++

			// Switch sections after phrasesPerSection
			if c.phraseCounter >= c.phrasesPerSection {
				if c.currentSection == SectionChord {
					c.currentSection = SectionPercussion
				} else {
					c.currentSection = SectionChord
					c.chordPlayed = false // Reset for new chord section
				}
				c.phraseCounter = 0
			}
		}

		c.lastTick = tick
	}
}

// GetEventsForTick implements the Pattern interface
func (c *ChordPattern) GetEventsForTick(tick int64) []events.TickEvent {
	// Update section state
	c.updateSection(tick)

	// When paused, return no events
	if c.paused {
		return nil
	}

	// Only play during chord section
	if c.currentSection != SectionChord {
		return nil
	}

	// Only play chord once per section
	if c.chordPlayed {
		return nil
	}

	// Play chord only on first tick of phrase
	if c.tickInPhrase != 0 {
		return nil
	}

	// Mark chord as played
	c.chordPlayed = true

	// Create 4-note chord (root, 3rd, 5th, 7th from scale)
	chordDegrees := []int{0, 2, 4, 6} // I, iii, V, vii in minor scale
	// Subtle modulation for smooth, techno sound
	modIndices := []float32{0.3, 0.4, 0.35, 0.38}

	var tickEvents []events.TickEvent
	for i, degree := range chordDegrees {
		midiNote := c.scale.NoteAt(c.rootNote, degree)
		tickEvents = append(tickEvents, events.TickEvent{
			Event: events.Event{
				Name: "fm",
				Type: events.EventTypeNote,
				Params: map[string]float32{
					"midi_note":  float32(midiNote),
					"amp":        float32(c.velocity / 4),
					"modRatio":   1.0,           // Unison for smooth, warm tone
					"modIndex":   modIndices[i], // Subtle variation
					"max_voices": 4,
				},
			},
			Tick:          tick,
			OffsetPercent: 0.0,                    // On the beat
			DurationTicks: float64(c.phraseLength), // Full phrase duration
		})
	}

	return tickEvents
}
