package conductors

import (
	"time"
)

// PhraseConductor implements a conductor with variable tick rate and phrase tracking
type PhraseConductor struct {
	baseTickDuration time.Duration
	rateMultiplier   float64 // 1.0 = normal, 2.0 = double speed, 0.5 = half speed
	phraseLength     int     // length of phrase in ticks
	tickInPhrase     int     // current tick within the phrase (0 to phraseLength-1)
	lastTickTime     time.Time
	rateChanges      chan float64  // channel for rate multiplier updates
	ticks            chan struct{} // channel for tick events
}

// NewPhraseConductor creates a new modulated conductor
// baseTickDuration: base duration between ticks (before rate modulation)
// phraseLength: number of ticks in one phrase
func NewPhraseConductor(baseTickDuration time.Duration, phraseLength int) *PhraseConductor {
	return &PhraseConductor{
		baseTickDuration: baseTickDuration,
		rateMultiplier:   1.0,
		phraseLength:     phraseLength,
		tickInPhrase:     0,
		rateChanges:      make(chan float64, 10),
		ticks:            make(chan struct{}, 100),
	}
}

// RateChanges returns the channel for sending rate multiplier updates
func (c *PhraseConductor) RateChanges() chan<- float64 {
	return c.rateChanges
}

// GetTickDuration returns the current tick duration (base * 1/rate)
func (c *PhraseConductor) GetTickDuration() time.Duration {
	return time.Duration(float64(c.baseTickDuration) / c.rateMultiplier)
}

// GetBaseTickDuration returns the base tick duration (unmodulated)
func (c *PhraseConductor) GetBaseTickDuration() time.Duration {
	return c.baseTickDuration
}

// GetRateMultiplier returns the current rate multiplier
func (c *PhraseConductor) GetRateMultiplier() float64 {
	return c.rateMultiplier
}

// GetPhraseLength returns the phrase length in ticks
func (c *PhraseConductor) GetPhraseLength() int {
	return c.phraseLength
}

// GetNextTickInPhrase returns the next tick position within the phrase
func (c *PhraseConductor) GetNextTickInPhrase() int {
	next := c.tickInPhrase + 1
	if next >= c.phraseLength {
		return 0
	}
	return next
}

// GetNextTickTime returns the absolute wall-clock time of the next tick
func (c *PhraseConductor) GetNextTickTime() time.Time {
	return c.lastTickTime.Add(c.GetTickDuration())
}

// GetLastTickTime returns the time of the last tick
func (c *PhraseConductor) GetLastTickTime() time.Time {
	return c.lastTickTime
}

// Start begins ticking continuously
func (c *PhraseConductor) Start() {
	now := time.Now()
	c.lastTickTime = now
	c.tickInPhrase = 0
	c.scheduleNextTick()
}

// scheduleNextTick schedules the next tick using AfterFunc
func (c *PhraseConductor) scheduleNextTick() {
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
func (c *PhraseConductor) processRateChanges() {
	for {
		select {
		case newRate := <-c.rateChanges:
			if newRate > 0 {
				c.rateMultiplier = newRate
			}
		default:
			return
		}
	}
}

// advanceTick increments the tick counter
func (c *PhraseConductor) advanceTick() {
	c.lastTickTime = c.lastTickTime.Add(c.GetTickDuration())
	c.tickInPhrase++

	// Wrap around at phrase boundary
	if c.tickInPhrase >= c.phraseLength {
		c.tickInPhrase = 0
	}

	// Send tick notification
	if c.ticks != nil {
		select {
		case c.ticks <- struct{}{}:
		default:
			// Don't block if channel is full
		}
	}
}

// Ticks returns the channel for tick events
func (c *PhraseConductor) Ticks() <-chan struct{} {
	return c.ticks
}
