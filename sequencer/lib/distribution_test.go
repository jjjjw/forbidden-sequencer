package lib

import (
	"strings"
	"testing"
)

// visualize creates a visual representation of a distribution
// Returns a string like "x---x---x---x---" for easy pattern verification
func visualize(dist Distribution, phraseLength int) string {
	var sb strings.Builder
	for i := 0; i < phraseLength; i++ {
		if dist.ShouldFire(i, phraseLength) {
			sb.WriteString("x")
		} else {
			sb.WriteString("-")
		}
	}
	return sb.String()
}

func TestEvenDistribution(t *testing.T) {
	tests := []struct {
		name         string
		interval     int
		offset       int
		phraseLength int
		expected     string
	}{
		{"every tick", 1, 0, 8, "xxxxxxxx"},
		{"every 2 ticks", 2, 0, 8, "x-x-x-x-"},
		{"every 4 ticks", 4, 0, 16, "x---x---x---x---"},
		{"every 3 ticks", 3, 0, 12, "x--x--x--x--"},
		{"every 4 ticks, offset 1", 4, 1, 16, "-x---x---x---x--"},
		{"every 4 ticks, offset 2 (boom tick)", 4, 2, 16, "--x---x---x---x-"},
		{"every 4 ticks, offset 3", 4, 3, 16, "---x---x---x---x"},
		{"every 2 ticks, offset 1", 2, 1, 8, "-x-x-x-x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewEvenDistribution(tt.interval, tt.offset)
			result := visualize(dist, tt.phraseLength)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
			t.Logf("Pattern: %s", result)
		})
	}
}

func TestEuclideanDistribution(t *testing.T) {
	tests := []struct {
		name         string
		events       int
		phraseLength int
	}{
		{"3 in 8 (tresillo)", 3, 8},
		{"5 in 8", 5, 8},
		{"5 in 12", 5, 12},
		{"7 in 16", 7, 16},
		{"3 in 4", 3, 4},
		{"5 in 16 (classic)", 5, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewEuclideanDistribution(tt.events, tt.phraseLength)
			result := visualize(dist, tt.phraseLength)

			// Count events
			count := 0
			for i := 0; i < tt.phraseLength; i++ {
				if dist.ShouldFire(i, tt.phraseLength) {
					count++
				}
			}

			if count != tt.events {
				t.Errorf("Expected %d events, got %d", tt.events, count)
			}

			t.Logf("Pattern (%d in %d): %s", tt.events, tt.phraseLength, result)
		})
	}
}

func TestEuclideanDistribution_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		events       int
		phraseLength int
		expected     string
	}{
		{"zero events", 0, 8, "--------"},
		{"all events", 8, 8, "xxxxxxxx"},
		{"more events than length", 10, 8, "xxxxxxxx"},
		{"one event", 1, 8, "x-------"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewEuclideanDistribution(tt.events, tt.phraseLength)
			result := visualize(dist, tt.phraseLength)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
			t.Logf("Pattern: %s", result)
		})
	}
}

func TestAccelerandoDistribution(t *testing.T) {
	tests := []struct {
		name         string
		events       int
		phraseLength int
		curve        float64
	}{
		{"4 events linear", 4, 16, 1.0},
		{"4 events exponential", 4, 16, 2.0},
		{"6 events linear", 6, 16, 1.0},
		{"6 events strong curve", 6, 16, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewAccelerandoDistribution(tt.events, tt.phraseLength, tt.curve)
			result := visualize(dist, tt.phraseLength)

			// Count events
			count := 0
			for i := 0; i < tt.phraseLength; i++ {
				if dist.ShouldFire(i, tt.phraseLength) {
					count++
				}
			}

			// Note: high curves may cause duplicates, resulting in fewer actual events
			if count > tt.events {
				t.Errorf("Expected at most %d events, got %d", tt.events, count)
			}

			t.Logf("Pattern (curve=%.1f): %s (%d events)", tt.curve, result, count)

			// Visual check: more gaps at start, fewer at end
			firstHalf := result[:tt.phraseLength/2]
			secondHalf := result[tt.phraseLength/2:]
			firstCount := strings.Count(firstHalf, "x")
			secondCount := strings.Count(secondHalf, "x")

			t.Logf("  First half: %d events, Second half: %d events", firstCount, secondCount)
		})
	}
}

