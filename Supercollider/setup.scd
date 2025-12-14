// Setup script for Forbidden Sequencer SuperCollider integration
// This script initializes Groups, SynthDefs, and audio buses
// Synths are triggered directly by server commands from Go (port 57110)
// Uses monophonic voices via groups that are freed and recreated on each event

// Boot server if not already running
s.waitForBoot({

  // Load SynthDefs first
  (thisProcess.nowExecutingPath.dirname +/+ "synthdefs.scd").load;

  s.sync; // Wait for SynthDefs to load

  // Free existing reverb bus if it exists, then create at fixed ID 10
  if (~reverbBus.notNil, {
    ~reverbBus.free;
  });
  ~reverbBus = Bus(\audio, 10, 2, s); // fixed bus index 10, 2 channels
  "Created reverb bus: %".format(~reverbBus.index).postln;

  // Free existing reverb synth if it exists
  if (~reverb.notNil, {
    ~reverb.free;
  });

  // Create reverb synth FIRST at the tail with a fixed node ID
  // Using /s_new with fixed ID so we can reference it for ordering
  s.sendMsg('/s_new', 'fdnReverb', 1000, 1, 0,
    'in', ~reverbBus.index,
    'out', 0,
    'size', 0.7,
    'feedback', 0.9,
    'wet', 0.5,
    'hpass', 200,
    'lpass', 16000
  );

  s.sync; // Wait for reverb to be created

  // Create monophonic voice groups with fixed IDs using server commands
  // These IDs must match the group mappings in Go adapter
  // /g_new [nodeID, addAction, targetID]
  // addAction: 2 = add before target (place before reverb synth)
  // This ensures all voice groups execute BEFORE the reverb
  s.sendMsg('/g_new', 100, 2, 1000); // kick group - before reverb
  s.sendMsg('/g_new', 101, 2, 1000); // snare group - before reverb
  s.sendMsg('/g_new', 102, 2, 1000); // hihat group - before reverb
  s.sendMsg('/g_new', 103, 2, 1000); // fm1 group - before reverb
  s.sendMsg('/g_new', 104, 2, 1000); // fm2 group - before reverb

  "Reverb initialized on bus %".format(~reverbBus.index).postln;

  "Groups created with fixed IDs:".postln;
  "  Kick group: 100".postln;
  "  Snare group: 101".postln;
  "  Hihat group: 102".postln;
  "  FM1 group: 103".postln;
  "  FM2 group: 104".postln;

  // Server is now ready to receive commands on port 57110
  // Go adapter will send timestamped bundles containing:
  // 1. /g_freeAll [groupID] - free all synths in group
  // 2. /s_new [synthDefName, -1, 1, groupID, controls...] - create new synth
  //
  // SynthDef mappings (set in Go):
  //   kick  -> bd (bass drum)
  //   snare -> cp (clap/snare)
  //   hihat -> hh (hihat)
  //   fm1   -> fm2op (2-op FM synth)
  //   fm2   -> fm2op (2-op FM synth)
  //
  // Group ID mappings (set in Go):
  //   kick  -> 100
  //   snare -> 101
  //   hihat -> 102
  //   fm1   -> 103
  //   fm2   -> 104
  //
  // Bus ID mappings (set in Go):
  //   kick  -> 0 (master out)
  //   snare -> 10 (reverb bus)
  //   hihat -> 0 (master out)
  //   fm1   -> 10 (reverb bus)
  //   fm2   -> 10 (reverb bus)

  "SuperCollider setup complete".postln;
  "Server listening on port 57110 for direct server commands".postln;
});
