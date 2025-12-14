package adapters

import (
	"fmt"
	"math"
	"time"

	"forbidden_sequencer/sequencer/events"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

// MIDIAdapter implements EventAdapter for MIDI output
type MIDIAdapter struct {
	out            drivers.Out
	send           func(msg midi.Message) error
	channelMapping map[string]uint8 // maps event names to MIDI channels
	driver         *rtmididrv.Driver
	currentPort    int
}

// MIDIPortInfo represents information about a MIDI output port
type MIDIPortInfo struct {
	Index int
	Name  string
}

// NewMIDIAdapter creates a new MIDI adapter
// portIndex: -1 for default port (0), or specific port index
func NewMIDIAdapter(portIndex int) (*MIDIAdapter, error) {
	drv, err := rtmididrv.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize rtmidi driver: %w", err)
	}

	// Use port 0 if -1 specified
	port := portIndex
	if portIndex < 0 {
		port = 0
	}

	// Get list of output ports
	outs, err := drv.Outs()
	if err != nil {
		return nil, fmt.Errorf("failed to get MIDI outputs: %w", err)
	}

	if len(outs) == 0 {
		return nil, fmt.Errorf("no MIDI output ports available")
	}

	if port >= len(outs) {
		return nil, fmt.Errorf("MIDI port %d not found (only %d ports available)", port, len(outs))
	}

	out := outs[port]
	err = out.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open MIDI port %d: %w", port, err)
	}

	send, err := midi.SendTo(out)
	if err != nil {
		return nil, fmt.Errorf("failed to create MIDI sender: %w", err)
	}

	return &MIDIAdapter{
		out:            out,
		send:           send,
		channelMapping: make(map[string]uint8),
		driver:         drv,
		currentPort:    port,
	}, nil
}

// SetChannelMapping sets the MIDI channel for a given event name
func (m *MIDIAdapter) SetChannelMapping(eventName string, channel uint8) {
	m.channelMapping[eventName] = channel
}

// GetChannelMapping returns the MIDI channel for a given event name
// Returns 0 (default channel) if not mapped
func (m *MIDIAdapter) GetChannelMapping(eventName string) uint8 {
	if channel, ok := m.channelMapping[eventName]; ok {
		return channel
	}
	return 0 // default to channel 0
}

// GetAllChannelMappings returns all channel mappings
func (m *MIDIAdapter) GetAllChannelMappings() map[string]uint8 {
	result := make(map[string]uint8)
	for k, v := range m.channelMapping {
		result[k] = v
	}
	return result
}

// GetCurrentPort returns the index of the currently selected MIDI port
func (m *MIDIAdapter) GetCurrentPort() int {
	return m.currentPort
}

// ListAvailablePorts returns a list of all available MIDI output ports
func (m *MIDIAdapter) ListAvailablePorts() ([]MIDIPortInfo, error) {
	if m.driver == nil {
		return nil, fmt.Errorf("MIDI driver not initialized")
	}

	outs, err := m.driver.Outs()
	if err != nil {
		return nil, fmt.Errorf("failed to get MIDI outputs: %w", err)
	}

	ports := make([]MIDIPortInfo, len(outs))
	for i, out := range outs {
		ports[i] = MIDIPortInfo{
			Index: i,
			Name:  out.String(),
		}
	}

	return ports, nil
}

// SetPort changes the MIDI output port
func (m *MIDIAdapter) SetPort(portIndex int) error {
	// Send NoteOff to all channels in use
	m.allNotesOff()

	// Close current port
	if m.out != nil {
		m.out.Close()
	}

	// Get list of output ports
	outs, err := m.driver.Outs()
	if err != nil {
		return fmt.Errorf("failed to get MIDI outputs: %w", err)
	}

	if portIndex < 0 || portIndex >= len(outs) {
		return fmt.Errorf("MIDI port %d not found (only %d ports available)", portIndex, len(outs))
	}

	// Open new port
	out := outs[portIndex]
	err = out.Open()
	if err != nil {
		return fmt.Errorf("failed to open MIDI port %d: %w", portIndex, err)
	}

	// Create new sender
	send, err := midi.SendTo(out)
	if err != nil {
		out.Close()
		return fmt.Errorf("failed to create MIDI sender: %w", err)
	}

	m.out = out
	m.send = send
	m.currentPort = portIndex

	return nil
}

