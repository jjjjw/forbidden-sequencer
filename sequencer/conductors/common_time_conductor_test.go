package conductors

import (
	"testing"
	"time"
)

func TestCommonTimeConductor_TickAdvancement(t *testing.T) {
	c := NewCommonTimeConductor(4, 120) // 4 ticks/beat, 120 BPM

	// Initial tick should be 0
	if c.GetCurrentTick() != 0 {
		t.Errorf("Expected initial tick to be 0, got %d", c.GetCurrentTick())
	}

	// Advance ticks manually
	c.AdvanceTick()
	if c.GetCurrentTick() != 1 {
		t.Errorf("Expected tick to be 1 after advance, got %d", c.GetCurrentTick())
	}

	c.AdvanceTick()
	c.AdvanceTick()
	if c.GetCurrentTick() != 3 {
		t.Errorf("Expected tick to be 3 after 3 advances, got %d", c.GetCurrentTick())
	}
}

func TestCommonTimeConductor_MusicalTimeHelpers(t *testing.T) {
	c := NewCommonTimeConductor(4, 120) // 4 ticks/beat

	// Tick 0: should be beat start
	if !c.IsBeatStart() {
		t.Error("Expected tick 0 to be beat start")
	}

	if c.GetTickInBeat() != 0 {
		t.Errorf("Expected tick 0 to be position 0 in beat, got %d", c.GetTickInBeat())
	}

	if c.GetBeat() != 0 {
		t.Errorf("Expected tick 0 to be beat 0, got %d", c.GetBeat())
	}

	// Advance to tick 2 (middle of beat)
	c.AdvanceTick()
	c.AdvanceTick()

	if c.IsBeatStart() {
		t.Error("Expected tick 2 to not be beat start")
	}

	if c.GetTickInBeat() != 2 {
		t.Errorf("Expected tick 2 to be position 2 in beat, got %d", c.GetTickInBeat())
	}

	if c.GetBeat() != 0 {
		t.Errorf("Expected tick 2 to still be beat 0, got %d", c.GetBeat())
	}

	// Advance to tick 4 (start of beat 1)
	c.AdvanceTick()
	c.AdvanceTick()

	if !c.IsBeatStart() {
		t.Error("Expected tick 4 to be beat start")
	}

	if c.GetTickInBeat() != 0 {
		t.Errorf("Expected tick 4 to be position 0 in beat, got %d", c.GetTickInBeat())
	}

	if c.GetBeat() != 1 {
		t.Errorf("Expected tick 4 to be beat 1, got %d", c.GetBeat())
	}
}

func TestCommonTimeConductor_BPMCalculation(t *testing.T) {
	c := NewCommonTimeConductor(4, 120) // 4 ticks/beat, 120 BPM

	// At 120 BPM: 60/120 = 0.5 seconds per beat
	// 0.5 / 4 = 0.125 seconds per tick = 125ms
	expectedDuration := 125 * time.Millisecond

	if c.GetTickDuration() != expectedDuration {
		t.Errorf("Expected tick duration %v, got %v", expectedDuration, c.GetTickDuration())
	}

	// Change BPM to 60
	c.SetBPM(60)

	// At 60 BPM: 60/60 = 1 second per beat
	// 1 / 4 = 0.25 seconds per tick = 250ms
	expectedDuration = 250 * time.Millisecond

	if c.GetTickDuration() != expectedDuration {
		t.Errorf("Expected tick duration %v after BPM change, got %v", expectedDuration, c.GetTickDuration())
	}

	if c.GetBPM() != 60 {
		t.Errorf("Expected BPM to be 60, got %f", c.GetBPM())
	}
}

func TestCommonTimeConductor_Reset(t *testing.T) {
	c := NewCommonTimeConductor(4, 120)

	// Advance a few ticks
	c.AdvanceTick()
	c.AdvanceTick()
	c.AdvanceTick()

	if c.GetCurrentTick() != 3 {
		t.Errorf("Expected tick to be 3, got %d", c.GetCurrentTick())
	}

	// Reset
	c.Reset()

	if c.GetCurrentTick() != 0 {
		t.Errorf("Expected tick to be 0 after reset, got %d", c.GetCurrentTick())
	}
}

func TestCommonTimeConductor_StartAndPause(t *testing.T) {
	c := NewCommonTimeConductor(4, 120)

	c.Start()
	defer c.Pause() // Stop the goroutine after test

	// Wait a bit for ticks to advance
	time.Sleep(300 * time.Millisecond) // Should advance ~2 ticks at 125ms/tick

	tick1 := c.GetCurrentTick()
	if tick1 < 1 {
		t.Error("Expected conductor to advance ticks while running")
	}

	// Pause
	c.Pause()
	time.Sleep(200 * time.Millisecond)

	tick2 := c.GetCurrentTick()
	if tick2 != tick1 {
		t.Error("Expected ticks to stop advancing while paused")
	}

	// Resume
	c.Resume()
	time.Sleep(300 * time.Millisecond)

	tick3 := c.GetCurrentTick()
	if tick3 <= tick2 {
		t.Error("Expected ticks to resume advancing after resume")
	}
}
