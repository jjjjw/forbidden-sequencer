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
	mu         sync.Mutex
}

func newMockPattern(conductor conductors.Conductor, maxEvents int) *mockPattern {
	return &mockPattern{
		conductor:  conductor,
		eventCount: 0,
		maxEvents:  maxEvents,
	}
}

func (m *mockPattern) GetNextScheduledEvent() (events.ScheduledEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.eventCount >= m.maxEvents {
		// Return a large delta to slow down after max events
		return events.ScheduledEvent{
			Event: events.Event{Name: "done"},
			Timing: events.Timing{
				Delta:    10 * time.Second,
				Duration: 0,
			},
		}, nil
	}

	m.eventCount++
	return events.ScheduledEvent{
		Event: events.Event{
			Name: fmt.Sprintf("event_%d", m.eventCount),
			Type: events.EventTypeNote,
			A:    60.0,
			B:    0.5,
		},
		Timing: events.Timing{
			Delta:    10 * time.Millisecond,
			Duration: 50 * time.Millisecond,
		},
	}, nil
}

func (m *mockPattern) getEventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.eventCount
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
	return len(m.events)
}

func (m *mockAdapter) getEvents() []events.ScheduledEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]events.ScheduledEvent, len(m.events))
	copy(result, m.events)
	return result
}

func TestSequencer_NewSequencer(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
	adapter := newMockAdapter()
	pattern := newMockPattern(conductor, 5)

	seq := NewSequencer([]Pattern{pattern}, conductor, adapter, false)

	if seq == nil {
		t.Fatal("Expected non-nil sequencer")
	}

	if seq.running {
		t.Error("Expected sequencer to not be running initially")
	}

	if len(seq.patterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(seq.patterns))
	}
}

func TestSequencer_StartAndPatternExecution(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
	adapter := newMockAdapter()
	pattern := newMockPattern(conductor, 3) // Generate 3 events

	seq := NewSequencer([]Pattern{pattern}, conductor, adapter, false)
	seq.Start()

	// Wait for pattern to generate events
	// At 120 BPM with 4 ticks/beat: tick duration = 125ms
	time.Sleep(300 * time.Millisecond)

	// Check that conductor is running
	if conductor.GetCurrentTick() < 1 {
		t.Errorf("Expected conductor to be running and advancing ticks, got tick %d", conductor.GetCurrentTick())
	}

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

func TestSequencer_PauseAndResume(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
	adapter := newMockAdapter()
	pattern := newMockPattern(conductor, 100) // Lots of events

	seq := NewSequencer([]Pattern{pattern}, conductor, adapter, false)
	seq.Start()

	// Let it run briefly (at least 2 tick durations = 250ms)
	time.Sleep(300 * time.Millisecond)

	// Get current tick
	tick1 := conductor.GetCurrentTick()
	if tick1 < 1 {
		t.Errorf("Expected conductor to advance ticks while running, got tick %d", tick1)
	}

	// Pause
	seq.Pause()
	time.Sleep(200 * time.Millisecond)

	tick2 := conductor.GetCurrentTick()
	if tick2 != tick1 {
		t.Error("Expected conductor to stop advancing ticks while paused")
	}

	// Resume
	seq.Resume()
	time.Sleep(300 * time.Millisecond)

	tick3 := conductor.GetCurrentTick()
	if tick3 <= tick2 {
		t.Error("Expected conductor to resume advancing ticks after resume")
	}
}

func TestSequencer_MultiplePatterns(t *testing.T) {
	conductor := conductors.NewCommonTimeConductor(4, 120)
	adapter := newMockAdapter()
	pattern1 := newMockPattern(conductor, 2)
	pattern2 := newMockPattern(conductor, 2)

	seq := NewSequencer([]Pattern{pattern1, pattern2}, conductor, adapter, false)
	seq.Start()

	// Wait for patterns to generate events
	time.Sleep(100 * time.Millisecond)

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
	conductor := conductors.NewCommonTimeConductor(4, 120)
	adapter := newMockAdapter()
	pattern := newMockPattern(conductor, 5)

	// Verify conductor starts at tick 0
	if conductor.GetCurrentTick() != 0 {
		t.Error("Expected conductor to start at tick 0")
	}

	seq := NewSequencer([]Pattern{pattern}, conductor, adapter, false)
	seq.Start()

	// Wait for at least 2 tick durations (250ms)
	time.Sleep(300 * time.Millisecond)

	// Verify conductor has advanced
	if conductor.GetCurrentTick() < 1 {
		t.Errorf("Expected conductor to advance after sequencer start, got tick %d", conductor.GetCurrentTick())
	}

	// Pause sequencer
	seq.Pause()
	pausedTick := conductor.GetCurrentTick()

	// Wait
	time.Sleep(200 * time.Millisecond)

	// Verify conductor stopped advancing
	if conductor.GetCurrentTick() != pausedTick {
		t.Error("Expected conductor to stop when sequencer paused")
	}
}
