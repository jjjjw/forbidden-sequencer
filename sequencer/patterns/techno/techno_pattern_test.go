package techno

import (
	"testing"
	"time"

	"forbidden_sequencer/sequencer/conductors"
)

func TestTechnoPattern_ReturnsEventsOnBeatBoundary(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4)
	pattern := NewTechnoPattern(conductor)
	pattern.Play()

	conductor.Start()

	// Advance to tick 3 so next tick (0) is a beat boundary
	conductor.AdvanceTick() // now at 1, next is 2
	conductor.AdvanceTick() // now at 2, next is 3
	conductor.AdvanceTick() // now at 3, next is 0

	nextTickTime := time.Now().Add(125 * time.Millisecond)
	tickDuration := conductor.GetTickDuration()

	events := pattern.GetScheduledEventsForTick(nextTickTime, tickDuration)

	if len(events) != 2 {
		t.Fatalf("Expected 2 events on beat boundary, got %d", len(events))
	}

	// First event should be kick
	if events[0].Event.Name != "kick" {
		t.Errorf("Expected first event to be 'kick', got '%s'", events[0].Event.Name)
	}

	// Second event should be hihat
	if events[1].Event.Name != "hihat" {
		t.Errorf("Expected second event to be 'hihat', got '%s'", events[1].Event.Name)
	}
}

func TestTechnoPattern_ReturnsNilOnNonBeatTicks(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4)
	pattern := NewTechnoPattern(conductor)
	pattern.Play()

	conductor.Start()

	// At tick 0, next tick is 1 (not a beat boundary)
	nextTickTime := time.Now().Add(125 * time.Millisecond)
	tickDuration := conductor.GetTickDuration()

	events := pattern.GetScheduledEventsForTick(nextTickTime, tickDuration)

	if events != nil {
		t.Errorf("Expected nil events on non-beat tick, got %d events", len(events))
	}
}

func TestTechnoPattern_HihatTimingOffset(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4) // 4 ticks per beat
	pattern := NewTechnoPattern(conductor)
	pattern.Play()

	conductor.Start()

	// Advance to tick 3 so next tick (0) is a beat boundary
	conductor.AdvanceTick()
	conductor.AdvanceTick()
	conductor.AdvanceTick()

	nextTickTime := time.Now().Add(125 * time.Millisecond)
	tickDuration := conductor.GetTickDuration()

	events := pattern.GetScheduledEventsForTick(nextTickTime, tickDuration)

	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}

	kick := events[0]
	hihat := events[1]

	// Hihat should be scheduled half a beat after kick
	// Half beat = 2 ticks = 250ms at 120 BPM with 4 ticks/beat
	expectedOffset := tickDuration * 2
	actualOffset := hihat.Timing.Timestamp.Sub(kick.Timing.Timestamp)

	if actualOffset != expectedOffset {
		t.Errorf("Expected hihat offset %v, got %v", expectedOffset, actualOffset)
	}
}

func TestTechnoPattern_PausedReturnsNil(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4)
	pattern := NewTechnoPattern(conductor)
	// Pattern starts paused by default

	conductor.Start()

	nextTickTime := time.Now().Add(125 * time.Millisecond)
	tickDuration := conductor.GetTickDuration()

	events := pattern.GetScheduledEventsForTick(nextTickTime, tickDuration)

	if events != nil {
		t.Errorf("Expected nil events when paused, got %d events", len(events))
	}
}

func TestTechnoPattern_PlayAndStop(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4)
	pattern := NewTechnoPattern(conductor)

	conductor.Start()

	// Advance to tick 3 so next tick (0) is a beat boundary
	conductor.AdvanceTick()
	conductor.AdvanceTick()
	conductor.AdvanceTick()

	nextTickTime := time.Now().Add(125 * time.Millisecond)
	tickDuration := conductor.GetTickDuration()

	// Initially paused
	events := pattern.GetScheduledEventsForTick(nextTickTime, tickDuration)
	if events != nil {
		t.Error("Expected nil events when paused")
	}

	// Play
	pattern.Play()
	events = pattern.GetScheduledEventsForTick(nextTickTime, tickDuration)
	if len(events) != 2 {
		t.Error("Expected 2 events after Play()")
	}

	// Stop
	pattern.Stop()
	events = pattern.GetScheduledEventsForTick(nextTickTime, tickDuration)
	if events != nil {
		t.Error("Expected nil events after Stop()")
	}
}
