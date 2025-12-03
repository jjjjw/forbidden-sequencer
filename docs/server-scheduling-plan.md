# Server Scheduling Architecture Plan

## Current Architecture (Language-side scheduling)

**Flow:**
```
Go → sclang (port 57121) via OSCdef → SystemClock.sched → synth creation
```

**Issues:**
- Language-side processing adds latency
- Causes "LATE" warnings at fast sequencer rates
- SystemClock.sched runs on sclang, not scsynth
- Extra layer of scheduling between OSC receipt and synth creation

## Proposed Architecture (Server-side scheduling)

**Flow:**
```
Go → scsynth (port 57110) via /s_new in timestamped bundles → server scheduler → synth creation
```

**Advantages:**
- Sample-accurate scheduling at server level
- No language-side overhead
- Eliminates "LATE" warnings
- Standard SuperCollider server communication pattern
- Server handles all timing internally

## Implementation Using .makeBundle

Instead of sending raw `/s_new` and `/g_freeAll` commands from Go, we use SuperCollider's `.makeBundle()` method to generate properly formatted server bundles from sclang.

### .makeBundle() Method

```supercollider
s.makeBundle(time, func, bundle)
```

**Parameters:**
- `time`: Delay in seconds (nil/number = auto-send, false = don't send)
- `func`: Function that generates OSC messages
- `bundle`: Optional pre-existing bundle to add to

**How it works:**
1. Function is evaluated
2. All OSC messages generated are captured
3. Messages are bundled with timestamp
4. Bundle is sent to server (if time ≠ false)

### Monophonic Percussion Handling

For monophonic retriggering behavior, each event needs:
1. Free all synths in the voice's group
2. Create new synth in that group

Both operations must happen at the same timestamp.

## Changes Required

### 1. SuperCollider setup (`Supercollider/setup_osc.scd`)

**Remove:**
- All OSCdef responders (kick, snare, hihat)
- SystemClock.sched calls
- Late logging (schedTime checks)

**Keep:**
- SynthDef loading
- Group creation (with fixed IDs):
  - `~kickGroup` → ID 100
  - `~snareGroup` → ID 101
  - `~hihatGroup` → ID 102
- Reverb bus and synth setup

**Add:**
- OSCdef that uses `.makeBundle()` to schedule on server:

```supercollider
OSCdef(\kick, { arg msg, time, addr, recvPort;
  var freq, vel, dur, c, d, amp, len, schedTime;

  freq = msg[1];
  vel = msg[2];
  dur = msg[3];
  c = msg[4];
  d = msg[5];

  amp = vel;
  len = dur;

  // Calculate delay from now until scheduled time
  schedTime = time - Main.elapsedTime;

  // Use makeBundle to schedule on server (not SystemClock)
  s.makeBundle(schedTime, {
    ~kickGroup.freeAll;
    Synth(\bd, [\freq, 50, \amp, amp, \len, len], ~kickGroup);
  });
}, '/trigger/kick');
```

### 2. Go OSC adapter (`sequencer/adapters/setup.go`)

**No changes needed** - continue sending to port 57121 (sclang)

### 3. Go OSC adapter (`sequencer/adapters/osc_adapter.go`)

**No changes needed** - continue sending same OSC message format with bundles

## Key Differences from Previous Plan

| Aspect | Previous Plan (Raw /s_new) | Current Plan (.makeBundle) |
|--------|---------------------------|---------------------------|
| Target | scsynth (port 57110) | sclang (port 57121) |
| Message Format | `/s_new`, `/g_freeAll` | `/trigger/*` (current) |
| Scheduling | Server internal | Server via .makeBundle |
| Monophonic Logic | Manual group IDs | Use existing Group objects |
| Go Changes | Extensive refactoring | None required |
| SC Changes | Remove OSCdef, raw commands | Small OSCdef modification |

## Benefits of .makeBundle Approach

1. **Minimal code changes** - only modify SuperCollider OSCdef handlers
2. **Keep existing API** - Go continues sending same messages
3. **Server-side scheduling** - bundles sent to server, not SystemClock
4. **Cleaner code** - use Group objects instead of raw IDs
5. **Easier to maintain** - familiar SuperCollider patterns

## Implementation Steps

1. Update `Supercollider/setup_osc.scd`:
   - Replace `SystemClock.sched()` with `s.makeBundle()` in all OSCdef handlers
   - Remove late logging (schedTime < 0 checks)

2. Test at various sequencer rates to verify no "LATE" warnings

3. Update `Supercollider/README.md` to document server-side scheduling architecture

## Expected Outcome

- No "LATE" warnings, even at fast sequencer rates
- Sample-accurate timing via server scheduler
- Cleaner SuperCollider code
- No changes required in Go codebase
