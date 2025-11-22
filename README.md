# Forbidden Sequencer

A modular, pattern-based MIDI sequencer with live performance controls.

## Architecture

### Core Concepts

**Events** - Protocol-agnostic musical data
- `Event`: name, type, a, b, c, d (float32 parameters)
- `ScheduledEvent`: Event + Timing (wait + duration)
- Example: frequency in Hz (not MIDI notes)

**Pattern** - Stateful event generator
- Interface: `GetNextScheduledEvent()`, `Reset()`, `Play()`, `Stop()`
- Receives Conductor reference at construction time
- Returns ScheduledEvents with Delta calculated from conductor's tick state
- Maintains internal state (isKick, lastBeatTick, paused)
- When paused, returns short rests (10ms)
- When playing, calculates next event timing from conductor

**Conductor** - Tick-based master clock
- Minimal interface: `GetCurrentTick()`, `GetTickDuration()`, `Start()`, `GetBeatsChannel()`
- Runs continuously once started, advancing ticks at precise intervals
- Uses absolute wall-clock time for drift-free scheduling
- `CommonTimeConductor` adds beat-awareness (ticksPerBeat, BPM)
- `GetNextBeatTick()` returns next beat boundary tick
- `GetAbsoluteTimeForTick()` converts tick to wall-clock time
- Beats channel sends beat events to TUI
- Patterns query conductor but cannot mutate it (read-only)

**Sequencer** - Pattern orchestration
- Manages pattern list and conductor lifecycle
- Starts conductor and schedules pattern events
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
1. **Conductor loop:** Advances ticks using absolute time scheduling
2. **Pattern scheduling:** Sequencer schedules each pattern's events via AfterFunc
3. **Adapter goroutines:** Handle MIDI note on/off timing

**Key properties:**
- Conductor runs continuously, patterns handle their own pause state
- Patterns return rests when paused, events when playing
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
- Alternates kick and hihat using internal state
- Uses `GetNextBeatTick()` to schedule kicks on beat boundaries
- Schedules hihats half a beat after each kick

**Result:** "boom tick boom tick" techno beat at 120 BPM

## Next Steps

- [x] Define Pattern interface
- [x] Implement Conductor with timing state
- [x] Refactor Sequencer to use Pattern + Conductor
- [x] Create basic patterns (Kick, Hihat)
- [x] Create Techno sequencer factory
- [x] Implement drift-free absolute time scheduling
- [x] Implement Play/Stop/Reset controls
- [ ] Add coordination primitives to Conductor (scratch space for pattern communication)
- [ ] Example: Kick Mutes Bass pattern coordination
  - Kick pattern writes to Conductor scratch space when it fires
  - Bass pattern checks Conductor for kicks on current beat
  - If kick present, bass returns empty (mute)
  - Result: Bass automatically mutes when kick plays, no direct coupling
- [ ] Build Controller for frontend parameter mapping
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
