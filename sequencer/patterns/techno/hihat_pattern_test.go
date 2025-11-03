package techno

import (
	"testing"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

func TestHihatPattern_FiresOnOffBeats(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120) // 4 ticks/beat, 120 BPM
	conductor.Reset()                                       // Initialize start time without starting tick loop
	hihat := NewHihatPattern(conductor)

	// First call fires at current + half beat (tick 0 + 2 = tick 2), but calculates next as 2+4=6
	scheduled, err := hihat.GetNextScheduledEvent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if scheduled.Event.Name != "hihat" {
		t.Errorf("Expected event name 'hihat', got '%s'", scheduled.Event.Name)
	}

	if scheduled.Event.Type != events.EventTypeNote {
		t.Errorf("Expected event type Note, got %v", scheduled.Event.Type)
	}

	if scheduled.Event.A != 42.0 {
		t.Errorf("Expected MIDI note 42, got %f", scheduled.Event.A)
	}

	// After first call, lastFireTick should be 6 (initialized at 2, then set to nextFireTick=6)
	if hihat.lastFireTick != 6 {
		t.Errorf("Expected first fire tick to be 6, got %d", hihat.lastFireTick)
	}

	// Delta should be 6 ticks (from tick 0 to tick 6) - with tolerance for timing precision
	expectedDelta := conductor.GetTickDuration() * 6
	tolerance := 1 * time.Millisecond
	if scheduled.Timing.Delta < expectedDelta-tolerance || scheduled.Timing.Delta > expectedDelta+tolerance {
		t.Errorf("Expected delta %v (±%v), got %v", expectedDelta, tolerance, scheduled.Timing.Delta)
	}

	// Second call should schedule for next off-beat (tick 10)
	scheduled, err = hihat.GetNextScheduledEvent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Delta should be 10 ticks (from tick 0 to tick 10) - with tolerance for timing precision
	expectedDelta = conductor.GetTickDuration() * 10
	if scheduled.Timing.Delta < expectedDelta-tolerance || scheduled.Timing.Delta > expectedDelta+tolerance {
		t.Errorf("Expected delta %v (±%v) for next off-beat, got %v", expectedDelta, tolerance, scheduled.Timing.Delta)
	}

	if hihat.lastFireTick != 10 {
		t.Errorf("Expected second fire tick to be 10, got %d", hihat.lastFireTick)
	}
}

func TestHihatPattern_ConsistentBeatInterval(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
	conductor.Reset()
	hihat := NewHihatPattern(conductor)

	// Generate several events and verify they're all one beat apart
	var lastFireTick int64 = -100 // Sentinel value

	for i := 0; i < 5; i++ {
		_, err := hihat.GetNextScheduledEvent()
		if err != nil {
			t.Fatalf("Unexpected error on iteration %d: %v", i, err)
		}

		currentFireTick := hihat.lastFireTick

		if lastFireTick != -100 {
			interval := currentFireTick - lastFireTick
			if interval != 4 {
				t.Errorf("Expected interval of 4 ticks between hihats, got %d", interval)
			}
		}

		lastFireTick = currentFireTick
	}
}

func TestHihatPattern_OffsetFromKick(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
	conductor.Reset()
	kick := NewKickPattern(conductor)
	hihat := NewHihatPattern(conductor)

	// Get first events
	kickEvent, _ := kick.GetNextScheduledEvent()
	hihatEvent, _ := hihat.GetNextScheduledEvent()

	// Hihat should fire 2 ticks after kick (half-beat offset)
	if hihat.lastFireTick-kick.lastFireTick != 2 {
		t.Errorf("Expected hihat to fire 2 ticks after kick, got offset of %d",
			hihat.lastFireTick-kick.lastFireTick)
	}

	// Verify the delta difference (with small tolerance for timing precision)
	deltaOffset := hihatEvent.Timing.Delta - kickEvent.Timing.Delta
	expectedOffset := conductor.GetTickDuration() * 2
	tolerance := 1 * time.Millisecond
	if deltaOffset < expectedOffset-tolerance || deltaOffset > expectedOffset+tolerance {
		t.Errorf("Expected delta offset of %v (±%v), got %v", expectedOffset, tolerance, deltaOffset)
	}
}

func TestHihatPattern_Duration(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
	conductor.Reset()
	hihat := NewHihatPattern(conductor)

	scheduled, _ := hihat.GetNextScheduledEvent()

	expectedDuration := 50 * time.Millisecond
	if scheduled.Timing.Duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, scheduled.Timing.Duration)
	}
}