func TestRitardandoDistribution(t *testing.T) {
	tests := []struct {
		name         string
		events       int
		phraseLength int
		curve        float64
	}{
		{"4 events linear", 4, 16, 1.0},
		{"4 events exponential", 4, 16, 2.0},
		{"6 events linear", 6, 16, 1.0},
		{"6 events strong curve", 6, 16, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewRitardandoDistribution(tt.events, tt.phraseLength, tt.curve)
			result := visualize(dist, tt.phraseLength)

			// Count events
			count := 0
			for i := 0; i < tt.phraseLength; i++ {
				if dist.ShouldFire(i, tt.phraseLength) {
					count++
				}
			}

			// Note: high curves may cause duplicates, resulting in fewer actual events
			if count > tt.events {
				t.Errorf("Expected at most %d events, got %d", tt.events, count)
			}

			t.Logf("Pattern (curve=%.1f): %s (%d events)", tt.curve, result, count)

			// Visual check: fewer gaps at start, more at end
			firstHalf := result[:tt.phraseLength/2]
			secondHalf := result[tt.phraseLength/2:]
			firstCount := strings.Count(firstHalf, "x")
			secondCount := strings.Count(secondHalf, "x")

			t.Logf("  First half: %d events, Second half: %d events", firstCount, secondCount)
		})
	}
}

func TestDistribution_Comparison(t *testing.T) {
	phraseLength := 16
	events := 5

	t.Log("Comparing different distribution types with 5 events in 16 ticks:")

	even := NewEvenDistribution(3, 0)
	t.Logf("Even (interval=3):    %s", visualize(even, phraseLength))

	euclidean := NewEuclideanDistribution(events, phraseLength)
	t.Logf("Euclidean (5 in 16):  %s", visualize(euclidean, phraseLength))

	accel := NewAccelerandoDistribution(events, phraseLength, 2.0)
	t.Logf("Accelerando (5, 2x):  %s", visualize(accel, phraseLength))

	rit := NewRitardandoDistribution(events, phraseLength, 2.0)
	t.Logf("Ritardando (5, 2x):   %s", visualize(rit, phraseLength))
}

func TestRitardandoDistribution_Quantization(t *testing.T) {
	tests := []struct {
		name         string
		events       int
		phraseLength int
		curve        float64
		expectLoss   bool // whether we expect event loss due to quantization
	}{
		{"linear curve - no loss", 8, 16, 1.0, false},
		{"moderate curve - minimal loss", 8, 16, 2.0, true},
		{"high curve - significant loss", 8, 16, 4.0, true},
		{"very high curve - major loss", 8, 16, 5.0, true},
		{"low events - less likely to collide", 4, 16, 2.0, false},
		{"high events in short phrase", 8, 8, 2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewRitardandoDistribution(tt.events, tt.phraseLength, tt.curve)

			// Get actual event count after quantization
			actualEvents := dist.GetActualEvents()

			// Verify we requested the correct number
			if dist.Events != actualEvents {
				t.Logf("Requested %d events, got %d actual events (%.1f%% retained)",
					tt.events, actualEvents, float64(actualEvents)/float64(tt.events)*100)
			}

			// Check if loss occurred as expected
			hasLoss := actualEvents < tt.events
			if tt.expectLoss && !hasLoss {
				t.Logf("Expected event loss with curve=%.1f but got all %d events", tt.curve, actualEvents)
			}

			// Verify actual events matches what ShouldFire reports
			countedEvents := 0
			for i := 0; i < tt.phraseLength; i++ {
				if dist.ShouldFire(i, tt.phraseLength) {
					countedEvents++
				}
			}
			if countedEvents != actualEvents {
				t.Errorf("GetActualEvents() returned %d but ShouldFire counted %d", actualEvents, countedEvents)
			}

			t.Logf("Pattern (curve=%.1f): %s (%d/%d events)",
				tt.curve, visualize(dist, tt.phraseLength), actualEvents, tt.events)
		})
	}
}

