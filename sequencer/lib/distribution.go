package lib

import "math"

// Distribution determines which ticks in a phrase should fire events
type Distribution interface {
	// ShouldFire returns true if an event should fire at this tick in the phrase
	ShouldFire(tickInPhrase int, phraseLength int) bool
}

// EvenDistribution fires every N ticks, optionally offset from the start
// Example: Interval=4, Offset=0 → x---x---x---x--- (on beat)
// Example: Interval=4, Offset=2 → --x---x---x---x- (boom tick, offbeat)
type EvenDistribution struct {
	Interval int // ticks between events
	Offset   int // starting tick position (0 to Interval-1)
}

// NewEvenDistribution creates a distribution that fires every N ticks, starting at offset
func NewEvenDistribution(interval int, offset int) *EvenDistribution {
	if interval < 1 {
		interval = 1
	}
	// Ensure offset is within valid range
	offset = offset % interval
	if offset < 0 {
		offset = 0
	}
	return &EvenDistribution{
		Interval: interval,
		Offset:   offset,
	}
}

// ShouldFire returns true if this tick should fire
func (d *EvenDistribution) ShouldFire(tickInPhrase int, phraseLength int) bool {
	// Fire if tick is at offset position within the interval cycle
	return tickInPhrase >= d.Offset && (tickInPhrase-d.Offset)%d.Interval == 0
}

// EuclideanDistribution distributes K events across N ticks using Euclidean algorithm
// Creates maximally even distribution (Bjorklund algorithm)
// Example: 5 events in 12 ticks → x--x-x--x-x- (or similar maximally even pattern)
type EuclideanDistribution struct {
	Events  int   // number of events to distribute
	pattern []bool // precomputed pattern
}

// NewEuclideanDistribution creates a distribution using Euclidean rhythm algorithm
func NewEuclideanDistribution(events int, phraseLength int) *EuclideanDistribution {
	if events < 0 {
		events = 0
	}
	if events > phraseLength {
		events = phraseLength
	}

	pattern := bjorklund(events, phraseLength)
	return &EuclideanDistribution{
		Events:  events,
		pattern: pattern,
	}
}

// ShouldFire returns true if this tick should fire based on Euclidean pattern
func (d *EuclideanDistribution) ShouldFire(tickInPhrase int, phraseLength int) bool {
	if tickInPhrase >= len(d.pattern) {
		return false
	}
	return d.pattern[tickInPhrase]
}

// bjorklund implements the Bjorklund algorithm for Euclidean rhythms
// Distributes k pulses over n steps as evenly as possible
func bjorklund(pulses int, steps int) []bool {
	if pulses == 0 || steps == 0 {
		result := make([]bool, steps)
		return result
	}

	if pulses >= steps {
		result := make([]bool, steps)
		for i := range result {
			result[i] = true
		}
		return result
	}

	// Build pattern using Bjorklund algorithm
	pattern := make([][]bool, steps)

	// Initialize with pulses as [1] and rests as [0]
	for i := 0; i < pulses; i++ {
		pattern[i] = []bool{true}
	}
	for i := pulses; i < steps; i++ {
		pattern[i] = []bool{false}
	}

	// Recursively distribute
	return bjorklundRecurse(pattern, pulses, steps-pulses)
}

func bjorklundRecurse(pattern [][]bool, ones int, zeros int) []bool {
	if zeros <= 1 {
		// Flatten pattern
		result := []bool{}
		for _, p := range pattern {
			result = append(result, p...)
		}
		return result
	}

	if zeros > ones {
		// Append zeros to ones
		for i := 0; i < ones; i++ {
			pattern[i] = append(pattern[i], pattern[ones+i]...)
		}
		return bjorklundRecurse(pattern[:ones+zeros-ones], ones, zeros-ones)
	} else {
		// Append ones to zeros
		for i := 0; i < zeros; i++ {
			pattern[ones+i] = append(pattern[ones+i], pattern[i]...)
		}
		return bjorklundRecurse(pattern[zeros:], zeros, ones-zeros)
	}
}

// AccelerandoDistribution creates events that get closer together over the phrase
// Uses exponential spacing: x---x--x-xx
type AccelerandoDistribution struct {
	Events              int       // number of events in phrase
	Curve               float64   // curve factor (1.0 = linear, >1 = exponential acceleration)
	ticks               []int     // precomputed tick positions (quantized)
	continuousPositions []float64 // continuous positions for non-quantized mode
}

