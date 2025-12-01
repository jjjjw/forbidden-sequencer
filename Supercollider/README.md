# SuperCollider Integration

This folder contains SuperCollider SynthDefs and setup scripts for the Forbidden Sequencer.

## Overview

The sequencer sends MIDI events to SuperCollider, which triggers audio synthesis using custom SynthDefs:
- **kick** (MIDI note 36) → `\bd` (bass drum)
- **snare** (MIDI note 37) → `\cp` (clap)
- **hihat** (MIDI note 42) → `\hh` (hi-hat)

Each voice is **monophonic** and **retriggers** on every MIDI event, preventing overlapping instances.

## Setup

### 1. Configure MIDI Virtual Port

#### macOS
1. Open **Audio MIDI Setup** (Applications → Utilities)
2. Window → Show MIDI Studio
3. Double-click **IAC Driver**
4. Check "Device is online"
5. Note the port name (e.g., "IAC Driver Bus 1")

### 2. Start SuperCollider

1. Open SuperCollider IDE
2. Boot the audio server: `s.boot;`
3. Load and run the setup script:

```supercollider
// Navigate to the Supercollider folder and run:
(thisProcess.nowExecutingPath.dirname +/+ "setup.scd").load;
```

Or execute the setup script directly from the IDE:
```supercollider
// Load setup (which also loads synthdefs)
"<path-to-forbidden_sequencer>/Supercollider/setup.scd".load;
```

### 3. Start the Sequencer

1. Run the forbidden_sequencer:
   ```bash
   go run .
   ```
2. Press `s` for Settings
3. Select "MIDI Port" and choose the same virtual port (e.g., IAC Driver)
4. Press `esc` to return to main screen
5. Press `space` or `p` to start playback

## Files

- **synthdefs.scd** - SynthDef definitions for bd, cp, and hh
- **setup.scd** - MIDI initialization and monophonic voice management
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

### Adjust Note Durations
Edit the `len` parameter in `setup.scd`:
```supercollider
var len = 0.5; // Change this value for longer/shorter notes
```

### Modify Synth Parameters
You can pass additional parameters when creating synths:
```supercollider
Synth(\bd, [\amp, amp, \len, len, \freq, 60], ~kickGroup);
```

### Change MIDI Note Mappings
Edit the `noteNum` parameter in the MIDIdef declarations in `setup.scd`.
