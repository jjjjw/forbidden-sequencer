package adapters

import (
	"fmt"

	"forbidden_sequencer/sequencer/events"

	"github.com/hypebeast/go-osc/osc"
)

// SuperColliderAdapter implements EventAdapter for SuperCollider server commands
// Sends OSC messages directly to scsynth (port 57110) using server command protocol
type SuperColliderAdapter struct {
	client          *osc.Client
	host            string
	port            int
	synthDefMapping map[string]string // maps event names to SynthDef names
	groupIDMapping  map[string]int32  // maps event names to Group IDs
	busIDMapping    map[string]int32  // maps event names to output bus IDs
}

// NewSuperColliderAdapter creates a new SuperCollider adapter
// host: target host (e.g., "localhost")
// port: scsynth port (default: 57110)
func NewSuperColliderAdapter(host string, port int) (*SuperColliderAdapter, error) {
	client := osc.NewClient(host, port)

	return &SuperColliderAdapter{
		client:          client,
		host:            host,
		port:            port,
		synthDefMapping: make(map[string]string),
		groupIDMapping:  make(map[string]int32),
		busIDMapping:    make(map[string]int32),
	}, nil
}

// SetSynthDefMapping sets the SynthDef name for a given event name
// For example: SetSynthDefMapping("kick", "bd")
func (sc *SuperColliderAdapter) SetSynthDefMapping(eventName string, synthDefName string) {
	sc.synthDefMapping[eventName] = synthDefName
}

// SetGroupID sets the Group ID for a given event name
// For example: SetGroupID("kick", 100)
func (sc *SuperColliderAdapter) SetGroupID(eventName string, groupID int32) {
	sc.groupIDMapping[eventName] = groupID
}

// SetBusID sets the output bus ID for a given event name
// For example: SetBusID("snare", 10) to route snare to bus 10 (reverb)
// Default is 0 (master out) if not set
func (sc *SuperColliderAdapter) SetBusID(eventName string, busID int32) {
	sc.busIDMapping[eventName] = busID
}

// GetSynthDefName returns the SynthDef name for a given event name
func (sc *SuperColliderAdapter) GetSynthDefName(eventName string) string {
	if name, ok := sc.synthDefMapping[eventName]; ok {
		return name
	}
	// Default: use event name as synthdef name
	return eventName
}

// GetGroupID returns the Group ID for a given event name
func (sc *SuperColliderAdapter) GetGroupID(eventName string) int32 {
	if id, ok := sc.groupIDMapping[eventName]; ok {
		return id
	}
	// Default group ID if not mapped
	return 1 // default group
}

// GetBusID returns the output bus ID for a given event name
func (sc *SuperColliderAdapter) GetBusID(eventName string) int32 {
	if id, ok := sc.busIDMapping[eventName]; ok {
		return id
	}
	// Default: master out (bus 0)
	return 0
}

// GetHost returns the current host
func (sc *SuperColliderAdapter) GetHost() string {
	return sc.host
}

// GetPort returns the current port
func (sc *SuperColliderAdapter) GetPort() int {
	return sc.port
}

// GetAllSynthDefMappings returns all synthdef mappings
func (sc *SuperColliderAdapter) GetAllSynthDefMappings() map[string]string {
	result := make(map[string]string)
	for k, v := range sc.synthDefMapping {
		result[k] = v
	}
	return result
}

// GetAllGroupMappings returns all group ID mappings
func (sc *SuperColliderAdapter) GetAllGroupMappings() map[string]int32 {
	result := make(map[string]int32)
	for k, v := range sc.groupIDMapping {
		result[k] = v
	}
	return result
}

// GetAllBusMappings returns all bus ID mappings
func (sc *SuperColliderAdapter) GetAllBusMappings() map[string]int32 {
	result := make(map[string]int32)
	for k, v := range sc.busIDMapping {
		result[k] = v
	}
	return result
}

// Send implements EventAdapter.Send
// Sends server commands directly to scsynth using timestamped bundles
func (sc *SuperColliderAdapter) Send(scheduled events.ScheduledEvent) error {
	switch scheduled.Event.Type {
	case events.EventTypeNote:
		return sc.sendNote(scheduled)
	case events.EventTypeModulation:
		return sc.sendModulation(scheduled)
	case events.EventTypeRest:
		// Rest is a no-op
		return nil
	}
	return nil
}

// sendNote sends server commands for note events
// Creates a bundle with /g_freeAll and /s_new commands for monophonic retriggering
func (sc *SuperColliderAdapter) sendNote(scheduled events.ScheduledEvent) error {
	event := scheduled.Event
	timing := scheduled.Timing

	// Get synthdef name, group ID, and output bus
	synthDefName := sc.GetSynthDefName(event.Name)
	groupID := sc.GetGroupID(event.Name)
	outputBus := sc.GetBusID(event.Name)

	// Message 1: /g_freeAll - free all synths in the group (monophonic retrigger)
	freeAllMsg := osc.NewMessage("/g_freeAll")
	freeAllMsg.Append(groupID)

	// Message 2: /s_new - create new synth
	// Format: /s_new synthDefName nodeID addAction targetID [controls...]
	// nodeID: -1 (auto-generate)
	// addAction: 1 (add to tail of group)
	// targetID: groupID
	newSynthMsg := osc.NewMessage("/s_new")
	newSynthMsg.Append(synthDefName)                        // synthdef name
	newSynthMsg.Append(int32(-1))                           // nodeID (-1 = auto-generate)
	newSynthMsg.Append(int32(1))                            // addAction (1 = tail)
	newSynthMsg.Append(groupID)                             // target group ID
	newSynthMsg.Append("amp")                               // control name
	newSynthMsg.Append(event.B)                             // velocity (amp)
	newSynthMsg.Append("len")                               // control name
	newSynthMsg.Append(float32(timing.Duration.Seconds()))  // duration
	newSynthMsg.Append("out")                               // control name
	newSynthMsg.Append(outputBus)                           // output bus

	// Create bundle with both messages and timestamp
	bundle := osc.NewBundle(timing.Timestamp)
	bundle.Append(freeAllMsg)
	bundle.Append(newSynthMsg)

	// Send the bundle to scsynth
	err := sc.client.Send(bundle)
	if err != nil {
		return fmt.Errorf("failed to send SuperCollider note bundle: %w", err)
	}

	return nil
}

// sendModulation sends server commands for modulation/CC events
// Uses /n_set to modify synth parameters
func (sc *SuperColliderAdapter) sendModulation(scheduled events.ScheduledEvent) error {
	event := scheduled.Event
	timing := scheduled.Timing

	// For modulation, we'd need to track node IDs or use group-level controls
	// This is a placeholder - modulation support would need more infrastructure
	_ = event
	_ = timing

	// TODO: Implement modulation support if needed
	return nil
}

// Close closes the SuperCollider adapter (no-op for OSC)
func (sc *SuperColliderAdapter) Close() error {
	// OSC clients don't need cleanup
	return nil
}
