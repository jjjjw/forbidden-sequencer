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
}

// NewModulatedRhythmConductor creates a conductor with rhythm decision-making
func NewModulatedRhythmConductor(phraseConductor *PhraseConductor) *ModulatedRhythmConductor {
	rdc := &ModulatedRhythmConductor{
		PhraseConductor: phraseConductor,
	}

	// Make initial decisions
	rdc.makeRhythmDecisions()

	// Subscribe to ticks to detect phrase boundaries
	tickChan := phraseConductor.Ticks()
	go rdc.watchPhraseChanges(tickChan)

	return rdc
}

// watchPhraseChanges listens for tick events and makes new decisions at phrase boundaries
func (r *ModulatedRhythmConductor) watchPhraseChanges(tickChan <-chan struct{}) {
	for range tickChan {
		// Check if we're at the start of a new phrase (tick 0)
		if r.GetCurrentTickInPhrase() == 0 {
			r.makeRhythmDecisions()
		}
	}
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
	return r.snareWillTrigger
}

// IsHihatClosed returns whether the hihat is closed (true) or open (false) this phrase
func (r *ModulatedRhythmConductor) IsHihatClosed() bool {
	return r.hihatClosed
}

// GetSnareTriggerTick returns the tick position where snare triggers (3/4 of phrase)
func (r *ModulatedRhythmConductor) GetSnareTriggerTick() int {
	return r.GetPhraseLength() * 3 / 4
}