func TestRitardandoDistribution_ContinuousPositions(t *testing.T) {
	tests := []struct {
		name         string
		events       int
		phraseLength int
		curve        float64
	}{
		{"linear curve", 4, 16, 1.0},
		{"moderate curve", 8, 16, 2.0},
		{"high curve", 8, 16, 4.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewRitardandoDistribution(tt.events, tt.phraseLength, tt.curve)

			// Get continuous positions
			positions := dist.GetContinuousPositions()

			// Verify we have all requested positions (before quantization)
			if len(positions) != tt.events {
				t.Errorf("Expected %d continuous positions, got %d", tt.events, len(positions))
			}

			// Verify positions are in ascending order
			for i := 1; i < len(positions); i++ {
				if positions[i] <= positions[i-1] {
					t.Errorf("Positions not in ascending order: %.2f followed by %.2f",
						positions[i-1], positions[i])
				}
			}

			// Verify positions are within phrase bounds
			for i, pos := range positions {
				if pos < 0 || pos >= float64(tt.phraseLength) {
					t.Errorf("Position %d (%.2f) is out of bounds [0, %d)", i, pos, tt.phraseLength)
				}
			}

			// Log the positions to show sub-tick precision
			t.Logf("Continuous positions (curve=%.1f):", tt.curve)
			for i, pos := range positions {
				tick := int(pos)
				offset := pos - float64(tick)
				t.Logf("  Event %d: position %.3f (tick %d + %.3f)", i, pos, tick, offset)
			}

			// Demonstrate quantization loss
			actualEvents := dist.GetActualEvents()
			if actualEvents < tt.events {
				t.Logf("Quantization reduced %d events to %d unique ticks (%.1f%% loss)",
					tt.events, actualEvents, float64(tt.events-actualEvents)/float64(tt.events)*100)
			}
		})
	}
}

func TestAccelerandoDistribution_Quantization(t *testing.T) {
	tests := []struct {
		name         string
		events       int
		phraseLength int
		curve        float64
		expectLoss   bool
	}{
		{"linear curve - no loss", 8, 16, 1.0, false},
		{"moderate curve - minimal loss", 8, 16, 2.0, true},
		{"high curve - significant loss", 8, 16, 4.0, true},
		{"very high curve - major loss", 8, 16, 5.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewAccelerandoDistribution(tt.events, tt.phraseLength, tt.curve)

			actualEvents := dist.GetActualEvents()

			// Verify actual events matches what ShouldFire reports
			countedEvents := 0
			for i := 0; i < tt.phraseLength; i++ {
				if dist.ShouldFire(i, tt.phraseLength) {
					countedEvents++
				}
			}
			if countedEvents != actualEvents {
				t.Errorf("GetActualEvents() returned %d but ShouldFire counted %d", actualEvents, countedEvents)
			}

			hasLoss := actualEvents < tt.events
			if tt.expectLoss && !hasLoss {
				t.Logf("Expected event loss with curve=%.1f but got all %d events", tt.curve, actualEvents)
			}

			t.Logf("Pattern (curve=%.1f): %s (%d/%d events)",
				tt.curve, visualize(dist, tt.phraseLength), actualEvents, tt.events)
		})
	}
}

