package adapters

import "forbidden_sequencer/sequencer"

// EventAdapter is an interface for adapting Events to different protocols
// This allows the sequencer to output to MIDI, OSC, or other software protocols
type EventAdapter interface {
	// Send sends an event through the adapter's protocol
	Send(event sequencer.Event) error
}
