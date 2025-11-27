package conductors

import (
	"testing"
	"time"
)

func TestCommonTimeConductor_TickAdvancement(t *testing.T) {
	c := NewCommonTimeConductor(120, 4) // 120 BPM, 4 ticks/beat

	// Start conductor to initialize times
	c.Start()

	// Initial state: current tick is 0, so next tick in beat should be 1
	if c.GetNextTickInBeat() != 1 {
		t.Errorf("Expected next tick in beat to be 1, got %d", c.GetNextTickInBeat())
	}

	// Advance ticks manually
	c.AdvanceTick() // now at tick 1, next is 2
	if c.GetNextTickInBeat() != 2 {
		t.Errorf("Expected next tick in beat to be 2 after advance, got %d", c.GetNextTickInBeat())
	}

	c.AdvanceTick() // now at tick 2, next is 3
	c.AdvanceTick() // now at tick 3, next is 0 (wraps)
	if c.GetNextTickInBeat() != 0 {
		t.Errorf("Expected next tick in beat to be 0 after 3 advances, got %d", c.GetNextTickInBeat())
	}

	// Advance one more to complete the beat
	c.AdvanceTick() // now at tick 0, next is 1
	if c.GetNextTickInBeat() != 1 {
		t.Errorf("Expected next tick in beat to be 1, got %d", c.GetNextTickInBeat())
	}
}

func TestCommonTimeConductor_BPMCalculation(t *testing.T) {
	c := NewCommonTimeConductor(120, 4) // 120 BPM, 4 ticks/beat

	// At 120 BPM: 60/120 = 0.5 seconds per beat
	// 0.5 / 4 = 0.125 seconds per tick = 125ms
	expectedDuration := 125 * time.Millisecond

	if c.GetTickDuration() != expectedDuration {
		t.Errorf("Expected tick duration %v, got %v", expectedDuration, c.GetTickDuration())
	}
}

func TestCommonTimeConductor_GetNextBeatTime(t *testing.T) {
	c := NewCommonTimeConductor(120, 4) // 120 BPM, 4 ticks/beat

	// Start conductor
	c.Start()
	startTime := c.GetLastBeatTime()

	// Next beat should be one beat duration away
	beatDuration := c.GetTickDuration() * time.Duration(c.GetTicksPerBeat())
	expectedNextBeat := startTime.Add(beatDuration)

	nextBeat := c.GetNextBeatTime()
	if !nextBeat.Equal(expectedNextBeat) {
		t.Errorf("Expected next beat at %v, got %v", expectedNextBeat, nextBeat)
	}

	// Advance to complete a beat
	c.AdvanceTick()
	c.AdvanceTick()
	c.AdvanceTick()
	c.AdvanceTick()

	// Now next beat should be another beat duration away from the new lastBeatTime
	newBeatTime := c.GetLastBeatTime()
	expectedNextBeat2 := newBeatTime.Add(beatDuration)
	nextBeat2 := c.GetNextBeatTime()

	if !nextBeat2.Equal(expectedNextBeat2) {
		t.Errorf("Expected next beat at %v, got %v", expectedNextBeat2, nextBeat2)
	}
}

func TestCommonTimeConductor_Start(t *testing.T) {
	c := NewCommonTimeConductor(120, 4) // 120 BPM, 4 ticks/beat

	// Subscribe to ticks BEFORE starting
	tickChan := c.Ticks()

	c.Start()

	// Wait a bit for ticks to advance
	// At 120 BPM with 4 ticks/beat: tick duration = 125ms
	time.Sleep(200 * time.Millisecond) // Should advance ~1-2 ticks

	// Check if we received tick notifications
	select {
	case <-tickChan:
		// Good, received a tick
	default:
		t.Error("Expected to receive at least one tick notification")
	}
}

func TestCommonTimeConductor_TickNotification(t *testing.T) {
	c := NewCommonTimeConductor(120, 4) // 120 BPM, 4 ticks/beat

	// Subscribe to ticks BEFORE starting
	tickChan := c.Ticks()

	c.Start()

	// Manually advance a tick
	c.AdvanceTick()

	// Should have received a tick notification
	select {
	case <-tickChan:
		// Good
	default:
		t.Error("Expected tick notification after advance")
	}
}