// NewAccelerandoDistribution creates a distribution with accelerating spacing
func NewAccelerandoDistribution(events int, phraseLength int, curve float64) *AccelerandoDistribution {
	if events <= 0 {
		return &AccelerandoDistribution{
			Events:              0,
			Curve:               curve,
			ticks:               []int{},
			continuousPositions: []float64{},
		}
	}
	if curve <= 0 {
		curve = 1.0
	}

	// Calculate both continuous and quantized positions
	tickSet := make(map[int]bool)
	ticks := make([]int, 0, events)
	continuousPositions := make([]float64, 0, events)

	for i := 0; i < events; i++ {
		// Normalize position 0.0 to 1.0
		t := float64(i) / float64(events-1)
		if events == 1 {
			t = 0
		}

		// Apply curve (lower values get more space early, less later)
		// We want more space early, less later, so invert: 1 - t^curve
		curved := 1.0 - math.Pow(1.0-t, curve)

		// Store continuous position (in ticks as float)
		continuousPos := curved * float64(phraseLength-1)
		continuousPositions = append(continuousPositions, continuousPos)

		// Map to quantized tick position
		tick := int(continuousPos)

		// Only add if not duplicate (for quantized mode)
		if !tickSet[tick] {
			tickSet[tick] = true
			ticks = append(ticks, tick)
		}
	}

	return &AccelerandoDistribution{
		Events:              len(ticks), // actual number of unique events (quantized)
		Curve:               curve,
		ticks:               ticks,
		continuousPositions: continuousPositions,
	}
}

// ShouldFire returns true if this tick should fire
func (d *AccelerandoDistribution) ShouldFire(tickInPhrase int, phraseLength int) bool {
	for _, tick := range d.ticks {
		if tick == tickInPhrase {
			return true
		}
	}
	return false
}

// GetActualEvents returns the actual number of unique events after quantization
func (d *AccelerandoDistribution) GetActualEvents() int {
	return len(d.ticks)
}

// GetContinuousPositions returns the continuous positions for non-quantized mode
func (d *AccelerandoDistribution) GetContinuousPositions() []float64 {
	return d.continuousPositions
}

// RitardandoDistribution creates events that get further apart over the phrase
// Uses exponential spacing: xx-x--x---x
type RitardandoDistribution struct {
	Events            int       // number of events in phrase
	Curve             float64   // curve factor (1.0 = linear, >1 = exponential deceleration)
	ticks             []int     // precomputed tick positions (quantized)
	continuousPositions []float64 // continuous positions for non-quantized mode
}

// NewRitardandoDistribution creates a distribution with decelerating spacing
func NewRitardandoDistribution(events int, phraseLength int, curve float64) *RitardandoDistribution {
	if events <= 0 {
		return &RitardandoDistribution{
			Events:              0,
			Curve:               curve,
			ticks:               []int{},
			continuousPositions: []float64{},
		}
	}
	if curve <= 0 {
		curve = 1.0
	}

	// Calculate both continuous and quantized positions
	tickSet := make(map[int]bool)
	ticks := make([]int, 0, events)
	continuousPositions := make([]float64, 0, events)

	for i := 0; i < events; i++ {
		// Normalize position 0.0 to 1.0
		t := float64(i) / float64(events-1)
		if events == 1 {
			t = 0
		}

		// Apply curve (events get further apart)
		curved := math.Pow(t, curve)

		// Store continuous position (in ticks as float)
		continuousPos := curved * float64(phraseLength-1)
		continuousPositions = append(continuousPositions, continuousPos)

		// Map to quantized tick position
		tick := int(continuousPos)

		// Only add if not duplicate (for quantized mode)
		if !tickSet[tick] {
			tickSet[tick] = true
			ticks = append(ticks, tick)
		}
	}

	return &RitardandoDistribution{
		Events:              len(ticks), // actual number of unique events (quantized)
		Curve:               curve,
		ticks:               ticks,
		continuousPositions: continuousPositions,
	}
}

// ShouldFire returns true if this tick should fire
func (d *RitardandoDistribution) ShouldFire(tickInPhrase int, phraseLength int) bool {
	for _, tick := range d.ticks {
		if tick == tickInPhrase {
			return true
		}
	}
	return false
}

// GetActualEvents returns the actual number of unique events after quantization
func (d *RitardandoDistribution) GetActualEvents() int {
	return len(d.ticks)
}

// GetContinuousPositions returns the continuous positions for non-quantized mode
func (d *RitardandoDistribution) GetContinuousPositions() []float64 {
	return d.continuousPositions
}
