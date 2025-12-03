package adapters

import (
	"fmt"

	"forbidden_sequencer/sequencer/events"

	"github.com/hypebeast/go-osc/osc"
)

// OSCAdapter implements EventAdapter for OSC output
type OSCAdapter struct {
	client         *osc.Client
	host           string
	port           int
	addressMapping map[string]string // maps event names to OSC addresses
}

// NewOSCAdapter creates a new OSC adapter
// host: target host (e.g., "localhost" or "127.0.0.1")
// port: target port (e.g., 57120 for SuperCollider default)
func NewOSCAdapter(host string, port int) (*OSCAdapter, error) {
	client := osc.NewClient(host, port)

	return &OSCAdapter{
		client:         client,
		host:           host,
		port:           port,
		addressMapping: make(map[string]string),
	}, nil
}

// SetAddressMapping sets the OSC address for a given event name
// For example: SetAddressMapping("kick", "/trigger/kick")
func (o *OSCAdapter) SetAddressMapping(eventName string, address string) {
	o.addressMapping[eventName] = address
}

// GetAddressMapping returns the OSC address for a given event name
// Returns a default pattern if not mapped
func (o *OSCAdapter) GetAddressMapping(eventName string) string {
	if address, ok := o.addressMapping[eventName]; ok {
		return address
	}
	// Default pattern: /trigger/<eventname>
	return fmt.Sprintf("/trigger/%s", eventName)
}

// GetAllAddressMappings returns all address mappings
func (o *OSCAdapter) GetAllAddressMappings() map[string]string {
	result := make(map[string]string)
	for k, v := range o.addressMapping {
		result[k] = v
	}
	return result
}

// GetHost returns the current OSC host
func (o *OSCAdapter) GetHost() string {
	return o.host
}

// GetPort returns the current OSC port
func (o *OSCAdapter) GetPort() int {
	return o.port
}

// SetTarget changes the OSC target host and port
func (o *OSCAdapter) SetTarget(host string, port int) {
	o.host = host
	o.port = port
	o.client = osc.NewClient(host, port)
}

// Send implements EventAdapter.Send
// Uses OSC bundles with timestamps for precise timing
func (o *OSCAdapter) Send(scheduled events.ScheduledEvent) error {
	switch scheduled.Event.Type {
	case events.EventTypeNote:
		return o.sendNote(scheduled)
	case events.EventTypeModulation:
		return o.sendModulation(scheduled)
	case events.EventTypeRest:
		// Rest is a no-op
		return nil
	}
	return nil
}

// sendNote sends OSC bundle with timestamp for note events
// Bundle contains the timestamp, message contains all event parameters
// Message format: <address> <freq> <velocity> <duration> [<c> <d>]
func (o *OSCAdapter) sendNote(scheduled events.ScheduledEvent) error {
	event := scheduled.Event
	timing := scheduled.Timing

	// Get OSC address from mapping
	address := o.GetAddressMapping(event.Name)

	// Build OSC message with all parameters
	msg := osc.NewMessage(address)
	msg.Append(event.A)                            // frequency or note number (float32)
	msg.Append(event.B)                            // velocity (0.0-1.0)
	msg.Append(float32(timing.Duration.Seconds())) // duration in seconds
	msg.Append(event.C)                            // additional parameter C
	msg.Append(event.D)                            // additional parameter D

	// Create bundle with timestamp for precise timing
	bundle := osc.NewBundle(timing.Timestamp)
	bundle.Append(msg)

	// Send the bundle
	err := o.client.Send(bundle)
	if err != nil {
		return fmt.Errorf("failed to send OSC note bundle: %w", err)
	}

	return nil
}

// sendModulation sends OSC bundle with timestamp for modulation/CC events
// Message format: <address> <param> <value> [<c> <d>]
func (o *OSCAdapter) sendModulation(scheduled events.ScheduledEvent) error {
	event := scheduled.Event
	timing := scheduled.Timing

	// Get OSC address from mapping
	address := o.GetAddressMapping(event.Name)

	// Build OSC message
	msg := osc.NewMessage(address)
	msg.Append(event.A) // parameter number
	msg.Append(event.B) // value (0.0-1.0)
	msg.Append(event.C) // additional parameter C
	msg.Append(event.D) // additional parameter D

	// Create bundle with timestamp
	bundle := osc.NewBundle(timing.Timestamp)
	bundle.Append(msg)

	// Send the bundle
	err := o.client.Send(bundle)
	if err != nil {
		return fmt.Errorf("failed to send OSC modulation bundle: %w", err)
	}

	return nil
}

// Close closes the OSC adapter (no-op for OSC, included for interface compatibility)
func (o *OSCAdapter) Close() error {
	// OSC clients don't need cleanup
	return nil
}
