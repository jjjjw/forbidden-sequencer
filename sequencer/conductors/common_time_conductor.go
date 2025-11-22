package conductors

import (
	"time"
)

// CommonTimeConductor implements a tick-based conductor with common time (beat-based) concepts
type CommonTimeConductor struct {
	tickDuration time.Duration
	ticksPerBeat int
	bpm          float64
	lastBeatTime time.Time // time of the last beat
	lastTickTime time.Time // time of the last tick
	tickInBeat   int       // current tick within the beat (0 to ticksPerBeat-1)
	Beats        chan struct{}
}

// NewCommonTimeConductor creates a new common time conductor
// ticksPerBeat: number of ticks in one beat (e.g., 4 = 16th notes, 8 = 32nd notes)
// bpm: beats per minute
func NewCommonTimeConductor(bpm float64, ticksPerBeat int) *CommonTimeConductor {
	c := &CommonTimeConductor{
		ticksPerBeat: ticksPerBeat,
		bpm:          bpm,
		tickInBeat:   0,
		Beats:        make(chan struct{}, 100),
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

// GetTickDuration returns the duration of a single tick
func (c *CommonTimeConductor) GetTickDuration() time.Duration {
	return c.tickDuration
}

// GetTicksPerBeat returns the number of ticks per beat
func (c *CommonTimeConductor) GetTicksPerBeat() int {
	return c.ticksPerBeat
}

// GetNextBeatTime returns the absolute wall-clock time of the next beat boundary
func (c *CommonTimeConductor) GetNextBeatTime() time.Time {
	beatDuration := c.tickDuration * time.Duration(c.ticksPerBeat)
	return c.lastBeatTime.Add(beatDuration)
}

// GetNextTickTime returns the absolute wall-clock time of the next tick
func (c *CommonTimeConductor) GetNextTickTime() time.Time {
	return c.lastTickTime.Add(c.tickDuration)
}

// Start begins ticking continuously
func (c *CommonTimeConductor) Start() {
	now := time.Now()
	c.lastBeatTime = now
	c.lastTickTime = now
	c.tickInBeat = 0
	c.scheduleNextTick()
}

// scheduleNextTick schedules the next tick using AfterFunc
func (c *CommonTimeConductor) scheduleNextTick() {
	nextTickTime := c.GetNextTickTime()
	delay := time.Until(nextTickTime)

	time.AfterFunc(delay, func() {
		c.AdvanceTick()
		c.scheduleNextTick()
	})
}

// AdvanceTick increments the tick counter (called by run loop)
func (c *CommonTimeConductor) AdvanceTick() {
	c.lastTickTime = c.GetNextTickTime()
	c.tickInBeat++

	// Check if we've completed a beat
	if c.tickInBeat >= c.ticksPerBeat {
		c.tickInBeat = 0
		c.lastBeatTime = c.lastTickTime

		// Send beat notification
		if c.Beats != nil {
			select {
			case c.Beats <- struct{}{}:
			default:
				// Don't block if channel is full
			}
		}
	}
}

// GetBeatsChannel returns the channel for beat events
func (c *CommonTimeConductor) GetBeatsChannel() chan struct{} {
	return c.Beats
}

// GetLastBeatTime returns the time of the last beat
func (c *CommonTimeConductor) GetLastBeatTime() time.Time {
	return c.lastBeatTime
}

// GetLastTickTime returns the time of the last tick
func (c *CommonTimeConductor) GetLastTickTime() time.Time {
	return c.lastTickTime
}

// GetTickInBeat returns the current tick position within the beat (0 to ticksPerBeat-1)
func (c *CommonTimeConductor) GetTickInBeat() int {
	return c.tickInBeat
}
