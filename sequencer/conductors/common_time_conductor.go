package conductors

import (
	"time"
)

// CommonTimeConductor implements a tick-based conductor with common time (beat-based) concepts
type CommonTimeConductor struct {
	currentTick  int64
	tickDuration time.Duration
	ticksPerBeat int
	bpm          float64
	startTime    time.Time // absolute start time for drift-free scheduling
	Beats        chan int64
}

// NewCommonTimeConductor creates a new common time conductor
// ticksPerBeat: number of ticks in one beat (e.g., 4 = 16th notes, 8 = 32nd notes)
// bpm: beats per minute
func NewCommonTimeConductor(bpm float64, ticksPerBeat int) *CommonTimeConductor {
	c := &CommonTimeConductor{
		ticksPerBeat: ticksPerBeat,
		bpm:          bpm,
		currentTick:  0,
		Beats:        make(chan int64, 100),
	}
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
	return c.currentTick
}

// GetTickDuration implements Conductor interface
func (c *CommonTimeConductor) GetTickDuration() time.Duration {
	return c.tickDuration
}

// GetTicksPerBeat returns the number of ticks per beat
func (c *CommonTimeConductor) GetTicksPerBeat() int {
	return c.ticksPerBeat
}

// GetNextBeatTick returns the tick number of the next beat boundary
func (c *CommonTimeConductor) GetNextBeatTick() int64 {
	currentTick := c.GetCurrentTick()
	ticksPerBeat := int64(c.ticksPerBeat)
	// Calculate ticks until next beat using modulo
	ticksUntilBeat := ticksPerBeat - (currentTick % ticksPerBeat)
	return currentTick + ticksUntilBeat
}

// GetAbsoluteTimeForTick returns the absolute wall-clock time for a given tick
// This enables drift-free scheduling by calculating when a tick should occur
// relative to the conductor's start time
func (c *CommonTimeConductor) GetAbsoluteTimeForTick(tick int64) time.Time {
	return c.startTime.Add(c.tickDuration * time.Duration(tick))
}

// Start begins ticking continuously
func (c *CommonTimeConductor) Start() {
	c.startTime = time.Now()
	c.scheduleNextTick()
}

// scheduleNextTick schedules the next tick using AfterFunc
func (c *CommonTimeConductor) scheduleNextTick() {
	nextTick := c.GetCurrentTick() + 1
	nextTickTime := c.GetAbsoluteTimeForTick(nextTick)
	delay := time.Until(nextTickTime)

	time.AfterFunc(delay, func() {
		c.AdvanceTick()
		c.scheduleNextTick()
	})
}

// AdvanceTick increments the tick counter (called by run loop)
func (c *CommonTimeConductor) AdvanceTick() {
	c.currentTick++
	// Send beat number on beat boundaries
	if c.Beats != nil && c.currentTick%int64(c.ticksPerBeat) == 0 {
		beat := c.currentTick / int64(c.ticksPerBeat)
		select {
		case c.Beats <- beat:
		default:
			// Don't block if channel is full
		}
	}
}

// GetBeatsChannel returns the channel for beat events
func (c *CommonTimeConductor) GetBeatsChannel() chan int64 {
	return c.Beats
}
