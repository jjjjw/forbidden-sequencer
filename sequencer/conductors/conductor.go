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

	// Pause pauses the conductor's clock (tick advancement stops)
	Pause()

	// Resume resumes the conductor's clock after pause
	Resume()
}
