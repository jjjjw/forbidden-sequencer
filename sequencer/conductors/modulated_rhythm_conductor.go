package conductors

import (
	"math/rand"
)

// ModulatedRhythmConductor extends PhraseConductor with rhythm pattern decisions
// Each phrase it makes random decisions about which instruments will play
type ModulatedRhythmConductor struct {
	*PhraseConductor
	snareWillTrigger bool // whether snare will trigger this phrase (33% chance)
	hihatClosed      bool // whether hihat is closed this phrase (75% closed, 25% open)
	lastSeenTick     int  // last tick we saw, to detect phrase boundaries
}

// NewModulatedRhythmConductor creates a conductor with rhythm decision-making
func NewModulatedRhythmConductor(phraseConductor *PhraseConductor) *ModulatedRhythmConductor {
	rdc := &ModulatedRhythmConductor{
		PhraseConductor: phraseConductor,
		lastSeenTick:    -1, // -1 means we haven't seen any tick yet
	}

	// Make initial decisions
	rdc.makeRhythmDecisions()

	return rdc
}

// checkAndUpdatePhrase checks if we've wrapped to a new phrase and updates decisions if so
func (r *ModulatedRhythmConductor) checkAndUpdatePhrase() {
	currentTick := r.GetNextTickInPhrase() // Get the NEXT tick (what patterns will use)

	// Check if we've wrapped around (phrase boundary)
	// This happens when currentTick < lastSeenTick or when we're seeing tick for first time
	if r.lastSeenTick != -1 && currentTick < r.lastSeenTick {
		r.makeRhythmDecisions()
	}

	r.lastSeenTick = currentTick
}

// makeRhythmDecisions generates random decisions for the current phrase
func (r *ModulatedRhythmConductor) makeRhythmDecisions() {
	// 33% chance for snare
	r.snareWillTrigger = rand.Float64() < 0.33

	// 75% chance for closed hihat, 25% for open
	r.hihatClosed = rand.Float64() < 0.75
}

// WillSnareTrigger returns whether the snare will trigger this phrase
func (r *ModulatedRhythmConductor) WillSnareTrigger() bool {
	r.checkAndUpdatePhrase()
	return r.snareWillTrigger
}

// IsHihatClosed returns whether the hihat is closed (true) or open (false) this phrase
func (r *ModulatedRhythmConductor) IsHihatClosed() bool {
	r.checkAndUpdatePhrase()
	return r.hihatClosed
}

// GetSnareTriggerTick returns the tick position where snare triggers (3/4 of phrase)
func (r *ModulatedRhythmConductor) GetSnareTriggerTick() int {
	return r.GetPhraseLength() * 3 / 4
}
