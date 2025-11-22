package techno

import (
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// TechnoPattern generates alternating kick and hihat events
type TechnoPattern struct {
	conductor    *conductors.CommonTimeConductor
	isKick       bool  // true = kick next, false = hihat next
	lastBeatTick int64 // the beat tick we're currently working on
	paused       bool
}

// NewTechnoPattern creates a new techno pattern
func NewTechnoPattern(c *conductors.CommonTimeConductor) *TechnoPattern {
	return &TechnoPattern{
		conductor:    c,
		isKick:       true, // start with kick
		lastBeatTick: 0,
		paused:       true,
	}
}

// Reset resets the pattern state to start with kick
func (t *TechnoPattern) Reset() {
	t.isKick = true
	t.lastBeatTick = t.conductor.GetNextBeatTick()
}

// Play resumes the pattern
func (t *TechnoPattern) Play() {
	t.paused = false
}

// Stop pauses the pattern
func (t *TechnoPattern) Stop() {
	t.paused = true
}

// GetNextScheduledEvent implements the Pattern interface
func (t *TechnoPattern) GetNextScheduledEvent() (events.ScheduledEvent, error) {
	// When paused, return short rests
	if t.paused {
		return events.ScheduledEvent{
			Event: events.Event{
				Name: "rest",
				Type: events.EventTypeRest,
			},
			Timing: events.Timing{
				Delta:    10 * time.Millisecond,
				Duration: 0,
			},
		}, nil
	}

	ticksPerBeat := t.conductor.GetTicksPerBeat()
	halfBeat := ticksPerBeat / 2

	var nextFireTick int64
	if t.isKick {
		// Kick fires on next beat boundary after lastBeatTick
		nextBeat := t.conductor.GetNextBeatTick()
		if nextBeat <= t.lastBeatTick {
			// Already scheduled this beat, advance to next one
			nextBeat = t.lastBeatTick + int64(ticksPerBeat)
		}
		t.lastBeatTick = nextBeat
		nextFireTick = t.lastBeatTick
	} else {
		// Hihat fires half a beat after the kick
		nextFireTick = t.lastBeatTick + int64(halfBeat)
	}

	// Calculate absolute wall-clock time for this tick (drift-free)
	nextFireTime := t.conductor.GetAbsoluteTimeForTick(nextFireTick)
	timeDelta := time.Until(nextFireTime)

	var event events.ScheduledEvent
	if t.isKick {
		// Create kick event (MIDI note 36 = bass drum)
		event = events.ScheduledEvent{
			Event: events.Event{
				Name: "kick",
				Type: events.EventTypeNote,
				A:    0.8, // velocity
			},
			Timing: events.Timing{
				Delta:    timeDelta,
				Duration: 100 * time.Millisecond,
			},
		}
	} else {
		// Create hihat event (MIDI note 42 = closed hihat)
		event = events.ScheduledEvent{
			Event: events.Event{
				Name: "hihat",
				Type: events.EventTypeNote,
				A:    0.6, // velocity
			},
			Timing: events.Timing{
				Delta:    timeDelta,
				Duration: 50 * time.Millisecond,
			},
		}
	}

	// Toggle for next call
	t.isKick = !t.isKick

	return event, nil
}
