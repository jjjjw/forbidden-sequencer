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
	active         map[string]noteState // track active notes for note-off
	channelMapping map[string]uint8     // maps event names to MIDI channels
}

type noteState struct {
	midiNote uint8
	channel  uint8
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
		active:         make(map[string]noteState),
		channelMapping: make(map[string]uint8),
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

// Send implements EventAdapter.Send
// Note: Fires immediately - caller is responsible for timing (Delta)
func (m *MIDIAdapter) Send(scheduled events.ScheduledEvent) error {
	switch scheduled.Event.Type {
	case events.EventTypeNote:
		return m.sendNote(scheduled)
	case events.EventTypeModulation:
		return m.sendCC(scheduled)
	case events.EventTypeRest:
		// Rest is a no-op
		return nil
	}
	return nil
}

// sendNote converts frequency to MIDI note and sends note on/off
func (m *MIDIAdapter) sendNote(scheduled events.ScheduledEvent) error {
	event := scheduled.Event
	timing := scheduled.Timing
	// Convert frequency to MIDI note number
	midiNote := frequencyToMIDI(event.A)

	// Convert normalized velocity (0.0-1.0) to MIDI velocity (0-127)
	velocity := uint8(event.B * 127.0)

	// Get channel from mapping
	channel := m.GetChannelMapping(event.Name)

	// Send note on
	if err := m.send(midi.NoteOn(channel, midiNote, velocity)); err != nil {
		return fmt.Errorf("failed to send note on: %w", err)
	}

	// Track active note
	m.active[event.Name] = noteState{midiNote: midiNote, channel: channel}

	// Schedule note off after duration
	if timing.Duration > 0 {
		go func() {
			time.Sleep(timing.Duration)
			m.send(midi.NoteOff(channel, midiNote))
			delete(m.active, event.Name)
		}()
	}

	return nil
}

// sendCC sends MIDI CC message
func (m *MIDIAdapter) sendCC(scheduled events.ScheduledEvent) error {
	event := scheduled.Event
	// a = CC number, b = value (0.0-1.0)
	ccNum := uint8(event.A)
	ccValue := uint8(event.B * 127.0)

	// Get channel from mapping
	channel := m.GetChannelMapping(event.Name)

	if err := m.send(midi.ControlChange(channel, ccNum, ccValue)); err != nil {
		return fmt.Errorf("failed to send CC: %w", err)
	}

	return nil
}

// Close closes the MIDI output
func (m *MIDIAdapter) Close() error {
	// Send note off for all active notes
	for name, state := range m.active {
		m.send(midi.NoteOff(state.channel, state.midiNote))
		delete(m.active, name)
	}

	if m.out != nil {
		return m.out.Close()
	}
	return nil
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
