package conductors

import (
	"sync"
	"time"
)

// Conductor implements a basic tick clock with adjustable tick duration
type Conductor struct {
	tickDuration time.Duration
	lastTickTime time.Time
	mu           sync.RWMutex // protects tickDuration
}

// NewConductor creates a new conductor
func NewConductor(tickDuration time.Duration) *Conductor {
	return &Conductor{
		tickDuration: tickDuration,
	}
}

// GetTickDuration returns the current tick duration
func (c *Conductor) GetTickDuration() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tickDuration
}

// GetLastTickTime returns the time of the last tick
func (c *Conductor) GetLastTickTime() time.Time {
	return c.lastTickTime
}

// SetTickDuration sets a new tick duration
func (c *Conductor) SetTickDuration(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tickDuration = duration
}

// GetNextTickTime returns the absolute wall-clock time of the next tick
func (c *Conductor) GetNextTickTime() time.Time {
	return c.lastTickTime.Add(c.GetTickDuration())
}

// Start begins ticking continuously
func (c *Conductor) Start() {
	now := time.Now()
	c.lastTickTime = now
	c.scheduleNextTick()
}

// scheduleNextTick schedules the next tick using AfterFunc
func (c *Conductor) scheduleNextTick() {
	nextTickTime := c.GetNextTickTime()
	delay := time.Until(nextTickTime)

	time.AfterFunc(delay, func() {
		c.advanceTick()
		c.scheduleNextTick()
	})
}

// advanceTick increments the tick counter
func (c *Conductor) advanceTick() {
	c.lastTickTime = c.lastTickTime.Add(c.GetTickDuration())
}
