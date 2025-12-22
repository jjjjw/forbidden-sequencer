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
- Interface: `GetEventsForTick(tick int64) []TickEvent`, `Reset()`, `Play()`, `Stop()`
- Receives Conductor reference at construction time
- Called once per tick by sequencer via callback
- Returns `TickEvent` containing event data + `TickTiming` (tick number, offset, duration in ticks)
- When paused, returns nil (no events)
- When playing, returns tick events for the requested tick

**Conductor** - Tick-based master clock
- Core interface: `GetTickDuration()`, `GetLastTickTime()`, `GetCurrentTick()`, `SetTickCallback()`, `Start()`, `Reset()`
- Single source of truth for tick numbering
- Runs continuously once started, advancing ticks at precise intervals
- Invokes registered callback on each tick with current tick number
- Uses absolute wall-clock time for drift-free scheduling
- Patterns query conductor for timing information (read-only)

**Sequencer** - Pattern orchestration
- Manages pattern list and conductor lifecycle
- Registers `handleTick()` callback with conductor
- **Lookahead scheduling**: At tick N, generates and schedules events for tick N+10
- Converts tick-based events to wall-clock scheduled events
- Tracks `lastScheduledTime` to prevent overlapping events when tempo changes
- Delegates Play/Stop/Reset to patterns
- Sends scheduled events to adapter and TUI events channel

**Adapters** - Protocol output
- `MIDIAdapter`: Reads `midi_note` or `freq` from params, converts as needed, handles timing in goroutines
- `SuperColliderAdapter`: Sends OSC to scsynth, converts `midi_note` to `freq`, passes all params to SynthDef
- `OSCAdapter`: Generic OSC output, converts params to OSC message arguments

### System Flow

**Goroutines:**
1. **Conductor loop:** Advances ticks using absolute time scheduling, invokes tick callback
2. **Sequencer handleTick callback:** Called on each tick, queries patterns, converts to scheduled events, sends to adapter
3. **Adapter goroutines:** Handle protocol-specific timing (OSC bundles, MIDI note on/off)

**Key properties:**
- Conductor drives timing via callback, patterns respond to tick numbers
- Lookahead scheduling: events generated 10 ticks ahead for stable timing
- Patterns return nil when paused, tick events when playing
- Each pattern called exactly once per tick
- Wall-clock time tracking prevents overlapping events during tempo changes
- No shared mutable state between patterns (conductor is read-only)

### Concrete Example: Simple Kick Pattern

**Setup:**
```go
conductor := conductors.NewConductor(100 * time.Millisecond) // 10 ticks/second
kickPattern := modulated.NewSimpleKickPattern(conductor, "kick", 0.8, 4) // fires every 4 ticks
eventsChan := make(chan events.ScheduledEvent, 100)
sequencer := NewSequencer([]Pattern{kickPattern}, conductor, oscAdapter, eventsChan, false)
sequencer.Start()
sequencer.Play()
```

**Pattern logic:**
- Tracks tick position in phrase
- Returns `TickEvent` with tick number + timing when it's time to fire
- Sequencer receives tick callback, asks pattern for tick N+10
- Sequencer converts tick-relative timing to wall-clock scheduled event
- Adapter receives scheduled event with absolute timestamp

**Result:** Kick drum every 400ms (4 ticks × 100ms)

### Event Creation Examples

**Patterns return `TickEvent` with tick-based timing:**
```go
// Simple kick drum pattern
events.TickEvent{
    Event: events.Event{
        Name: "kick",
        Type: events.EventTypeNote,
        Params: map[string]float32{
            "midi_note": 36.0,  // C1 kick
            "amp":       0.8,   // 80% amplitude
        },
    },
    TickTiming: events.TickTiming{
        Tick:          tick,    // Which tick this event belongs to
        OffsetPercent: 0.0,     // Start of tick (0.0 to 1.0)
        DurationTicks: 0.5,     // Half a tick duration
    },
}
```

**FM synth with swing timing:**
```go
events.TickEvent{
    Event: events.Event{
        Name: "fm",
        Type: events.EventTypeNote,
        Params: map[string]float32{
            "midi_note": 48.0,    // C3
            "amp":       0.7,     // 70% amplitude
            "modRatio":  2.0,     // Modulator at 2x carrier freq
            "modIndex":  1.5,     // Modulation depth
        },
    },
    TickTiming: events.TickTiming{
        Tick:          tick,
        OffsetPercent: 0.33,    // Delayed 33% into the tick (swing feel)
        DurationTicks: 1.25,    // Longer than one tick
    },
}
```

**Sequencer converts to `ScheduledEvent` with absolute wall-clock timing:**
```go
events.ScheduledEvent{
    Event: tickEvent.Event,  // Same event data
    Timing: events.Timing{
        Timestamp: time.Time, // Absolute wall-clock time
        Duration:  time.Duration, // Absolute duration
    },
}
```

**Note:**
- Patterns work in tick-relative time (ticks + offsets)
- Sequencer converts to wall-clock time for adapters
- Adapters handle protocol-specific conversions (midi_note ↔ freq)

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
