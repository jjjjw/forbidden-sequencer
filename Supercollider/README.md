# SuperCollider Integration

This folder contains SuperCollider SynthDefs and setup scripts for the Forbidden Sequencer.

## Overview

The sequencer sends events to SuperCollider via **direct server commands** (OSC protocol):

- Go → **scsynth** (port 57110) using `/g_freeAll` and `/s_new` commands
- **Server-side scheduling** with timestamped OSC bundles for sample-accurate timing
- **Monophonic voices** via Groups: each event frees the group then creates new synth
- No language-side (sclang) scheduling overhead

### Event Flow

```
Go Sequencer
    ↓ (timestamped OSC bundle)
scsynth (port 57110)
    ↓ (server scheduler)
/g_freeAll [groupID]  → free old synth
/s_new [synthDef...]  → create new synth
```

### Mappings

| Event | SynthDef | Group ID | Bus |
|-------|----------|----------|-----|
| kick  | bd       | 100      | 0 (master out) |
| snare | cp       | 101      | 10 (reverb) |
| hihat | hh       | 102      | 0 (master out) |

## Setup

### 1. Start SuperCollider
1. Open SuperCollider IDE
2. Boot the audio server: `s.boot;`
3. Load and run the setup script:

```supercollider
// Navigate to the Supercollider folder and run:
"setup.scd".load;
```

Or execute the setup script directly from the IDE:
```supercollider
// Load setup (which also loads synthdefs)
"<path-to-forbidden_sequencer>/Supercollider/setup.scd".load;
```

### 2. Start the Sequencer
1. Run the forbidden_sequencer:
   ```bash
   go run .
   ```
2. Press `space` or `p` to start playback

SuperCollider server will listen on port **57110** (default scsynth port) for direct server commands.

## Files

- **synthdefs.scd** - SynthDef definitions for bd, cp, hh, and reverb
- **setup.scd** - Initialize Groups, audio buses, and reverb
- **README.md** - This file

## SynthDefs

### `\bd` (Bass Drum)
- Pitched sine wave with frequency sweep
- Parameters: `freq` (50 Hz default), `len`, `amp`, `out`, `ratio`, `sweep`

### `\cp` (Clap)
- Filtered noise with randomized envelope for natural clap sound
- Parameters: `len`, `amp`, `out`

### `\hh` (Hi-Hat)
- Bandpass filtered noise burst
- Parameters: `len`, `amp`, `out`

### `\allpassReverb` (Reverb Effect)
- Allpass reverb with modulated delay times
- Routes from bus 10 to master out (bus 0)
- Parameters: `size`, `decay`, `wet`, `hpass`, `lpass`

## Architecture Details

### Groups
Groups are created with fixed IDs in `setup.scd`:
- `~kickGroup` = 100
- `~snareGroup` = 101
- `~hihatGroup` = 102

These IDs must match the mappings in Go (`sequencer/adapters/setup.go`).

### Buses
- **Bus 0** - Master stereo out (default)
- **Bus 10** - Reverb input (2 channels)

The reverb synth processes audio from bus 10 and outputs to bus 0.

### Server Commands

Each note event sends a timestamped bundle with two commands:

1. **`/g_freeAll groupID`** - Frees all synths in the group (monophonic retrigger)
2. **`/s_new synthDefName -1 1 groupID "amp" vel "len" dur "out" bus`** - Creates new synth
   - nodeID: -1 (auto-generate)
   - addAction: 1 (add to tail of group)
   - targetID: groupID

## Customization

### Modify Synth Parameters
Edit the SynthDefs in `synthdefs.scd` to add new controls or change synthesis.

### Change Mappings
Edit both:
1. `Supercollider/setup.scd` - Group IDs and bus numbers
2. `sequencer/adapters/setup.go` - Go-side mappings to match

### Add New Sounds
1. Add SynthDef to `synthdefs.scd`
2. Create Group in `setup.scd` with unique ID
3. Add mappings in `sequencer/adapters/setup.go`
