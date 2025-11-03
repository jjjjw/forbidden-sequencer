package conductors

import (
	"sync/atomic"
	"time"
)

// CommonTimeConductor implements a tick-based conductor with common time (beat-based) concepts
type CommonTimeConductor struct {
	currentTick  atomic.Int64
	tickDuration time.Duration
	ticksPerBeat int
	bpm          float64
	running      bool
	stopCh       chan struct{}
	updateCh     chan struct{} // signals run loop to update ticker
	startTime    time.Time     // absolute start time for drift-free scheduling
}

// NewCommonTimeConductor creates a new common time conductor
// ticksPerBeat: number of ticks in one beat (e.g., 4 = 16th notes, 8 = 32nd notes)
// bpm: beats per minute
func NewCommonTimeConductor(ticksPerBeat int, bpm float64) *CommonTimeConductor {
	c := &CommonTimeConductor{
		ticksPerBeat: ticksPerBeat,
		bpm:          bpm,
		running:      false,
	}
	c.currentTick.Store(0)
	c.updateTickDuration()
	return c
}

// updateTickDuration calculates tick duration from BPM and ticks per beat
func (c *CommonTimeConductor) updateTickDuration() {
	// 60 seconds / BPM = seconds per beat
	// seconds per beat / ticks per beat = seconds per tick
	secondsPerBeat := 60.0 / c.bpm
	secondsPerTick := secondsPerBeat / float64(c.ticksPerBeat)
	c.tickDuration = time.Duration(secondsPerTick * float64(time.Second))
}

// GetCurrentTick implements Conductor interface
func (c *CommonTimeConductor) GetCurrentTick() int64 {
	return c.currentTick.Load()
}

// GetTickDuration implements Conductor interface
func (c *CommonTimeConductor) GetTickDuration() time.Duration {
	return c.tickDuration
}

// GetTicksPerBeat returns the number of ticks per beat
func (c *CommonTimeConductor) GetTicksPerBeat() int {
	return c.ticksPerBeat
}

// GetBPM returns the current tempo in beats per minute
func (c *CommonTimeConductor) GetBPM() float64 {
	return c.bpm
}

// SetBPM sets the tempo in beats per minute and updates tick duration
// If the conductor is running, it will signal the run loop to update the ticker
func (c *CommonTimeConductor) SetBPM(bpm float64) {
	c.bpm = bpm
	c.updateTickDuration()

	// Signal run loop to update ticker if running
	if c.running && c.updateCh != nil {
		select {
		case c.updateCh <- struct{}{}:
		default:
			// Don't block if channel is full
		}
	}
}

// IsBeatStart returns true if the current tick is on a beat boundary
func (c *CommonTimeConductor) IsBeatStart() bool {
	return c.GetCurrentTick()%int64(c.ticksPerBeat) == 0
}

// GetTickInBeat returns the current tick position within the beat (0 to ticksPerBeat-1)
func (c *CommonTimeConductor) GetTickInBeat() int {
	return int(c.GetCurrentTick() % int64(c.ticksPerBeat))
}

// GetBeat returns the current beat number (0-indexed)
func (c *CommonTimeConductor) GetBeat() int64 {
	return c.GetCurrentTick() / int64(c.ticksPerBeat)
}

// GetAbsoluteTimeForTick returns the absolute wall-clock time for a given tick
// This enables drift-free scheduling by calculating when a tick should occur
// relative to the conductor's start time
func (c *CommonTimeConductor) GetAbsoluteTimeForTick(tick int64) time.Time {
	return c.startTime.Add(c.tickDuration * time.Duration(tick))
}

// Start starts the conductor's clock in a goroutine
func (c *CommonTimeConductor) Start() {
	if c.running {
		return
	}

	c.running = true
	c.stopCh = make(chan struct{})
	c.updateCh = make(chan struct{}, 1)
	c.startTime = time.Now() // capture absolute start time

	go c.run()
}

// Stop stops the conductor's clock and goroutine
func (c *CommonTimeConductor) Stop() {
	if !c.running {
		return
	}

	c.running = false
	close(c.stopCh)
}

// run is the internal tick loop
func (c *CommonTimeConductor) run() {
	ticker := time.NewTicker(c.tickDuration)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-c.updateCh:
			// BPM changed, recreate ticker with new duration
			ticker.Stop()
			ticker = time.NewTicker(c.tickDuration)
		case <-ticker.C:
			c.AdvanceTick()
		}
	}
}

// AdvanceTick increments the tick counter (called by run loop)
func (c *CommonTimeConductor) AdvanceTick() {
	c.currentTick.Add(1)
}

// Reset resets the tick counter to 0 and updates start time to now
func (c *CommonTimeConductor) Reset() {
	c.currentTick.Store(0)
	c.startTime = time.Now()
}
