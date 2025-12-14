package sequencers

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"forbidden_sequencer/sequencer/conductors"
	"forbidden_sequencer/sequencer/events"
)

// mockPattern is a test pattern that fires a fixed number of times
type mockPattern struct {
	conductor  conductors.Conductor
	eventCount int
	maxEvents  int
	paused     bool
	mu         sync.Mutex
}

func newMockPattern(conductor conductors.Conductor, maxEvents int) *mockPattern {
	return &mockPattern{
		conductor:  conductor,
		eventCount: 0,
		maxEvents:  maxEvents,
		paused:     true,
	}
}

func (m *mockPattern) GetScheduledEventsForTick(nextTickTime time.Time, tickDuration time.Duration) []events.ScheduledEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return nil when paused
	if m.paused {
		return nil
	}

	if m.eventCount >= m.maxEvents {
		// No more events after max
		return nil
	}

	m.eventCount++
	return []events.ScheduledEvent{{
		Event: events.Event{
			Name: fmt.Sprintf("event_%d", m.eventCount),
			Type: events.EventTypeNote,
			A:    60.0,
			B:    0.5,
		},
		Timing: events.Timing{
			Timestamp: nextTickTime,
			Duration:  50 * time.Millisecond,
		},
	}}
}

func (m *mockPattern) getEventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.eventCount
}

func (m *mockPattern) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventCount = 0
}

func (m *mockPattern) Play() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paused = false
}

func (m *mockPattern) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paused = true
}

// mockAdapter tracks events sent to it
type mockAdapter struct {
	events []events.ScheduledEvent
	mu     sync.Mutex
}

func newMockAdapter() *mockAdapter {
	return &mockAdapter{
		events: make([]events.ScheduledEvent, 0),
	}
}

func (m *mockAdapter) Send(scheduled events.ScheduledEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, scheduled)
	return nil
}

func (m *mockAdapter) getEventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, e := range m.events {
		if e.Event.Type != events.EventTypeRest {
			count++
		}
	}
	return count
}

func (m *mockAdapter) getEvents() []events.ScheduledEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]events.ScheduledEvent, 0)
	for _, e := range m.events {
		if e.Event.Type != events.EventTypeRest {
			result = append(result, e)
		}
	}
	return result
}

func TestSequencer_NewSequencer(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4)
	adapter := newMockAdapter()
	pattern := newMockPattern(conductor, 5)
	eventChan := make(chan events.ScheduledEvent, 100)

	seq := NewSequencer([]Pattern{pattern}, conductor, adapter, eventChan)

	if seq == nil {
		t.Fatal("Expected non-nil sequencer")
	}

	if len(seq.patterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(seq.patterns))
	}
}

func TestSequencer_StartAndPatternExecution(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4)
	adapter := newMockAdapter()
	pattern := newMockPattern(conductor, 3) // Generate 3 events
	eventChan := make(chan events.ScheduledEvent, 100)

	seq := NewSequencer([]Pattern{pattern}, conductor, adapter, eventChan)
	seq.Start()
	seq.Play() // Start playback

	// Wait for pattern to generate events
	// At 120 BPM with 4 ticks/beat: tick duration = 125ms
	time.Sleep(300 * time.Millisecond)

	// Check that pattern generated events
	if pattern.getEventCount() < 1 {
		t.Error("Expected pattern to generate at least 1 event")
	}

	// Check that adapter received events
	if adapter.getEventCount() < 1 {
		t.Error("Expected adapter to receive at least 1 event")
	}

	// Verify event names
	receivedEvents := adapter.getEvents()
	if len(receivedEvents) > 0 {
		if receivedEvents[0].Event.Name != "event_1" {
			t.Errorf("Expected first event name 'event_1', got '%s'", receivedEvents[0].Event.Name)
		}
	}
}

