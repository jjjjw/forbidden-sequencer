package conductors

import (
	"sync"
	"time"
)

// TickCallback is called on every tick with the current tick number
type TickCallback func(tick int64)

// Conductor implements a basic tick clock with adjustable tick duration
type Conductor struct {
	tickDuration time.Duration
	lastTickTime time.Time
	currentTick  int64
	callback     TickCallback
	mu           sync.RWMutex // protects tickDuration and callback
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

// GetCurrentTick returns the current tick number
func (c *Conductor) GetCurrentTick() int64 {
	return c.currentTick
}

// SetTickCallback sets the callback to be invoked on every tick
func (c *Conductor) SetTickCallback(callback TickCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.callback = callback
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
	c.currentTick = 0
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

// advanceTick increments the tick counter and invokes the callback
func (c *Conductor) advanceTick() {
	c.lastTickTime = c.lastTickTime.Add(c.GetTickDuration())
	c.currentTick++

	// Invoke callback if set (read with lock)
	c.mu.RLock()
	callback := c.callback
	c.mu.RUnlock()

	if callback != nil {
		callback(c.currentTick)
	}
}

// Reset resets the conductor to tick 0
func (c *Conductor) Reset() {
	c.currentTick = 0
	c.lastTickTime = time.Now()
}
