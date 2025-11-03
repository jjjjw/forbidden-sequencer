package conductors

import "time"

// Conductor is a minimal tick-based clock interface
// Implementations may add domain-specific time concepts (e.g., beats, measures)
type Conductor interface {
	// GetCurrentTick returns the current tick number
	GetCurrentTick() int64

	// GetTickDuration returns the duration of a single tick
	GetTickDuration() time.Duration

	// Start starts the conductor's clock
	Start()

	// Stop stops the conductor's clock and goroutine
	Stop()

	// Reset resets the conductor to initial state (tick 0, new start time)
	Reset()
}
