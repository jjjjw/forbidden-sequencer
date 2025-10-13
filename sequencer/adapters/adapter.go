package adapters

import "forbidden_sequencer/sequencer"

// EventAdapter is an interface for adapting ScheduledEvents to different protocols
// This allows the sequencer to output to MIDI, OSC, or other software protocols
type EventAdapter interface {
	// Send sends a scheduled event through the adapter's protocol
	Send(scheduled sequencer.ScheduledEvent) error
}
