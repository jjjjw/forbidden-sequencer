# SuperCollider Integration

This folder contains SuperCollider SynthDefs and setup scripts for the Forbidden Sequencer.

## Overview

The sequencer sends events to SuperCollider via **direct server commands** (OSC protocol):

- Go → **scsynth** (port 57110) using `/n_free` and `/s_new` commands
- **Server-side scheduling** with timestamped OSC bundles
- **Node-based voice management**: intelligent per-event polyphony control via `max_voices` parameter

### Event Flow

```
Go Sequencer
    ↓ (timestamped OSC bundle)
scsynth (port 57110)
    ↓ (server scheduler)
/n_free [nodeID]      → free oldest voice if max_voices exceeded (voice stealing)
/s_new [synthDef...]  → create new synth with unique node ID
```

### Mappings

| Event | SynthDef | Max Voices | Bus |
|-------|----------|------------|-----|
| kick  | bd       | 1 (default)| 0 (master out) |
| snare | cp       | 1 (default)| 10 (reverb) |
| hihat | hh       | 1 (default)| 0 (master out) |
| fm    | fm2op    | 2          | 10 (reverb) |
| arp   | arp      | 1 (default)| 10 (reverb) |

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

- **synthdefs.scd** - SynthDef definitions for bd, cp, hh, fm2op, arp, and fdnReverb
- **setup.scd** - Initialize audio buses and reverb
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

### `\fm2op` (2-Operator FM Synth)
- Two-operator frequency modulation synthesis
- Modulator frequency is a ratio of the carrier frequency
- Parameters: `freq`, `amp`, `len`, `out`, `modRatio` (0.5-7.0), `modIndex` (0.1-3.0)

### `\arp` (Arpeggiator)
- Pulse wave through resonant lowpass filter with envelope modulation
- Classic arpeggiator sound with percussive filter sweep
- Parameters: `freq`, `amp`, `len`, `out`, `cutoff` (2000 Hz default), `res` (0.5 default)

### `\fdnReverb` (FDN Reverb Effect)
- Feedback Delay Network reverb with Hadamard matrix diffusion
- Routes from bus 10 to master out (bus 0)
- Parameters: `in`, `out`, `size`, `feedback`, `wet`, `hpass`, `lpass`, `earlyMix`

## Architecture Details

### Voice Management

The Go adapter tracks active synth nodes per event name and implements intelligent voice stealing:

- Each event has a configurable `max_voices` parameter (defaults to 1)
- Adapter assigns unique node IDs (starting at 1001) and tracks their end times
- When a new event would exceed `max_voices`:
  - The oldest active node is freed via `/n_free` command
  - Only nodes still playing at the new event's timestamp are freed
- Active nodes are automatically cleaned up when their duration expires

### Buses
- **Bus 0** - Master stereo out (default)
- **Bus 10** - Reverb input (2 channels)

The reverb synth (node ID 1000) processes audio from bus 10 and outputs to bus 0.

### Node Tree Execution Order

SuperCollider executes nodes in the order they appear in the node tree. For the reverb to process audio from bus 10, voice synths **must execute before** effects.

The setup creates two groups with fixed IDs:
- **Group 100** - Synths group (executes first)
- **Group 200** - Effects group (executes second)

**Node tree structure:**
```
Group 0 (RootNode)
  ├── Group 100 (synths) - voice synths added here
  │     ├── 2000+ (voice nodes) - execute first, write to bus 0 or 10
  └── Group 200 (effects) - execute after synths
        └── 1000 (fdnReverb) - reads from bus 10, writes to bus 0
```

Run `s.queryAllNodes` in SuperCollider to verify the node tree structure.

### Server Commands

Each note event sends a timestamped bundle containing:

1. **`/n_free nodeID`** (optional) - Frees oldest voice if max_voices exceeded
2. **`/s_new synthDefName nodeID 0 100 "param1" val1 "param2" val2...`** - Creates new synth
   - nodeID: unique ID assigned by adapter (2000+)
   - addAction: 0 (add to head of group)
   - targetID: 100 (synths group)
   - All event parameters passed as control pairs

## Customization

### Modify Synth Parameters
Edit the SynthDefs in `synthdefs.scd` to add new controls or change synthesis.

### Change Bus Routing
Edit both:
1. `supercollider/setup.scd` - Bus numbers and reverb configuration
2. `sequencer/adapters/setup.go` - Go-side bus mappings via `SetBusID()`

### Add New Sounds
1. Add SynthDef to `synthdefs.scd`
2. Add SynthDef and bus mappings in `sequencer/adapters/setup.go`:
   ```go
   scAdapter.SetSynthDefMapping("newSound", "newSynthDef")
   scAdapter.SetBusID("newSound", 10) // route to reverb, or 0 for dry
   ```
3. Create pattern that emits events with name "newSound"

### Configure Polyphony
Set `max_voices` parameter in event patterns:
```go
Event{
    Name: "fm",
    Type: events.EventTypeNote,
    Params: map[string]float32{
        "midi_note":  60,
        "max_voices": 4,  // allow up to 4 simultaneous voices
    },
}
```
