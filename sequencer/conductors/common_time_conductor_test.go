package conductors

import (
	"testing"
	"time"
)

func TestCommonTimeConductor_TickAdvancement(t *testing.T) {
	c := NewCommonTimeConductor(120, 4) // 120 BPM, 4 ticks/beat

	// Initial tick should be 0
	if c.GetTickInBeat() != 0 {
		t.Errorf("Expected initial tick in beat to be 0, got %d", c.GetTickInBeat())
	}

	// Start conductor to initialize times
	c.Start()

	// Advance ticks manually
	c.AdvanceTick()
	if c.GetTickInBeat() != 1 {
		t.Errorf("Expected tick in beat to be 1 after advance, got %d", c.GetTickInBeat())
	}

	c.AdvanceTick()
	c.AdvanceTick()
	if c.GetTickInBeat() != 3 {
		t.Errorf("Expected tick in beat to be 3 after 3 advances, got %d", c.GetTickInBeat())
	}

	// Advance one more to complete the beat
	c.AdvanceTick()
	if c.GetTickInBeat() != 0 {
		t.Errorf("Expected tick in beat to wrap to 0, got %d", c.GetTickInBeat())
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

	c.Start()

	// Wait a bit for ticks to advance
	time.Sleep(300 * time.Millisecond) // Should advance ~2 ticks at 125ms/tick

	// Check that beats channel received notifications
	// At 120 BPM, 4 ticks/beat = 500ms per beat
	// In 300ms we won't complete a beat yet

	// Wait longer to complete at least one beat
	time.Sleep(300 * time.Millisecond)

	// Check if we received a beat notification
	select {
	case <-c.Beats:
		// Good, received a beat
	default:
		t.Error("Expected to receive at least one beat notification")
	}
}

func TestCommonTimeConductor_BeatNotification(t *testing.T) {
	c := NewCommonTimeConductor(120, 4) // 120 BPM, 4 ticks/beat

	c.Start()

	// Manually advance 4 ticks to trigger a beat
	c.AdvanceTick()
	c.AdvanceTick()
	c.AdvanceTick()
	c.AdvanceTick()

	// Should have received a beat notification
	select {
	case <-c.Beats:
		// Good
	default:
		t.Error("Expected beat notification after 4 ticks")
	}
}
