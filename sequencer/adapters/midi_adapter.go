package adapters

import (
	"fmt"
	"math"
	"time"

	"forbidden_sequencer/sequencer"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/portmididrv"
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
// portIndex: -1 for default port, or specific port index
func NewMIDIAdapter(portIndex int) (*MIDIAdapter, error) {
	drv, err := portmididrv.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize portmidi driver: %w", err)
	}

	var out drivers.Out
	if portIndex < 0 {
		// Use default output
		out, err = midi.OutPort(drv, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to open default MIDI port: %w", err)
		}
	} else {
		// Use specific port
		out, err = midi.OutPort(drv, portIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to open MIDI port %d: %w", portIndex, err)
		}
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
func (m *MIDIAdapter) Send(event sequencer.Event) error {
	switch event.Type {
	case sequencer.EventTypeNote:
		return m.sendNote(event)
	case sequencer.EventTypeModulation:
		return m.sendCC(event)
	case sequencer.EventTypeRest:
		// Rest is a no-op
		return nil
	default:
		return fmt.Errorf("unsupported event type: %s", event.Type)
	}
}

// sendNote converts frequency to MIDI note and sends note on/off
func (m *MIDIAdapter) sendNote(event sequencer.Event) error {
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
	if event.Duration > 0 {
		go func() {
			time.Sleep(event.Duration)
			m.send(midi.NoteOff(channel, midiNote))
			delete(m.active, event.Name)
		}()
	}

	return nil
}

// sendCC sends MIDI CC message
func (m *MIDIAdapter) sendCC(event sequencer.Event) error {
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
