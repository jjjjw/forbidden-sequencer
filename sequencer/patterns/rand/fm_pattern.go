package modulated

import (
	"fmt"
	"math/rand"
	"time"

	"forbidden_sequencer/sequencer/events"
	"forbidden_sequencer/sequencer/lib"
)

// FMPattern generates melodic patterns for 2-op FM synthesis
// Randomizes: pitch (from scale), duration, FM ratio, and FM modulation index
type FMPattern struct {
	name        string    // event name (fm1 or fm2)
	velocity    float64   // event velocity
	rootNote    uint8     // root MIDI note
	scale       lib.Scale // scale to use for pitch generation
	octaveRange int       // range of octaves to use (e.g., 2 = 2 octaves)
	paused      bool
	rng         *rand.Rand
}

// NewFMPattern creates a new FM pattern
// rootNote: root MIDI note for the scale
// scale: scale to use for pitch generation
// octaveRange: number of octaves to span
func NewFMPattern(
	name string,
	velocity float64,
	rootNote uint8,
	scale lib.Scale,
	octaveRange int,
	seed int64,
) *FMPattern {
	return &FMPattern{
		name:        name,
		velocity:    velocity,
		rootNote:    rootNote,
		scale:       scale,
		octaveRange: octaveRange,
		paused:      true,
		rng:         rand.New(rand.NewSource(seed)),
	}
}

// Reset resets the pattern state
func (f *FMPattern) Reset() {
	// No state to reset currently
}

// Play resumes the pattern
func (f *FMPattern) Play() {
	f.paused = false
}

// Stop pauses the pattern
func (f *FMPattern) Stop() {
	f.paused = true
}

// String returns a string representation of the pattern
func (f *FMPattern) String() string {
	return fmt.Sprintf("%s FM (random)", f.name)
}

// GetScheduledEventsForTick implements the Pattern interface
func (f *FMPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if f.paused {
		return nil
	}

	// TODO: Use Markov chain instead of random for triggering decisions
	// For now, random trigger with 30% probability
	if f.rng.Float64() > 0.3 {
		return nil
	}

	// Randomize pitch: select random scale degree within octave range
	scaleLength := len(f.scale)
	maxDegree := scaleLength * f.octaveRange
	degree := f.rng.Intn(maxDegree)
	midiNote := f.scale.NoteAt(f.rootNote, degree)

	// Randomize duration: 75% to 200% of tick duration
	durationFactor := 0.75 + f.rng.Float64()*1.25 // 0.75 to 2.0
	noteDuration := time.Duration(float64(tickDuration) * durationFactor)

	// Randomize FM ratio: common ratios for harmonicity
	// Ratios: 0.5, 1, 1.5, 2, 3, 4, 5, 7
	ratios := []float32{0.5, 1.0, 1.5, 2.0, 3.0, 4.0, 5.0, 7.0}
	modRatio := ratios[f.rng.Intn(len(ratios))]

	// Randomize FM modulation index (depth): 0.1 to 3.0
	modIndex := float32(0.1 + f.rng.Float64()*2.9) // 0.1 to 3.0

	return []events.ScheduledEvent{{
		Event: events.Event{
			Name: f.name,
			Type: events.EventTypeNote,
			A:    float32(midiNote), // MIDI note (will be converted to freq by adapter)
			B:    float32(f.velocity), // amplitude
			C:    modRatio, // FM ratio (C parameter)
			D:    modIndex, // FM modulation index (D parameter)
		},
		Timing: events.Timing{
			Timestamp: nextTickTime,
			Duration:  noteDuration,
		},
	}}
}
