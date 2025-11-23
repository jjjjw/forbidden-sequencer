package techno

import (
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// TechnoPattern generates kick and hihat events on beats
type TechnoPattern struct {
	conductor *conductors.CommonTimeConductor
	paused    bool
}

// NewTechnoPattern creates a new techno pattern
func NewTechnoPattern(c *conductors.CommonTimeConductor) *TechnoPattern {
	return &TechnoPattern{
		conductor: c,
		paused:    true,
	}
}

// Reset resets the pattern state
func (t *TechnoPattern) Reset() {
	// No state to reset
}

// Play resumes the pattern
func (t *TechnoPattern) Play() {
	t.paused = false
}

// Stop pauses the pattern
func (t *TechnoPattern) Stop() {
	t.paused = true
}

// GetScheduledEventsForTick implements the Pattern interface
func (t *TechnoPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if t.paused {
		return nil
	}

	// Only schedule on beat boundaries (next tick is 0 in beat)
	nextTickInBeat := t.conductor.GetNextTickInBeat()
	if nextTickInBeat != 0 {
		return nil
	}

	// Calculate half beat duration for hihat offset
	ticksPerBeat := t.conductor.GetTicksPerBeat()
	halfBeatDuration := tickDuration * time.Duration(ticksPerBeat) / 2

	// Schedule both kick and hihat for this beat
	return []events.ScheduledEvent{
		{
			// Kick on the beat
			Event: events.Event{
				Name: "kick",
				Type: events.EventTypeNote,
				A:    36,  // MIDI note number
				B:    0.8, // velocity
			},
			Timing: events.Timing{
				Timestamp: nextTickTime,
				Duration:  100 * time.Millisecond,
			},
		},
		{
			// Hihat on the offbeat
			Event: events.Event{
				Name: "hihat",
				Type: events.EventTypeNote,
				A:    42,  // MIDI note number
				B:    0.6, // velocity
			},
			Timing: events.Timing{
				Timestamp: nextTickTime.Add(halfBeatDuration),
				Duration:  50 * time.Millisecond,
			},
		},
	}
}
