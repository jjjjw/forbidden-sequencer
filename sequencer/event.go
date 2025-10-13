package sequencer

import "time"

// EventType represents the type of event
type EventType string

const (
	EventTypeNote       EventType = "note"
	EventTypeModulation EventType = "modulation"
	EventTypeRest       EventType = "rest"
)

// Event represents an abstract musical event in the sequencer
// All parameters are float32 for maximum flexibility across protocols.
// The meaning of parameters a, b, c, d depends on the event type:
//
// For EventTypeNote:
//   - a: frequency in Hz (e.g., 440.0 for A4)
//   - b: velocity/intensity (0.0-1.0 normalized)
//   - c: reserved for future use
//   - d: reserved for future use
//
// For EventTypeModulation (MIDI CC):
//   - a: CC number (e.g., 1.0 for mod wheel)
//   - b: CC value (0.0-1.0 normalized)
//   - c: reserved for future use
//   - d: reserved for future use
//
// For EventTypeRest:
//   - No parameters used (rest is a no-op)
type Event struct {
	Name string    // Event identifier/name
	Type EventType // Type of event
	A    float32   // First parameter (meaning depends on type)
	B    float32   // Second parameter (meaning depends on type)
	C    float32   // Third parameter (reserved for future use)
	D    float32   // Fourth parameter (reserved for future use)
}

// Timing represents when and how long an event should play
type Timing struct {
	Wait     time.Duration // Wait before triggering
	Duration time.Duration // How long event lasts
}

// ScheduledEvent pairs an Event with Timing information
type ScheduledEvent struct {
	Event  Event
	Timing Timing
}
