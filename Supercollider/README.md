# SuperCollider Integration

This folder contains SuperCollider SynthDefs, pattern implementations, and setup scripts for the Forbidden Sequencer.

## Overview

The Forbidden Sequencer uses **SuperCollider patterns** for generative music creation, controlled via **OSC messages** from a **Go TUI**.

### System Architecture

```
Go TUI (Terminal UI)
    ↓ (OSC control messages)
SuperCollider sclang (port 57120)
    ↓ (pattern control, synth triggering)
SuperCollider scsynth (audio server)
    ↓ (audio output)
Audio Interface
```

**Control Flow:**
- Go TUI sends OSC control messages to **sclang** (port 57120)
- Patterns run in sclang and respond to OSC commands
- Patterns trigger synths on **scsynth** server (internal communication)
- scsynth generates audio output

## Setup

### 1. Start SuperCollider

Open SuperCollider IDE and run:

```supercollider
// Boot the audio server
s.boot;

// Load the setup (synthdefs + effects)
"<path-to>/forbidden_sequencer/supercollider/setup.scd".load;

// Load patterns (choose one or more)
"<path-to>/forbidden_sequencer/supercollider/patterns/curve_time.scd".load;
"<path-to>/forbidden_sequencer/supercollider/patterns/markov_trig.scd".load;
"<path-to>/forbidden_sequencer/supercollider/patterns/markov_chord.scd".load;
```

SuperCollider will listen for OSC messages on port **57120** (default sclang port).

### 2. Start the Go TUI

```bash
cd forbidden_sequencer
go run .
```

The TUI will connect to sclang on `localhost:57120` and send OSC control messages to the loaded patterns.


## Architecture Details

### Audio Buses
- **Bus 0** - Master stereo out (default)
- **Bus 10** - Reverb input (2 channels)

The reverb synth (node ID 1000) processes audio from bus 10 and outputs to bus 0.

### Node Tree Execution Order

SuperCollider executes nodes in the order they appear in the node tree. For the reverb to process audio from bus 10, synths **must execute before** effects.

The setup creates two groups with fixed IDs:
- **Group 100** - Synths group (executes first)
- **Group 200** - Effects group (executes second)

**Node tree structure:**
```
Group 0 (RootNode)
  ├── Group 100 (synths) - pattern-triggered synths added here
  │     ├── 0+ (voice nodes) - execute first, write to bus 0 or 10
  └── Group 200 (effects) - execute after synths
        └── 1000 (fdnReverb) - reads from bus 10, writes to bus 0
```

Run `s.queryAllNodes` in SuperCollider to verify the node tree structure.

### Pattern Implementation

Patterns are implemented as **Tasks** running on **SystemClock**:

```supercollider
~myPattern.mainTask = Task({
    inf.do {
        // Pattern logic here
        // Trigger synths with s.bind
        s.bind {
            Synth(\bd, [\amp, 0.8, \len, 0.1], target: 100);
        };

        // Yield to wait between events
        0.125.yield;
    };
}, SystemClock);
```

### OSC Responders

Patterns define OSCdefs to respond to control messages:

```supercollider
OSCdef(\myPatternPlay, {
    ~myPattern.mainTask.reset;
    ~myPattern.mainTask.start;
}, '/pattern/my_pattern/play');

OSCdef(\myPatternPause, {
    ~myPattern.mainTask.pause;
}, '/pattern/my_pattern/pause');
```


## See Also

- [Main README](../README.md) - Overall system architecture and setup
- [Markov Chain Library](lib/markov.scd) - Markov chain implementation
- [Distribution Library](lib/synthdef.scd) - Synthdefs