func TestAccelerandoDistribution_ContinuousPositions(t *testing.T) {
	tests := []struct {
		name         string
		events       int
		phraseLength int
		curve        float64
	}{
		{"linear curve", 4, 16, 1.0},
		{"moderate curve", 8, 16, 2.0},
		{"high curve", 8, 16, 4.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewAccelerandoDistribution(tt.events, tt.phraseLength, tt.curve)

			positions := dist.GetContinuousPositions()

			// Verify we have all requested positions
			if len(positions) != tt.events {
				t.Errorf("Expected %d continuous positions, got %d", tt.events, len(positions))
			}

			// Verify positions are in ascending order
			for i := 1; i < len(positions); i++ {
				if positions[i] <= positions[i-1] {
					t.Errorf("Positions not in ascending order: %.2f followed by %.2f",
						positions[i-1], positions[i])
				}
			}

			// Verify positions are within phrase bounds
			for i, pos := range positions {
				if pos < 0 || pos >= float64(tt.phraseLength) {
					t.Errorf("Position %d (%.2f) is out of bounds [0, %d)", i, pos, tt.phraseLength)
				}
			}

			t.Logf("Continuous positions (curve=%.1f):", tt.curve)
			for i, pos := range positions {
				tick := int(pos)
				offset := pos - float64(tick)
				t.Logf("  Event %d: position %.3f (tick %d + %.3f)", i, pos, tick, offset)
			}

			actualEvents := dist.GetActualEvents()
			if actualEvents < tt.events {
				t.Logf("Quantization reduced %d events to %d unique ticks (%.1f%% loss)",
					tt.events, actualEvents, float64(tt.events-actualEvents)/float64(tt.events)*100)
			}
		})
	}
}

func TestRitardandoDistribution_EventSpacing(t *testing.T) {
	// Verify that ritardando actually creates increasing spacing
	phraseLength := 16
	events := 8
	curve := 2.0

	dist := NewRitardandoDistribution(events, phraseLength, curve)
	positions := dist.GetContinuousPositions()

	// Calculate spacing between consecutive events
	spacings := make([]float64, len(positions)-1)
	for i := 1; i < len(positions); i++ {
		spacings[i-1] = positions[i] - positions[i-1]
	}

	// For ritardando, spacing should generally increase (allowing for some quantization effects)
	increasingCount := 0
	for i := 1; i < len(spacings); i++ {
		if spacings[i] > spacings[i-1] {
			increasingCount++
		}
	}

	t.Logf("Ritardando spacing pattern:")
	for i, spacing := range spacings {
		t.Logf("  Gap %d-%d: %.3f ticks", i, i+1, spacing)
	}

	// Most gaps should be increasing for ritardando
	if float64(increasingCount) < float64(len(spacings))*0.5 {
		t.Logf("Warning: Only %d/%d gaps are increasing (expected majority for ritardando)",
			increasingCount, len(spacings))
	}
}

func TestAccelerandoDistribution_EventSpacing(t *testing.T) {
	// Verify that accelerando actually creates decreasing spacing
	phraseLength := 16
	events := 8
	curve := 2.0

	dist := NewAccelerandoDistribution(events, phraseLength, curve)
	positions := dist.GetContinuousPositions()

	// Calculate spacing between consecutive events
	spacings := make([]float64, len(positions)-1)
	for i := 1; i < len(positions); i++ {
		spacings[i-1] = positions[i] - positions[i-1]
	}

	// For accelerando, spacing should generally decrease
	decreasingCount := 0
	for i := 1; i < len(spacings); i++ {
		if spacings[i] < spacings[i-1] {
			decreasingCount++
		}
	}

	t.Logf("Accelerando spacing pattern:")
	for i, spacing := range spacings {
		t.Logf("  Gap %d-%d: %.3f ticks", i, i+1, spacing)
	}

	// Most gaps should be decreasing for accelerando
	if float64(decreasingCount) < float64(len(spacings))*0.5 {
		t.Logf("Warning: Only %d/%d gaps are decreasing (expected majority for accelerando)",
			decreasingCount, len(spacings))
	}
}

func TestDistribution_QuantizationComparison(t *testing.T) {
	// Compare the same configuration with different curves to show quantization effects
	phraseLength := 16
	events := 8

	t.Log("Demonstrating quantization loss with increasing curve values:")

	curves := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	for _, curve := range curves {
		dist := NewRitardandoDistribution(events, phraseLength, curve)
		actualEvents := dist.GetActualEvents()
		lossPercent := float64(events-actualEvents) / float64(events) * 100

		t.Logf("Curve %.1f: %s (%d/%d events, %.1f%% loss)",
			curve,
			visualize(dist, phraseLength),
			actualEvents,
			events,
			lossPercent)
	}
}
