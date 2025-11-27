package conductors

import "time"

// TickCallback is called on each tick
type TickCallback func(tick int64)

// Conductor is a minimal tick-based clock interface
// Implementations may add domain-specific time concepts (e.g., beats, measures)
type Conductor interface {
	// GetTickDuration returns the duration of a single tick
	GetTickDuration() time.Duration

	// Start begins ticking (runs continuously once started)
	Start()

	// Ticks returns a channel that emits on each tick
	// Implementation note: Each call to Ticks() should return a fresh channel
	// to support multiple subscribers. Store channels in a slice and notify all on each tick.
	Ticks() <-chan struct{}
}
