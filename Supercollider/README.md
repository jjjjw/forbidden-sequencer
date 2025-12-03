# SuperCollider Integration

This folder contains SuperCollider SynthDefs and setup scripts for the Forbidden Sequencer.

## Overview

The sequencer sends events to SuperCollider via **OSC** (Open Sound Control):

- **kick** → `/trigger/kick` with precise timestamps
- **snare** → `/trigger/snare` with precise timestamps
- **hihat** → `/trigger/hihat` with precise timestamps
- Uses OSC bundles with timestamps for sample-accurate scheduling
- Sends frequency, velocity, duration, and additional parameters

Each voice is **monophonic** and **retriggers** on every event, preventing overlapping instances.

## Setup

### 1. Start SuperCollider
1. Open SuperCollider IDE
2. Boot the audio server: `s.boot;`
3. Load and run the OSC setup script:

```supercollider
// Navigate to the Supercollider folder and run:
(thisProcess.nowExecutingPath.dirname +/+ "setup_osc.scd").load;
```

Or execute the setup script directly from the IDE:
```supercollider
// Load OSC setup (which also loads synthdefs)
"<path-to-forbidden_sequencer>/Supercollider/setup_osc.scd".load;
```

### 2. Start the Sequencer
1. Run the forbidden_sequencer:
   ```bash
   go run .
   ```
2. Press `space` or `p` to start playback

SuperCollider will listen on port **57121** for OSC bundles with timestamps.

## Files

- **synthdefs.scd** - SynthDef definitions for bd, cp, hh, and reverb
- **setup_osc.scd** - OSC initialization with timestamp-based scheduling
- **README.md** - This file

## SynthDefs

### `\bd` (Bass Drum)
- Pitched sine wave with frequency sweep
- Parameters: `freq` (50 Hz default), `len`, `amp`, `ratio`, `sweep`

### `\cp` (Clap)
- Filtered noise with randomized envelope for natural clap sound
- Parameters: `freq` (unused), `len`, `amp`

### `\hh` (Hi-Hat)
- Bandpass filtered noise burst
- Parameters: `freq` (unused), `len`, `amp`

## Customization

### Modify Synth Parameters
You can pass additional parameters when creating synths in `setup_osc.scd`:
```supercollider
Synth(\bd, [\amp, amp, \len, len, \freq, 60], ~kickGroup);
```

### Change OSC Address Mappings
Edit the OSC address paths in the `OSCdef` declarations in `setup_osc.scd`, and update the corresponding mappings in the Go code (`main.go`).
