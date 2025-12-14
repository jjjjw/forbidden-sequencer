package adapters

import (
	"fmt"
	"log"
	"math"
	"os"

	"forbidden_sequencer/sequencer/events"

	"github.com/hypebeast/go-osc/osc"
)

// ParameterMapping maps event parameters (A, B, C, D) to synth control names
type ParameterMapping struct {
	A string // maps Event.A to this control name
	B string // maps Event.B to this control name
	C string // maps Event.C to this control name
	D string // maps Event.D to this control name
}

// SuperColliderAdapter implements EventAdapter for SuperCollider server commands
// Sends OSC messages directly to scsynth (port 57110) using server command protocol
type SuperColliderAdapter struct {
	client           *osc.Client
	host             string
	port             int
	synthDefMapping  map[string]string           // maps event names to SynthDef names
	groupIDMapping   map[string]int32            // maps event names to Group IDs
	busIDMapping     map[string]int32            // maps event names to output bus IDs
	parameterMapping map[string]ParameterMapping // maps event names to parameter mappings
	debug            bool                        // enable debug logging
	debugLog         *log.Logger                 // debug logger for OSC messages
}

// NewSuperColliderAdapter creates a new SuperCollider adapter
// host: target host (e.g., "localhost")
// port: scsynth port (default: 57110)
// debug: enable debug logging to debug/sc_adapter_osc.log
func NewSuperColliderAdapter(host string, port int, debug bool) (*SuperColliderAdapter, error) {
	client := osc.NewClient(host, port)

	var debugLogger *log.Logger
	if debug {
		// Create debug log file
		debugFile, err := os.Create("debug/sc_adapter_osc.log")
		if err != nil {
			return nil, fmt.Errorf("failed to create debug log: %w", err)
		}
		debugLogger = log.New(debugFile, "", log.LstdFlags|log.Lmicroseconds)
	}

	return &SuperColliderAdapter{
		client:           client,
		host:             host,
		port:             port,
		synthDefMapping:  make(map[string]string),
		groupIDMapping:   make(map[string]int32),
		busIDMapping:     make(map[string]int32),
		parameterMapping: make(map[string]ParameterMapping),
		debug:            debug,
		debugLog:         debugLogger,
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

// SetParameterMapping sets the parameter mapping for a given event name
// For example: SetParameterMapping("kick", ParameterMapping{A: "freq", B: "amp", C: "ratio", D: "sweep"})
func (sc *SuperColliderAdapter) SetParameterMapping(eventName string, mapping ParameterMapping) {
	sc.parameterMapping[eventName] = mapping
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

// GetAllParameterMappings returns all parameter mappings
func (sc *SuperColliderAdapter) GetAllParameterMappings() map[string]ParameterMapping {
	result := make(map[string]ParameterMapping)
	for k, v := range sc.parameterMapping {
		result[k] = v
	}
	return result
}

// GetParameterMapping returns the parameter mapping for a given event name
func (sc *SuperColliderAdapter) GetParameterMapping(eventName string) ParameterMapping {
	if mapping, ok := sc.parameterMapping[eventName]; ok {
		return mapping
	}
	// Default mapping
	return ParameterMapping{A: "freq", B: "amp", C: "", D: ""}
}

// midiToFreq converts MIDI note number to frequency in Hz
func midiToFreq(midiNote float32) float32 {
	return float32(440.0 * math.Pow(2.0, (float64(midiNote)-69.0)/12.0))
}

// Send implements EventAdapter.Send
// Sends server commands directly to scsynth using timestamped bundles
func (sc *SuperColliderAdapter) Send(scheduled events.ScheduledEvent) error {
	switch scheduled.Event.Type {
	case events.EventTypeNote:
		return sc.sendNote(scheduled)
	case events.EventTypeFrequency:
		return sc.sendFrequency(scheduled)
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

	// Get synthdef name, group ID, output bus, and parameter mapping
	synthDefName := sc.GetSynthDefName(event.Name)
	groupID := sc.GetGroupID(event.Name)
	outputBus := sc.GetBusID(event.Name)
	paramMapping := sc.GetParameterMapping(event.Name)

	// Message 1: /g_freeAll - free all synths in the group (monophonic retrigger)
	freeAllMsg := osc.NewMessage("/g_freeAll")
	freeAllMsg.Append(groupID)

	// Message 2: /s_new - create new synth
	// Format: /s_new synthDefName nodeID addAction targetID [controls...]
	// nodeID: -1 (auto-generate)
	// addAction: 1 (add to tail of group)
	// targetID: groupID
	newSynthMsg := osc.NewMessage("/s_new")
	newSynthMsg.Append(synthDefName)  // synthdef name
	newSynthMsg.Append(int32(-1))     // nodeID (-1 = auto-generate)
	newSynthMsg.Append(int32(1))      // addAction (1 = tail)
	newSynthMsg.Append(groupID)       // target group ID

	// Add parameters based on mapping
	if paramMapping.A != "" {
		newSynthMsg.Append(paramMapping.A)
		// If parameter is "freq", convert MIDI note to frequency
		if paramMapping.A == "freq" {
			newSynthMsg.Append(midiToFreq(event.A))
		} else {
			newSynthMsg.Append(event.A)
		}
	}
	if paramMapping.B != "" {
		newSynthMsg.Append(paramMapping.B)
		newSynthMsg.Append(event.B)
	}
	if paramMapping.C != "" {
		newSynthMsg.Append(paramMapping.C)
		newSynthMsg.Append(event.C)
	}
	if paramMapping.D != "" {
		newSynthMsg.Append(paramMapping.D)
		newSynthMsg.Append(event.D)
	}

	// Always add len and out
	newSynthMsg.Append("len")
	newSynthMsg.Append(float32(timing.Duration.Seconds()))
	newSynthMsg.Append("out")
	newSynthMsg.Append(outputBus)

	// Debug log the message
	if sc.debugLog != nil {
		sc.debugLog.Printf("Event: %s -> SynthDef: %s, Group: %d, Bus: %d", event.Name, synthDefName, groupID, outputBus)
		sc.debugLog.Printf("  Param mapping: A=%s, B=%s, C=%s, D=%s", paramMapping.A, paramMapping.B, paramMapping.C, paramMapping.D)
		if paramMapping.A == "freq" {
			sc.debugLog.Printf("  Param values: A=%v (MIDI) -> %v Hz, B=%v, C=%v, D=%v", event.A, midiToFreq(event.A), event.B, event.C, event.D)
		} else {
			sc.debugLog.Printf("  Param values: A=%v, B=%v, C=%v, D=%v", event.A, event.B, event.C, event.D)
		}
		sc.debugLog.Printf("  len=%v, out=%v", timing.Duration.Seconds(), outputBus)
	}

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

// sendFrequency sends server commands for frequency events (Event.A is already in Hz)
// Creates a bundle with /g_freeAll and /s_new commands for monophonic retriggering
func (sc *SuperColliderAdapter) sendFrequency(scheduled events.ScheduledEvent) error {
	event := scheduled.Event
	timing := scheduled.Timing

	// Get synthdef name, group ID, output bus, and parameter mapping
	synthDefName := sc.GetSynthDefName(event.Name)
	groupID := sc.GetGroupID(event.Name)
	outputBus := sc.GetBusID(event.Name)
	paramMapping := sc.GetParameterMapping(event.Name)

	// Message 1: /g_freeAll - free all synths in the group (monophonic retrigger)
	freeAllMsg := osc.NewMessage("/g_freeAll")
	freeAllMsg.Append(groupID)

	// Message 2: /s_new - create new synth
	newSynthMsg := osc.NewMessage("/s_new")
	newSynthMsg.Append(synthDefName)  // synthdef name
	newSynthMsg.Append(int32(-1))     // nodeID (-1 = auto-generate)
	newSynthMsg.Append(int32(1))      // addAction (1 = tail)
	newSynthMsg.Append(groupID)       // target group ID

	// Add parameters based on mapping
	// For EventTypeFrequency, Event.A is already in Hz, no conversion needed
	if paramMapping.A != "" {
		newSynthMsg.Append(paramMapping.A)
		newSynthMsg.Append(event.A) // already in Hz
	}
	if paramMapping.B != "" {
		newSynthMsg.Append(paramMapping.B)
		newSynthMsg.Append(event.B)
	}
	if paramMapping.C != "" {
		newSynthMsg.Append(paramMapping.C)
		newSynthMsg.Append(event.C)
	}
	if paramMapping.D != "" {
		newSynthMsg.Append(paramMapping.D)
		newSynthMsg.Append(event.D)
	}

	// Always add len and out
	newSynthMsg.Append("len")
	newSynthMsg.Append(float32(timing.Duration.Seconds()))
	newSynthMsg.Append("out")
	newSynthMsg.Append(outputBus)

	// Debug log the message
	if sc.debugLog != nil {
		sc.debugLog.Printf("Event: %s (frequency) -> SynthDef: %s, Group: %d, Bus: %d", event.Name, synthDefName, groupID, outputBus)
		sc.debugLog.Printf("  Param mapping: A=%s, B=%s, C=%s, D=%s", paramMapping.A, paramMapping.B, paramMapping.C, paramMapping.D)
		sc.debugLog.Printf("  Param values: A=%v Hz, B=%v, C=%v, D=%v", event.A, event.B, event.C, event.D)
		sc.debugLog.Printf("  len=%v, out=%v", timing.Duration.Seconds(), outputBus)
	}

	// Create bundle with both messages and timestamp
	bundle := osc.NewBundle(timing.Timestamp)
	bundle.Append(freeAllMsg)
	bundle.Append(newSynthMsg)

	// Send the bundle to scsynth
	err := sc.client.Send(bundle)
	if err != nil {
		return fmt.Errorf("failed to send SuperCollider frequency bundle: %w", err)
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
