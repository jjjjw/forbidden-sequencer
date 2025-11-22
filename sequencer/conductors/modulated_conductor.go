package conductors

import (
	"time"
)

// ModulatedConductor implements a conductor with variable tick rate and phrase tracking
type ModulatedConductor struct {
	baseTickDuration time.Duration
	rateMultiplier   float64 // 1.0 = normal, 2.0 = double speed, 0.5 = half speed
	phraseLength     int     // length of phrase in ticks
	tickInPhrase     int     // current tick within the phrase (0 to phraseLength-1)
	lastTickTime     time.Time
	RateChanges      chan float64 // channel for rate multiplier updates
}

// NewModulatedConductor creates a new modulated conductor
// baseTickDuration: base duration between ticks (before rate modulation)
// phraseLength: number of ticks in one phrase
func NewModulatedConductor(baseTickDuration time.Duration, phraseLength int) *ModulatedConductor {
	return &ModulatedConductor{
		baseTickDuration: baseTickDuration,
		rateMultiplier:   1.0,
		phraseLength:     phraseLength,
		tickInPhrase:     0,
		RateChanges:      make(chan float64, 10),
	}
}

// GetTickDuration returns the current tick duration (base * 1/rate)
func (c *ModulatedConductor) GetTickDuration() time.Duration {
	return time.Duration(float64(c.baseTickDuration) / c.rateMultiplier)
}

// GetBaseTickDuration returns the base tick duration (unmodulated)
func (c *ModulatedConductor) GetBaseTickDuration() time.Duration {
	return c.baseTickDuration
}

// GetRateMultiplier returns the current rate multiplier
func (c *ModulatedConductor) GetRateMultiplier() float64 {
	return c.rateMultiplier
}

// GetPhraseLength returns the phrase length in ticks
func (c *ModulatedConductor) GetPhraseLength() int {
	return c.phraseLength
}

// GetTickInPhrase returns the current tick position within the phrase
func (c *ModulatedConductor) GetTickInPhrase() int {
	return c.tickInPhrase
}

// GetNextTickInPhrase returns the next tick position within the phrase
func (c *ModulatedConductor) GetNextTickInPhrase() int {
	next := c.tickInPhrase + 1
	if next >= c.phraseLength {
		return 0
	}
	return next
}

// GetNextTickTime returns the absolute wall-clock time of the next tick
func (c *ModulatedConductor) GetNextTickTime() time.Time {
	return c.lastTickTime.Add(c.GetTickDuration())
}

// GetLastTickTime returns the time of the last tick
func (c *ModulatedConductor) GetLastTickTime() time.Time {
	return c.lastTickTime
}

// Start begins ticking continuously
func (c *ModulatedConductor) Start() {
	now := time.Now()
	c.lastTickTime = now
	c.tickInPhrase = 0
	c.scheduleNextTick()
}

// scheduleNextTick schedules the next tick using AfterFunc
func (c *ModulatedConductor) scheduleNextTick() {
	nextTickTime := c.GetNextTickTime()
	delay := time.Until(nextTickTime)

	time.AfterFunc(delay, func() {
		// Process any pending rate changes
		c.processRateChanges()

		c.advanceTick()
		c.scheduleNextTick()
	})
}

// processRateChanges drains the rate changes channel and applies the latest value
func (c *ModulatedConductor) processRateChanges() {
	for {
		select {
		case newRate := <-c.RateChanges:
			if newRate > 0 {
				c.rateMultiplier = newRate
			}
		default:
			return
		}
	}
}

// advanceTick increments the tick counter
func (c *ModulatedConductor) advanceTick() {
	c.lastTickTime = c.lastTickTime.Add(c.GetTickDuration())
	c.tickInPhrase++

	// Wrap around at phrase boundary
	if c.tickInPhrase >= c.phraseLength {
		c.tickInPhrase = 0
	}
}

// GetBeatsChannel returns nil - this conductor doesn't use beat events
func (c *ModulatedConductor) GetBeatsChannel() chan struct{} {
	return nil
}
