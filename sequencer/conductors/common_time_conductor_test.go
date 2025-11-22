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

func TestCommonTimeConductor_BPMCalculation(t *testing.T) {
	c := NewCommonTimeConductor(4, 120) // 4 ticks/beat, 120 BPM

	// At 120 BPM: 60/120 = 0.5 seconds per beat
	// 0.5 / 4 = 0.125 seconds per tick = 125ms
	expectedDuration := 125 * time.Millisecond

	if c.GetTickDuration() != expectedDuration {
		t.Errorf("Expected tick duration %v, got %v", expectedDuration, c.GetTickDuration())
	}
}

func TestCommonTimeConductor_GetNextBeatTick(t *testing.T) {
	c := NewCommonTimeConductor(4, 120) // 4 ticks/beat

	// At tick 0, next beat is at tick 4
	if c.GetNextBeatTick() != 4 {
		t.Errorf("Expected next beat at tick 4, got %d", c.GetNextBeatTick())
	}

	// Advance to tick 1
	c.AdvanceTick()
	if c.GetNextBeatTick() != 4 {
		t.Errorf("Expected next beat at tick 4, got %d", c.GetNextBeatTick())
	}

	// Advance to tick 4 (on beat boundary)
	c.AdvanceTick()
	c.AdvanceTick()
	c.AdvanceTick()
	if c.GetNextBeatTick() != 8 {
		t.Errorf("Expected next beat at tick 8, got %d", c.GetNextBeatTick())
	}
}

func TestCommonTimeConductor_Start(t *testing.T) {
	c := NewCommonTimeConductor(4, 120)

	c.Start()

	// Wait a bit for ticks to advance
	time.Sleep(300 * time.Millisecond) // Should advance ~2 ticks at 125ms/tick

	tick1 := c.GetCurrentTick()
	if tick1 < 1 {
		t.Error("Expected conductor to advance ticks while running")
	}

	// Wait more and verify ticks continue to advance
	time.Sleep(300 * time.Millisecond)

	tick2 := c.GetCurrentTick()
	if tick2 <= tick1 {
		t.Error("Expected ticks to continue advancing")
	}
}
