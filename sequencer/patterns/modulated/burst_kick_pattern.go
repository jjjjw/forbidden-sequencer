package modulated

import (
	"fmt"
	"math/rand"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// BurstKickPattern fires kick events in bursts of 3-4 hits, then pauses for 2-3 ticks
// Burst and pause lengths are randomly chosen each time a burst completes
type BurstKickPattern struct {
	conductor         *conductors.PhraseConductor
	name              string  // event name
	note              uint8   // MIDI note number
	velocity          float64 // event velocity
	paused            bool
	inBurst           bool // true if currently in a burst, false if in pause
	ticksInCurrentPhase int    // how many ticks we've been in current burst/pause
	currentBurstLength  int    // length of current burst (3-4)
	currentPauseLength  int    // length of current pause (2-3)
	lastSeenTick        int    // for detecting phrase boundaries
}

// NewBurstKickPattern creates a new burst kick pattern
func NewBurstKickPattern(
	conductor *conductors.PhraseConductor,
	name string,
	note uint8,
	velocity float64,
) *BurstKickPattern {
	p := &BurstKickPattern{
		conductor:    conductor,
		name:         name,
		note:         note,
		velocity:     velocity,
		paused:       true,
		inBurst:      true,
		lastSeenTick: -1,
	}
	p.chooseBurstAndPause()
	return p
}

// chooseBurstAndPause randomly selects burst length (3-4) and pause length (2-3)
func (b *BurstKickPattern) chooseBurstAndPause() {
	b.currentBurstLength = 3 + rand.Intn(2)  // 3 or 4
	b.currentPauseLength = 2 + rand.Intn(2)  // 2 or 3
}

// Reset resets the pattern state
func (b *BurstKickPattern) Reset() {
	b.inBurst = true
	b.ticksInCurrentPhase = 0
	b.lastSeenTick = -1
	b.chooseBurstAndPause()
}

// Play resumes the pattern
func (b *BurstKickPattern) Play() {
	b.paused = false
}

// Stop pauses the pattern
func (b *BurstKickPattern) Stop() {
	b.paused = true
}

// String returns a string representation of the pattern
func (b *BurstKickPattern) String() string {
	return fmt.Sprintf("%s (burst %d, pause %d)", b.name, b.currentBurstLength, b.currentPauseLength)
}

// updateBurstState advances the burst/pause state machine
func (b *BurstKickPattern) updateBurstState() {
	b.ticksInCurrentPhase++

	if b.inBurst {
		// Check if burst is complete
		if b.ticksInCurrentPhase >= b.currentBurstLength {
			b.inBurst = false
			b.ticksInCurrentPhase = 0
		}
	} else {
		// Check if pause is complete
		if b.ticksInCurrentPhase >= b.currentPauseLength {
			b.inBurst = true
			b.ticksInCurrentPhase = 0
			// Choose new burst and pause lengths for the next cycle
			b.chooseBurstAndPause()
		}
	}
}

// checkForPhraseReset detects phrase boundaries and resets state if needed
func (b *BurstKickPattern) checkForPhraseReset() {
	currentTick := b.conductor.GetNextTickInPhrase()

	// Detect phrase wrap (tick went backwards)
	if b.lastSeenTick != -1 && currentTick < b.lastSeenTick {
		b.Reset()
	}

	b.lastSeenTick = currentTick
}

// GetScheduledEventsForTick implements the Pattern interface
func (b *BurstKickPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	// When paused, return no events
	if b.paused {
		return nil
	}

	// Check for phrase reset
	b.checkForPhraseReset()

	// Only fire during first 50% of phrase
	nextTickInPhrase := b.conductor.GetNextTickInPhrase()
	phraseLength := b.conductor.GetPhraseLength()
	halfPoint := phraseLength / 2
	inActiveRange := nextTickInPhrase < halfPoint

	// Determine if we should fire (in burst AND in first 50% of phrase)
	shouldFire := b.inBurst && inActiveRange

	// Update state for next tick
	b.updateBurstState()

	if shouldFire {
		// Fire event with duration = 75% of tick
		noteDuration := time.Duration(float64(tickDuration) * 0.75)
		return []events.ScheduledEvent{{
			Event: events.Event{
				Name: b.name,
				Type: events.EventTypeNote,
				A:    float32(b.note),
				B:    float32(b.velocity),
			},
			Timing: events.Timing{
				Timestamp: nextTickTime,
				Duration:  noteDuration,
			},
		}}
	}

	return nil
}
