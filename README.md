# Forbidden Sequencer

A modular, pattern-based MIDI sequencer with live performance controls.

## Architecture

### Core Concepts

**Events** - Protocol-agnostic musical data
- `Event`: name, type, params (dictionary of string → float32)
  - Common params: `freq`, `midi_note`, `amp`, `len`
  - Synth-specific params: `modRatio`, `modIndex`, etc.
- `ScheduledEvent`: Event + Timing (timestamp + duration)
- Adapters handle conversions (e.g., midi_note → freq for SuperCollider)

**Pattern** - Tick-driven event generator
- Interface: `GetScheduledEventsForTick(nextTickTime, tickDuration)`, `Reset()`, `Play()`, `Stop()`
- Receives Conductor reference at construction time
- Called once per tick by sequencer
- **Always schedules ahead**: patterns schedule events from `nextTickTime` forward, never for "now"
- Queries conductor for next tick state (`GetNextTickInBeat()`, `GetNextTickInPhrase()`)
- When paused, returns nil (no events)
- When playing, returns events for the next tick period

**Conductor** - Tick-based master clock
- Minimal interface: `GetTickDuration()`, `Start()`, `Ticks()`
- Runs continuously once started, advancing ticks at precise intervals
- Emits on `Ticks()` channel when each tick fires
- Uses absolute wall-clock time for drift-free scheduling
- `CommonTimeConductor` adds beat-awareness (ticksPerBeat, BPM)
  - `GetNextTickInBeat()` returns the next tick position (0 to ticksPerBeat-1)
  - `GetNextTickTime()` returns absolute time of next tick
- `PhraseConductor` adds phrase tracking (phraseLength, variable rate)
  - `GetNextTickInPhrase()` returns the next tick position in phrase
- Patterns query conductor but cannot mutate it (read-only)

**Sequencer** - Pattern orchestration
- Manages pattern list and conductor lifecycle
- Listens for conductor ticks and calls patterns on each tick
- Schedules returned events using `time.AfterFunc`
- Delegates Play/Stop/Reset to patterns
- Events channel for sending scheduled events to TUI

**Adapters** - Protocol output
- `MIDIAdapter`: Reads `midi_note` or `freq` from params, converts as needed, handles timing in goroutines
- `SuperColliderAdapter`: Sends OSC to scsynth, converts `midi_note` to `freq`, passes all params to SynthDef
- `OSCAdapter`: Generic OSC output, converts params to OSC message arguments

### System Flow

**Goroutines:**
1. **Conductor loop:** Advances ticks using absolute time scheduling, emits on Ticks channel
2. **Sequencer tick loop:** Listens for ticks, calls patterns, schedules returned events via AfterFunc
3. **Adapter goroutines:** Handle MIDI note on/off timing

**Key properties:**
- Conductor drives timing, patterns respond to ticks
- Patterns always schedule ahead (never for "now")
- Patterns return nil when paused, events when playing
- Each pattern called exactly once per tick
- No shared mutable state between patterns (conductor is read-only)

### Concrete Example: Techno Sequencer

**Setup:**
```go
conductor := NewCommonTimeConductor(120) // 120 BPM
pattern := NewTechnoPattern(conductor)   // Alternates kick and hihat
sequencer := NewSequencer([]Pattern{pattern}, conductor, midiAdapter, false)
sequencer.Start()
sequencer.Play()
```

**Techno pattern logic:**
- Checks `GetNextTickInBeat()` - if next tick is 0 (beat boundary), schedule events
- Returns both kick (at nextTickTime) and hihat (at nextTickTime + halfBeat)
- No internal state needed - pattern is a pure function of conductor state

**Result:** "boom tick boom tick" techno beat at 120 BPM

### Event Creation Examples

**Simple kick drum with frequency:**
```go
events.Event{
    Name: "kick",
    Type: events.EventTypeNote,
    Params: map[string]float32{
        "freq": 60.0,  // 60 Hz kick
        "amp":  0.8,   // 80% amplitude
    },
}
```

**Arpeggio note with MIDI note number:**
```go
events.Event{
    Name: "arp",
    Type: events.EventTypeNote,
    Params: map[string]float32{
        "midi_note": 60.0,  // Middle C
        "amp":       0.9,   // 90% amplitude
    },
}
```

**FM synth with custom parameters:**
```go
events.Event{
    Name: "fm1",
    Type: events.EventTypeNote,
    Params: map[string]float32{
        "midi_note": 48.0,    // C3
        "amp":       0.7,     // 70% amplitude
        "modRatio":  2.0,     // Modulator at 2x carrier freq
        "modIndex":  1.5,     // Modulation depth
    },
}
```

**Note:** Adapters automatically handle conversions:
- SuperCollider: `midi_note` → `freq` conversion
- MIDI: `freq` → `midi_note` conversion
- Params are passed directly to synths/instruments

## Next Steps

- [x] Define Pattern interface
- [x] Implement Conductor with timing state
- [x] Refactor Sequencer to use Pattern + Conductor
- [x] Create basic patterns (Kick, Hihat)
- [x] Create Techno sequencer factory
- [x] Implement drift-free absolute time scheduling
- [x] Implement Play/Stop/Reset controls
- [ ] Build pattern library (LogarithmicTiming, ClusterStructure, VelocityCurve)

## Development

### Running

```bash
go run .
```

Or build and run:
```bash
go build -o forbidden-sequencer .
./forbidden-sequencer
```

### Controls

**Main Screen:**
- `space/p` - Play/Stop sequencer
- `r` - Reset to beginning
- `s` - Settings
- `q` - Quit

**Settings Screen:**
- `j/k` or arrows - Navigate
- `enter` - Select option
- `esc` - Back

**MIDI Port / Channel Mapping:**
- `j/k` or arrows - Navigate
- `enter` - Select/Edit
- `esc` - Back to Settings

### Requirements

- Go 1.24+
- A MIDI output port (virtual or physical)
  - On macOS: Enable IAC Driver in Audio MIDI Setup > Window > Show MIDI Studio

### Testing

Run all tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test ./... -v
```

Run tests for a specific package:
```bash
go test ./sequencer/conductors -v
go test ./sequencer/patterns/techno -v
```
