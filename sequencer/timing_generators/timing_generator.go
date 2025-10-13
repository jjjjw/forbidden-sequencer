package timing_generators

import "time"

// TimingGenerator is an interface for generating timing information
type TimingGenerator interface {
	// GetNextTiming returns the wait duration and event duration
	GetNextTiming() (wait time.Duration, duration time.Duration, err error)
}
