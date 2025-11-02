package techno

import (
	"testing"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

func TestHihatPattern_FiresOnOffBeats(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120) // 4 ticks/beat, 120 BPM
	hihat := NewHihatPattern(conductor)

	// First call should fire on tick 2 (half-beat, off-beat)
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

	// First fire tick should be 2 (half-beat)
	if hihat.lastFireTick != 2 {
		t.Errorf("Expected first fire tick to be 2, got %d", hihat.lastFireTick)
	}

	// Delta should be 2 ticks (from tick 0 to tick 2)
	expectedDelta := conductor.GetTickDuration() * 2
	if scheduled.Timing.Delta != expectedDelta {
		t.Errorf("Expected delta %v, got %v", expectedDelta, scheduled.Timing.Delta)
	}

	// Second call should schedule for next off-beat (tick 6)
	scheduled, err = hihat.GetNextScheduledEvent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Delta should be 6 ticks (from tick 0 to tick 6)
	expectedDelta = conductor.GetTickDuration() * 6
	if scheduled.Timing.Delta != expectedDelta {
		t.Errorf("Expected delta %v for next off-beat, got %v", expectedDelta, scheduled.Timing.Delta)
	}

	if hihat.lastFireTick != 6 {
		t.Errorf("Expected second fire tick to be 6, got %d", hihat.lastFireTick)
	}
}

func TestHihatPattern_ConsistentBeatInterval(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
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

	// Verify the delta difference
	deltaOffset := hihatEvent.Timing.Delta - kickEvent.Timing.Delta
	expectedOffset := conductor.GetTickDuration() * 2
	if deltaOffset != expectedOffset {
		t.Errorf("Expected delta offset of %v, got %v", expectedOffset, deltaOffset)
	}
}

func TestHihatPattern_Duration(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
	hihat := NewHihatPattern(conductor)

	scheduled, _ := hihat.GetNextScheduledEvent()

	expectedDuration := 50 * time.Millisecond
	if scheduled.Timing.Duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, scheduled.Timing.Duration)
	}
}