// Send implements EventAdapter.Send
// Schedules MIDI messages to be sent at the event's timestamp
func (m *MIDIAdapter) Send(scheduled events.ScheduledEvent) error {
	// Schedule the event to fire at its timestamp
	time.AfterFunc(time.Until(scheduled.Timing.Timestamp), func() {
		switch scheduled.Event.Type {
		case events.EventTypeNote:
			m.sendNote(scheduled)
		case events.EventTypeModulation:
			m.sendCC(scheduled)
		case events.EventTypeRest:
			// Rest is a no-op
		}
	})
	return nil
}

// sendNote sends MIDI note on/off
func (m *MIDIAdapter) sendNote(scheduled events.ScheduledEvent) error {
	event := scheduled.Event
	timing := scheduled.Timing

	// Get MIDI note - prefer midi_note, but convert from freq if needed
	var midiNote uint8
	if midiNoteParam, hasMidiNote := event.Params["midi_note"]; hasMidiNote {
		midiNote = uint8(midiNoteParam)
	} else if freq, hasFreq := event.Params["freq"]; hasFreq {
		midiNote = frequencyToMIDI(freq)
	} else {
		// Default to middle C if no note specified
		midiNote = 60
	}

	// Get velocity/amplitude - default to 0.8 if not specified
	amp := float32(0.8)
	if ampParam, hasAmp := event.Params["amp"]; hasAmp {
		amp = ampParam
	}

	// Convert normalized amplitude (0.0-1.0) to MIDI velocity (0-127)
	velocity := uint8(amp * 127.0)

	// Get channel from mapping
	channel := m.GetChannelMapping(event.Name)

	// Send note on
	err := m.send(midi.NoteOn(channel, midiNote, velocity))
	if err != nil {
		return fmt.Errorf("failed to send note on: %w", err)
	}

	// Schedule note off after duration
	if timing.Duration > 0 {
		time.AfterFunc(timing.Duration, func() {
			m.send(midi.NoteOff(channel, midiNote))
		})
	}

	return nil
}

// sendCC sends MIDI CC message
func (m *MIDIAdapter) sendCC(scheduled events.ScheduledEvent) error {
	event := scheduled.Event

	// Get CC number and value from params
	ccNum := uint8(event.Params["cc_num"])
	ccValue := uint8(event.Params["cc_value"] * 127.0)

	// Get channel from mapping
	channel := m.GetChannelMapping(event.Name)

	err := m.send(midi.ControlChange(channel, ccNum, ccValue))
	if err != nil {
		return fmt.Errorf("failed to send CC: %w", err)
	}

	return nil
}

// Close closes the MIDI output
func (m *MIDIAdapter) Close() error {
	m.allNotesOff()

	if m.out != nil {
		return m.out.Close()
	}
	return nil
}

// allNotesOff sends NoteOff for all notes on all channels in use
func (m *MIDIAdapter) allNotesOff() {
	if m.send == nil {
		return
	}
	// Send NoteOff for all 128 notes on each channel in use
	for _, channel := range m.channelMapping {
		for note := uint8(0); note < 128; note++ {
			m.send(midi.NoteOff(channel, note))
		}
	}
}

// frequencyToMIDI converts frequency in Hz to MIDI note number
// Uses formula: n = 12 * log2(f / 440) + 69
func frequencyToMIDI(freq float32) uint8 {
	if freq <= 0 {
		return 0
	}

	note := 12.0*math.Log2(float64(freq)/440.0) + 69.0

	// Clamp to valid MIDI range (0-127)
	if note < 0 {
		return 0
	}
	if note > 127 {
		return 127
	}

	return uint8(math.Round(note))
}
