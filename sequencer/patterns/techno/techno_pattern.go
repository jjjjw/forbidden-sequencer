package techno

import (
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// TechnoPattern generates alternating kick and hihat events
type TechnoPattern struct {
	conductor *conductors.CommonTimeConductor
	isKick    bool // true = kick next, false = hihat next
	paused    bool
}

// NewTechnoPattern creates a new techno pattern
func NewTechnoPattern(c *conductors.CommonTimeConductor) *TechnoPattern {
	return &TechnoPattern{
		conductor: c,
		isKick:    true, // start with kick
		paused:    true,
	}
}

// Reset resets the pattern state to start with kick
func (t *TechnoPattern) Reset() {
	t.isKick = true
}

// Play resumes the pattern
func (t *TechnoPattern) Play() {
	t.paused = false
	// t.isKick = true
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
				Timestamp: time.Now().Add(10 * time.Millisecond),
				Duration:  0,
			},
		}, nil
	}

	tickDuration := t.conductor.GetTickDuration()
	ticksPerBeat := t.conductor.GetTicksPerBeat()
	beatDuration := tickDuration * time.Duration(ticksPerBeat)
	halfBeatDuration := beatDuration / 2

	// Get next beat time from conductor
	nextBeatTime := t.conductor.GetNextBeatTime()

	var nextFireTime time.Time
	if t.isKick {
		// Kick fires on next beat boundary
		nextFireTime = nextBeatTime
	} else {
		// Hihat fires half a beat before the next beat (i.e., half beat after last kick)
		nextFireTime = nextBeatTime.Add(-halfBeatDuration)
	}

	var event events.ScheduledEvent
	if t.isKick {
		// Create kick event (MIDI note 36 = bass drum)
		event = events.ScheduledEvent{
			Event: events.Event{
				Name: "kick",
				Type: events.EventTypeNote,
				A:    36,  // MIDI note number
				B:    0.8, // velocity
			},
			Timing: events.Timing{
				Timestamp: nextFireTime,
				Duration:  100 * time.Millisecond,
			},
		}
	} else {
		// Create hihat event (MIDI note 42 = closed hihat)
		event = events.ScheduledEvent{
			Event: events.Event{
				Name: "hihat",
				Type: events.EventTypeNote,
				A:    42,  // MIDI note number
				B:    0.6, // velocity
			},
			Timing: events.Timing{
				Timestamp: nextFireTime,
				Duration:  50 * time.Millisecond,
			},
		}
	}

	// Toggle for next call
	t.isKick = !t.isKick

	return event, nil
}
