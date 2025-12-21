package adapters

import (
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"forbidden_sequencer/sequencer/events"

	"github.com/hypebeast/go-osc/osc"
)

// NodeInfo tracks an active synth node
type NodeInfo struct {
	NodeID  int32
	EndTime time.Time
}

// SuperColliderAdapter implements EventAdapter for SuperCollider server commands
// Sends OSC messages directly to scsynth (port 57110) using server command protocol
type SuperColliderAdapter struct {
	client          *osc.Client
	host            string
	port            int
	synthDefMapping map[string]string     // maps event names to SynthDef names
	busIDMapping    map[string]int32      // maps event names to output bus IDs
	activeNodes     map[string][]NodeInfo // tracks active nodes per event name
	nextNodeID      int32                 // auto-incrementing node ID
	nodesMutex      sync.Mutex            // protects activeNodes and nextNodeID
	debug           bool                  // enable debug logging
	debugLog        *log.Logger           // debug logger for OSC messages
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
		client:          client,
		host:            host,
		port:            port,
		synthDefMapping: make(map[string]string),
		busIDMapping:    make(map[string]int32),
		activeNodes:     make(map[string][]NodeInfo),
		nextNodeID:      2000, // Start node IDs at 2000
		debug:           debug,
		debugLog:        debugLogger,
	}, nil
}

// SetSynthDefMapping sets the SynthDef name for a given event name
// For example: SetSynthDefMapping("kick", "bd")
func (sc *SuperColliderAdapter) SetSynthDefMapping(eventName string, synthDefName string) {
	sc.synthDefMapping[eventName] = synthDefName
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

// GetAllBusMappings returns all bus ID mappings
func (sc *SuperColliderAdapter) GetAllBusMappings() map[string]int32 {
	result := make(map[string]int32)
	for k, v := range sc.busIDMapping {
		result[k] = v
	}
	return result
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
	case events.EventTypeModulation:
		return sc.sendModulation(scheduled)
	case events.EventTypeRest:
		// Rest is a no-op
		return nil
	}
	return nil
}

// sendNote sends server commands for note events using node-based voice management
func (sc *SuperColliderAdapter) sendNote(scheduled events.ScheduledEvent) error {
	event := scheduled.Event
	timing := scheduled.Timing

	// Get synthdef name and output bus
	synthDefName := sc.GetSynthDefName(event.Name)
	outputBus := sc.GetBusID(event.Name)

	// Get max_voices from params (default to 1 if not specified)
	maxVoices := 1
	if maxVoicesParam, hasMaxVoices := event.Params["max_voices"]; hasMaxVoices {
		maxVoices = int(maxVoicesParam)
	}

	// Lock for node management
	sc.nodesMutex.Lock()

	// Clean up finished nodes (nodes that will be finished by the new event's timestamp)
	activeNodes := sc.activeNodes[event.Name]
	filteredNodes := make([]NodeInfo, 0, len(activeNodes))
	for _, node := range activeNodes {
		if timing.Timestamp.Before(node.EndTime) {
			filteredNodes = append(filteredNodes, node)
		}
	}
	activeNodes = filteredNodes

	// Prepare bundle for timestamped messages
	bundle := osc.NewBundle(timing.Timestamp)

	// If at voice limit, free oldest active voice
	if len(activeNodes) >= maxVoices {
		oldestNode := activeNodes[0]
		// Only free if it will still be playing when this event starts
		if timing.Timestamp.Before(oldestNode.EndTime) {
			freeMsg := osc.NewMessage("/n_free")
			freeMsg.Append(oldestNode.NodeID)
			bundle.Append(freeMsg)
		}
		activeNodes = activeNodes[1:]
	}

	// Assign new node ID
	nodeID := sc.nextNodeID
	sc.nextNodeID++

	// Track new node
	endTime := timing.Timestamp.Add(timing.Duration)
	activeNodes = append(activeNodes, NodeInfo{NodeID: nodeID, EndTime: endTime})
	sc.activeNodes[event.Name] = activeNodes

	sc.nodesMutex.Unlock()

	// Create /s_new message
	// Format: /s_new synthDefName nodeID addAction targetID [controls...]
	// addAction: 0 (add to head of group)
	// targetID: 100 (synths group - executes before effects group 200)
	newSynthMsg := osc.NewMessage("/s_new")
	newSynthMsg.Append(synthDefName) // synthdef name
	newSynthMsg.Append(nodeID)       // node ID
	newSynthMsg.Append(int32(0))     // addAction (0 = add to head)
	newSynthMsg.Append(int32(100))   // target (100 = synths group)

	// Add parameters from Params dict
	// Handle midi_note -> freq conversion if needed
	if midiNote, hasMidiNote := event.Params["midi_note"]; hasMidiNote {
		newSynthMsg.Append("freq")
		newSynthMsg.Append(midiToFreq(midiNote))
	} else if freq, hasFreq := event.Params["freq"]; hasFreq {
		newSynthMsg.Append("freq")
		newSynthMsg.Append(freq)
	}

	// Add all other parameters (except midi_note, freq, len, max_voices)
	for key, value := range event.Params {
		if key != "midi_note" && key != "freq" && key != "len" && key != "max_voices" {
			newSynthMsg.Append(key)
			newSynthMsg.Append(value)
		}
	}

	// Add len - use from Params if present, otherwise from Timing.Duration
	if lenParam, hasLen := event.Params["len"]; hasLen {
		newSynthMsg.Append("len")
		newSynthMsg.Append(lenParam)
	} else {
		newSynthMsg.Append("len")
		newSynthMsg.Append(float32(timing.Duration.Seconds()))
	}

	// Always add out
	newSynthMsg.Append("out")
	newSynthMsg.Append(outputBus)

	// Add to bundle
	bundle.Append(newSynthMsg)

	// Debug log
	if sc.debugLog != nil {
		sc.debugLog.Printf("%v", bundle)
	}

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
