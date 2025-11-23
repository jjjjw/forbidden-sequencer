# Forbidden Sequencer

A modular, pattern-based MIDI sequencer with live performance controls.

## Architecture

### Core Concepts

**Events** - Protocol-agnostic musical data
- `Event`: name, type, a, b, c, d (float32 parameters)
- `ScheduledEvent`: Event + Timing (wait + duration)
- Example: frequency in Hz (not MIDI notes)

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
- `ModulatedConductor` adds phrase tracking (phraseLength, variable rate)
  - `GetNextTickInPhrase()` returns the next tick position in phrase
- Patterns query conductor but cannot mutate it (read-only)

**Sequencer** - Pattern orchestration
- Manages pattern list and conductor lifecycle
- Listens for conductor ticks and calls patterns on each tick
- Schedules returned events using `time.AfterFunc`
- Delegates Play/Stop/Reset to patterns
- Events channel for sending scheduled events to TUI

**Controller** - Frontend ↔ Backend bridge
- Receives control changes from frontend
- Maps controls to pattern parameters
- One controller per sequencer

**Adapters** - Protocol output
- `MIDIAdapter`: Converts frequency → MIDI, handles timing in goroutines
- Future: OSC, etc.

### System Flow

**Goroutines:**
1. **Conductor loop:** Advances ticks using absolute time scheduling, emits on Ticks channel
2. **Sequencer tick loop:** Listens for ticks, calls patterns, schedules returned events via AfterFunc
3. **Adapter goroutines:** Handle MIDI note on/off timing

**Key properties:**
- Conductor drives timing, patterns respond to ticks
- Patterns always schedule ahead (never for "now")
- Patterns return nil when paused, events when playing
- No deduplication needed - each pattern called exactly once per tick
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

## Project Structure

```
sequencer/
├── events/
│   └── event.go              # Event, Timing, ScheduledEvent types
├── conductors/
│   ├── conductor.go          # Conductor interface (minimal tick-based clock)
│   └── common_time_conductor.go # Beat-aware implementation
├── patterns/
│   └── techno/
│       └── techno_pattern.go # Alternating kick and hihat
├── sequencers/
│   ├── sequencer.go          # Pattern interface + orchestration
│   └── techno.go             # Techno sequencer factory
├── adapters/
│   ├── adapter.go            # EventAdapter interface
│   └── midi_adapter.go       # MIDI implementation
├── lib/
│   └── markov.go             # Markov chain engine
├── event_generators/         # Currently unused
│   ├── event_generator.go
│   └── markov_generator.go
└── timing_generators/        # Currently unused
    ├── timing_generator.go
    └── fixed_rate.go
```
