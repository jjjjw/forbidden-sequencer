package techno

import (
	"testing"

	"forbidden_sequencer/sequencer/conductors"
)

func TestTechnoPattern_FirstEventWithNegativeOffset(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(8, 120)
	pattern := NewTechnoPattern(conductor)
	pattern.Play()

	// First event should be kick
	event, err := pattern.GetNextScheduledEvent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if event.Event.Name != "kick" {
		t.Errorf("Expected first event to be 'kick', got '%s'", event.Event.Name)
	}
}

func TestTechnoPattern_AlternatesKickAndHihat(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(8, 120)
	pattern := NewTechnoPattern(conductor)
	pattern.Play()

	// First scheduled event: kick
	event1, _ := pattern.GetNextScheduledEvent()
	if event1.Event.Name != "kick" {
		t.Errorf("Expected first scheduled event to be 'kick', got '%s'", event1.Event.Name)
	}

	// Second scheduled event: hihat
	event2, _ := pattern.GetNextScheduledEvent()
	if event2.Event.Name != "hihat" {
		t.Errorf("Expected second scheduled event to be 'hihat', got '%s'", event2.Event.Name)
	}

	// Third scheduled event: kick
	event3, _ := pattern.GetNextScheduledEvent()
	if event3.Event.Name != "kick" {
		t.Errorf("Expected third scheduled event to be 'kick', got '%s'", event3.Event.Name)
	}
}

func TestTechnoPattern_Reset(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(8, 120)
	pattern := NewTechnoPattern(conductor)
	pattern.Play()

	// Get first event (kick)
	event1, _ := pattern.GetNextScheduledEvent()
	if event1.Event.Name != "kick" {
		t.Errorf("Expected first event to be 'kick', got '%s'", event1.Event.Name)
	}

	// Get second event (hihat)
	event2, _ := pattern.GetNextScheduledEvent()
	if event2.Event.Name != "hihat" {
		t.Errorf("Expected second event to be 'hihat', got '%s'", event2.Event.Name)
	}

	// Reset pattern
	pattern.Reset()

	// After reset, first event should be kick again
	event3, _ := pattern.GetNextScheduledEvent()
	if event3.Event.Name != "kick" {
		t.Errorf("Expected first event after reset to be 'kick', got '%s'", event3.Event.Name)
	}
}
