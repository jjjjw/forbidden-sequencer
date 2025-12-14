package events

import "time"

// EventType represents the type of event
type EventType string

const (
	EventTypeNote       EventType = "note"       // Note event (can have freq or midi_note in Params)
	EventTypeModulation EventType = "modulation" // Modulation/CC event
	EventTypeRest       EventType = "rest"       // Rest/silence event
)

// Event represents an abstract musical event in the sequencer
// Parameters are stored as a flexible map of string keys to float32 values.
//
// Common parameter names:
//   - freq: frequency in Hz (e.g., 440.0 for A4)
//   - midi_note: MIDI note number (0-127, e.g., 69.0 for A4)
//   - amp: amplitude/velocity (0.0-1.0 normalized)
//   - len: duration in seconds (optional, Timing.Duration is used by default)
//
// For EventTypeNote:
//   - Use either "freq" OR "midi_note" (adapters will convert as needed)
//   - Common params: amp, len, and any synth-specific parameters
//
// For EventTypeModulation (MIDI CC):
//   - cc_num: CC number (e.g., 1.0 for mod wheel)
//   - cc_value: CC value (0.0-1.0 normalized)
//
// For EventTypeRest:
//   - No parameters needed (rest is a no-op)
type Event struct {
	Name   string             // Event identifier/name
	Type   EventType          // Type of event
	Params map[string]float32 // Flexible parameter dictionary
}

// Timing represents when and how long an event should play
type Timing struct {
	Timestamp time.Time     // Absolute time of the event start
	Duration  time.Duration // How long event lasts (could be different than the delta between events)
}

// ScheduledEvent pairs an Event with Timing information
type ScheduledEvent struct {
	Event  Event
	Timing Timing
}
