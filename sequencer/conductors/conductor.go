package conductors

import "time"

// TickCallback is called on each tick
type TickCallback func(tick int64)

// Conductor is a minimal tick-based clock interface
// Implementations may add domain-specific time concepts (e.g., beats, measures)
type Conductor interface {
	// GetCurrentTick returns the current tick number
	GetCurrentTick() int64

	// GetTickDuration returns the duration of a single tick
	GetTickDuration() time.Duration

	// Start begins ticking (runs continuously once started)
	Start()

	// GetBeatsChannel returns the channel for beat events
	GetBeatsChannel() chan int64
}
