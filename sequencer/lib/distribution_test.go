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
