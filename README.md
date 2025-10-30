# Forbidden Sequencer

A modular, pattern-based MIDI sequencer with live performance controls.

## Architecture

### Core Concepts

**Events** - Protocol-agnostic musical data
- `Event`: name, type, a, b, c, d (float32 parameters)
- `ScheduledEvent`: Event + Timing (wait + duration)
- Example: frequency in Hz (not MIDI notes)

**Patterns** - Composable building blocks that emit ScheduledEvents
- Defined in Go code
- Can be composed, chained, and nested
- Reusable components (logarithmic timing, clusters, velocity curves, etc.)
- Built on top of EventGenerators + TimingGenerators

**Tracks** - Performance-ready compositions
- Collection of Patterns wired together
- Defines exposed controls for live manipulation
- Track = composition + performance interface specification

**Adapters** - Protocol output
- `MIDIAdapter`: Converts frequency → MIDI, handles timing in goroutines
- Future: OSC, etc.

> **Note**: The current Sequencer implementation (lanes, EventGenerators, TimingGenerators) is the foundation that will evolve into the Patterns + Tracks system.

### Concrete Example: "Kick + Synth Cluster" Pattern

**What it does:**
- Triggers kick and synth together via MIDI
- Events arranged in logarithmic timing clusters
- Cluster structure: N events close together, then gap, repeat

**Track Controls** (3-4 sliders):
1. **Cluster spacing** - gap between clusters
2. **Events per cluster** - density
3. **Synth velocity curve** - envelope across cluster
4. **Overall intensity** - global velocity scaling

**Pattern Library Components** (reusable):
- `LogarithmicTiming` - timing distribution
- `ClusterStructure` - N events + spacing pattern
- `VelocityCurve` - envelope across cluster
- Compose these to build specific patterns

**Frontend**:
- Renders sliders based on Track's control definitions
- Keyboard bindings for controls (like MIDI CC mapping)
- Live parameter tweaking during performance

### System Flow

1. **Define Patterns in Go** - Build from reusable components
2. **Define Track in Go** - Wire patterns + expose controls
3. **Load Track** - Backend sends control schema to frontend
4. **Frontend renders controls** - Sliders, knobs based on Track definition
5. **User performs** - Adjust controls live, optional keyboard bindings
6. **Backend modulates** - Pattern parameters respond to control changes
7. **Patterns emit ScheduledEvents** - Continuous event stream
8. **Adapters output** - MIDI, OSC, etc.

## Next Steps

- [ ] Define Pattern interface (builds on EventGenerator + TimingGenerator)
- [ ] Define Track structure
- [ ] Build pattern library (LogarithmicTiming, ClusterStructure, VelocityCurve)
- [ ] Implement pattern composition system
- [ ] Control → Pattern parameter mapping
- [ ] Frontend track selector + control rendering
- [ ] Keyboard binding system

## Development

### Live Development

Run `wails dev` in the project directory for hot reload.

Dev server: http://localhost:34115

### Building

Build production package: `wails build`

## Project Structure

```
sequencer/
├── event.go              # Event, Timing, ScheduledEvent types
├── markov.go            # Markov chain engine
├── sequencer.go         # Foundation (will evolve into Patterns + Tracks)
├── adapters/
│   ├── adapter.go       # EventAdapter interface
│   └── midi_adapter.go  # MIDI implementation
├── event_generators/
│   ├── event_generator.go
│   └── markov_generator.go
└── timing_generators/
    ├── timing_generator.go
    └── fixed_rate.go
```
