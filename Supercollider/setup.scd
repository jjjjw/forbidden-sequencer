// Setup script for Forbidden Sequencer SuperCollider integration
// This script initializes Groups, SynthDefs, and audio buses
// Synths are triggered directly by server commands from Go (port 57110)
// Uses monophonic voices via groups that are freed and recreated on each event

// Boot server if not already running
s.waitForBoot({

  // Load SynthDefs first
  (thisProcess.nowExecutingPath.dirname +/+ "synthdefs.scd").load;

  s.sync; // Wait for SynthDefs to load

  // Create audio bus for reverb with fixed ID (or reuse existing)
  if (~reverbBus.isNil, {
    ~reverbBus = Bus.audio(s, 2, 10); // 2 channels, starting at bus 10
    "Created reverb bus: %".format(~reverbBus.index).postln;
  }, {
    "Reusing existing reverb bus: %".format(~reverbBus.index).postln;
  });

  // Free existing reverb synth if it exists
  if (~reverb.notNil, {
    ~reverb.free;
  });

  // Create reverb synth at the tail (processes after sound sources)
  ~reverb = Synth.tail(s, \allpassReverb, [
    \in, ~reverbBus,
    \out, 0,
    \size, 0.7,
    \decay, 0.6,
    \wet, 0.5,
    \hpass, 150,
    \lpass, 9000
  ]);

  "Reverb initialized on bus %".format(~reverbBus.index).postln;

  // Create monophonic voice groups with fixed IDs using server commands
  // These IDs must match the group mappings in Go adapter
  // /g_new [nodeID, addAction, targetID]
  // addAction: 1 = add to tail of target (default group 0)
  s.sendMsg('/g_new', 100, 1, 0); // kick group
  s.sendMsg('/g_new', 101, 1, 0); // snare group
  s.sendMsg('/g_new', 102, 1, 0); // hihat group

  "Groups created with fixed IDs:".postln;
  "  Kick group: 100".postln;
  "  Snare group: 101".postln;
  "  Hihat group: 102".postln;

  // Server is now ready to receive commands on port 57110
  // Go adapter will send timestamped bundles containing:
  // 1. /g_freeAll [groupID] - free all synths in group
  // 2. /s_new [synthDefName, -1, 1, groupID, controls...] - create new synth
  //
  // SynthDef mappings (set in Go):
  //   kick  -> bd (bass drum)
  //   snare -> cp (clap/snare)
  //   hihat -> hh (hihat)
  //
  // Group ID mappings (set in Go):
  //   kick  -> 100
  //   snare -> 101
  //   hihat -> 102
  //
  // Bus ID mappings (set in Go):
  //   kick  -> 0 (master out)
  //   snare -> 10 (reverb bus)
  //   hihat -> 0 (master out)

  "SuperCollider setup complete".postln;
  "Server listening on port 57110 for direct server commands".postln;
});