func TestSequencer_StopAndPlay(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4)
	adapter := newMockAdapter()
	pattern := newMockPattern(conductor, 100) // Lots of events
	eventChan := make(chan events.ScheduledEvent, 100)

	seq := NewSequencer([]Pattern{pattern}, conductor, adapter, eventChan)
	seq.Start()
	seq.Play() // Start playing

	// Let it run - need to wait for at least one tick (125ms at 120 BPM, 4 ticks/beat)
	time.Sleep(300 * time.Millisecond)

	adapterCount1 := adapter.getEventCount()
	if adapterCount1 < 1 {
		t.Error("Expected adapter to receive events while playing")
	}

	// Stop (patterns stop processing and reset)
	seq.Stop()

	// Small sleep to let any in-flight events complete
	time.Sleep(20 * time.Millisecond)
	adapterCount2 := adapter.getEventCount()

	// Wait longer and verify no more events come through
	time.Sleep(200 * time.Millisecond)
	adapterCount3 := adapter.getEventCount()

	// Adapter should not receive many new events while stopped
	// Allow for up to 2 in-flight events due to scheduling race
	if adapterCount3 > adapterCount2+2 {
		t.Errorf("Expected minimal new events while stopped, got %d after waiting vs %d right after stop", adapterCount3, adapterCount2)
	}

	// Play again (resets conductor and patterns start fresh)
	seq.Play()
	time.Sleep(300 * time.Millisecond)

	adapterCount4 := adapter.getEventCount()
	// Adapter should receive new events after play
	if adapterCount4 <= adapterCount3 {
		t.Error("Expected adapter to receive events after play")
	}
}

func TestSequencer_MultiplePatterns(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4)
	adapter := newMockAdapter()
	pattern1 := newMockPattern(conductor, 2)
	pattern2 := newMockPattern(conductor, 2)
	eventChan := make(chan events.ScheduledEvent, 100)

	seq := NewSequencer([]Pattern{pattern1, pattern2}, conductor, adapter, eventChan)
	seq.Start()
	seq.Play() // Start playback

	// Wait for patterns to generate events - need multiple ticks
	time.Sleep(300 * time.Millisecond)

	// Both patterns should have generated events
	if pattern1.getEventCount() < 1 {
		t.Error("Expected pattern1 to generate at least 1 event")
	}

	if pattern2.getEventCount() < 1 {
		t.Error("Expected pattern2 to generate at least 1 event")
	}

	// Adapter should receive events from both patterns
	totalEvents := adapter.getEventCount()
	if totalEvents < 2 {
		t.Errorf("Expected at least 2 events from both patterns, got %d", totalEvents)
	}
}

func TestSequencer_ConductorIntegration(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4)
	adapter := newMockAdapter()
	pattern := newMockPattern(conductor, 5)
	eventChan := make(chan events.ScheduledEvent, 100)

	seq := NewSequencer([]Pattern{pattern}, conductor, adapter, eventChan)
	seq.Start()
	seq.Play()

	// Wait for at least 2 tick durations (250ms)
	time.Sleep(300 * time.Millisecond)

	// Check that pattern generated events (indicates conductor is running)
	if pattern.getEventCount() < 1 {
		t.Error("Expected pattern to generate events, indicating conductor is running")
	}

	// Stop sequencer
	seq.Stop()

	// Reset and play again
	seq.Reset()
	seq.Play()

	// Give it a moment to run
	time.Sleep(100 * time.Millisecond)

	// Pattern should have generated more events after reset+play
	if adapter.getEventCount() < 2 {
		t.Error("Expected adapter to receive events after reset+play")
	}
}

func TestSequencer_PlayStopPlayCycle(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(120, 4)
	adapter := newMockAdapter()
	pattern := newMockPattern(conductor, 100)
	eventChan := make(chan events.ScheduledEvent, 100)

	seq := NewSequencer([]Pattern{pattern}, conductor, adapter, eventChan)
	seq.Start()

	// First play cycle
	seq.Play()
	time.Sleep(400 * time.Millisecond)

	count1 := adapter.getEventCount()
	if count1 < 1 {
		t.Error("Expected events during first play")
	}

	// Stop
	seq.Stop()
	time.Sleep(50 * time.Millisecond)

	count2 := adapter.getEventCount()

	// Second play cycle - reset then play
	seq.Reset()
	seq.Play()

	// Let it run and verify events are generated
	time.Sleep(300 * time.Millisecond)
	count3 := adapter.getEventCount()

	if count3 <= count2 {
		t.Errorf("Expected new events after second play, got %d events (was %d after first play)", count3, count2)
	}
}
