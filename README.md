# Forbidden Sequencer

A modular, pattern-based MIDI sequencer with live performance controls.

## Architecture

### Core Concepts

**Events** - Protocol-agnostic musical data
- `Event`: name, type, a, b, c, d (float32 parameters)
- `ScheduledEvent`: Event + Timing (wait + duration)
- Example: frequency in Hz (not MIDI notes)

**Pattern** - Stateful event generator
- Interface: `GetNextScheduledEvent() (ScheduledEvent, error)`
- Receives Conductor reference at construction time
- Stores conductor internally for timing queries
- Returns ScheduledEvents with Delta calculated from conductor's tick state
- Each pattern runs in its own goroutine
- Maintains internal state (conductor, lastFireTick, etc.)

**Conductor** - Tick-based master clock
- Minimal interface: `GetCurrentTick()`, `GetTickDuration()`, `Start()`, `Pause()`, `Resume()`
- Runs in its own goroutine, advancing ticks at precise intervals
- Single source of truth for timing (prevents drift)
- `CommonTimeConductor` implementation adds:
  - Beat-awareness (ticksPerBeat, BPM)
  - Musical time helpers (IsBeatStart, GetTickInBeat, GetBeat)
  - Dynamic tempo changes via SetBPM()
- Patterns query conductor but cannot mutate it (read-only interface)

**Sequencer** - Pattern orchestration
- Manages pattern list and conductor lifecycle
- Starts conductor in its own goroutine
- Launches each pattern in its own goroutine
- Each pattern independently queries conductor and sends events to adapter
- Provides playback controls: Start(), Pause(), Resume()

**Controller** - Frontend ↔ Backend bridge
- Receives control changes from frontend
- Maps controls to pattern parameters
- One controller per sequencer

**Adapters** - Protocol output
- `MIDIAdapter`: Converts frequency → MIDI, handles timing in goroutines
- Future: OSC, etc.

### System Flow

**Three independent goroutines:**
1. **Conductor loop:** Advances ticks at precise intervals (tickDuration)
2. **Pattern loops (one per pattern):** Each pattern independently:
   - Queries conductor for current tick
   - Calculates next fire tick based on musical logic
   - Computes Delta as `(nextFireTick - currentTick) * tickDuration`
   - Sleeps for Delta
   - Sends ScheduledEvent to adapter
3. **Adapter goroutines:** Handle MIDI note on/off timing

**Key properties:**
- Patterns can drift from conductor or stay in sync (their choice)
- Conductor ticks continue during pause (patterns check pause state)
- No shared mutable state between patterns (conductor is read-only)
- Supports both realtime (sleep on Delta) and non-realtime (collect events)

### Concrete Example: Techno Sequencer

**Setup:**
```go
conductor := NewCommonTimeConductor(4, 120) // 4 ticks/beat, 120 BPM
kick := NewKickPattern(conductor)           // Fires every beat
hihat := NewHihatPattern(conductor)         // Fires every half-beat
sequencer := NewSequencer([]Pattern{kick, hihat}, conductor, midiAdapter, false)
sequencer.Start()
```

**Kick pattern logic:**
- Stores conductor reference
- `k.conductor.GetCurrentTick() % k.conductor.GetTicksPerBeat() == 0` → on beat boundary
- Fires MIDI note 36 (bass drum)

**Hihat pattern logic:**
- Stores conductor reference
- Fires on half-beat (ticksPerBeat/2)
- Fires MIDI note 42 (closed hihat)

**Result:** "boom tick boom tick" techno beat at 120 BPM

## Next Steps

- [x] Define Pattern interface
- [x] Implement Conductor with timing state
- [x] Refactor Sequencer to use Pattern + Conductor
- [x] Create basic patterns (Kick, Hihat)
- [x] Create Techno sequencer factory
- [ ] Add coordination primitives to Conductor (scratch space for pattern communication)
- [ ] Example: Kick Mutes Bass pattern coordination
  - Kick pattern writes to Conductor scratch space when it fires
  - Bass pattern checks Conductor for kicks on current beat
  - If kick present, bass returns empty (mute)
  - Result: Bass automatically mutes when kick plays, no direct coupling
- [ ] Build Controller for frontend parameter mapping
- [ ] Build pattern library (LogarithmicTiming, ClusterStructure, VelocityCurve)
- [ ] Frontend control rendering + keyboard bindings

## Development

### Live Development

Run `wails dev` in the project directory for hot reload.

Dev server: http://localhost:34115

### Building

Build production package: `wails build`

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
│       ├── kick_pattern.go   # Kick on every beat
│       └── hihat_pattern.go  # Hihat on off-beats
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
