package techno

import (
	"testing"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

func TestKickPattern_FiresOnBeats(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120) // 4 ticks/beat, 120 BPM
	conductor.Reset()                                       // Initialize start time without starting tick loop
	kick := NewKickPattern(conductor)

	// First call fires at current + beat (tick 0, nextFireTick = 0+4 = 4)
	scheduled, err := kick.GetNextScheduledEvent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if scheduled.Event.Name != "kick" {
		t.Errorf("Expected event name 'kick', got '%s'", scheduled.Event.Name)
	}

	if scheduled.Event.Type != events.EventTypeNote {
		t.Errorf("Expected event type Note, got %v", scheduled.Event.Type)
	}

	if scheduled.Event.A != 36.0 {
		t.Errorf("Expected MIDI note 36, got %f", scheduled.Event.A)
	}

	// Delta should be 4 ticks (from tick 0 to tick 4) - with tolerance for timing precision
	expectedDelta := conductor.GetTickDuration() * 4
	tolerance := 1 * time.Millisecond
	if scheduled.Timing.Delta < expectedDelta-tolerance || scheduled.Timing.Delta > expectedDelta+tolerance {
		t.Errorf("Expected delta %v (±%v), got %v", expectedDelta, tolerance, scheduled.Timing.Delta)
	}

	// Second call should schedule for next beat (tick 8)
	scheduled, err = kick.GetNextScheduledEvent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Delta should be 8 ticks (from tick 0 to tick 8) - with tolerance for timing precision
	expectedDelta = conductor.GetTickDuration() * 8
	if scheduled.Timing.Delta < expectedDelta-tolerance || scheduled.Timing.Delta > expectedDelta+tolerance {
		t.Errorf("Expected delta %v (±%v) for next beat, got %v", expectedDelta, tolerance, scheduled.Timing.Delta)
	}
}

func TestKickPattern_ConsistentBeatInterval(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
	conductor.Reset()
	kick := NewKickPattern(conductor)

	// Generate several events and verify they're all one beat apart
	var lastFireTick int64 = -1

	for i := 0; i < 5; i++ {
		_, err := kick.GetNextScheduledEvent()
		if err != nil {
			t.Fatalf("Unexpected error on iteration %d: %v", i, err)
		}

		currentFireTick := kick.lastFireTick

		if lastFireTick >= 0 {
			interval := currentFireTick - lastFireTick
			if interval != 4 {
				t.Errorf("Expected interval of 4 ticks between kicks, got %d", interval)
			}
		}

		lastFireTick = currentFireTick
	}
}

func TestKickPattern_Duration(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
	kick := NewKickPattern(conductor)

	scheduled, _ := kick.GetNextScheduledEvent()

	expectedDuration := 100 * time.Millisecond
	if scheduled.Timing.Duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, scheduled.Timing.Duration)
	}
}
