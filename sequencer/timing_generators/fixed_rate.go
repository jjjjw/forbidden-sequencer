package timing_generators

import "time"

// FixedRateTiming generates events at a fixed rate
type FixedRateTiming struct {
	// Interval between events
	Interval time.Duration
	// Duration of each event
	Duration time.Duration
}

// NewFixedRateTiming creates a new fixed rate timing generator
func NewFixedRateTiming(interval, duration time.Duration) *FixedRateTiming {
	return &FixedRateTiming{
		Interval: interval,
		Duration: duration,
	}
}

// GetNextTiming implements TimingGenerator.GetNextTiming
func (f *FixedRateTiming) GetNextTiming() (time.Duration, time.Duration, error) {
	return f.Interval, f.Duration, nil
}
