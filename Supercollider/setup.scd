// Setup script for Forbidden Sequencer SuperCollider integration
// This script initializes Groups, SynthDefs, and audio buses
// Synths are triggered directly by server commands from Go (port 57110)
// Uses node-based voice management with synths group (100) and effects group (200)

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

  // Create groups with fixed IDs for proper execution order
  // Group 100: synths (executes first)
  // Group 200: effects (executes second, after synths)
  s.sendMsg('/g_new', 100, 0, 0); // synths group, add to head of default group
  s.sendMsg('/g_new', 200, 3, 100); // effects group, add after synths group

  s.sync;

  "Created synths group: 100".postln;
  "Created effects group: 200".postln;

  // Create reverb synth in the effects group (200)
  // The reverb processes audio from bus 10 and outputs to master (bus 0)
  s.sendMsg('/s_new', 'fdnReverb', 1000, 1, 200,
    'in', ~reverbBus.index,
    'out', 0,
    'size', 0.7,
    'feedback', 0.9,
    'wet', 0.5,
    'hpass', 200,
    'lpass', 16000
  );

  s.sync; // Wait for reverb to be created

  "Reverb initialized on bus % with node ID 1000".format(~reverbBus.index).postln;

  // Server is now ready to receive commands on port 57110
  // Go adapter will send timestamped bundles containing:
  // 1. /n_free [nodeID] - free specific node (voice stealing)
  // 2. /s_new [synthDefName, nodeID, 0, 100, controls...] - create new synth
  //
  // Node tree structure:
  //   Group 0 (RootNode)
  //     ├── Group 100 (synths) - voice synths added here, executes first
  //     └── Group 200 (effects) - reverb and other effects, executes second
  //         └── Node 1000 (fdnReverb) - processes bus 10 → bus 0
  //
  // SynthDef mappings (set in Go):
  //   kick  -> bd (bass drum)
  //   snare -> cp (clap/snare)
  //   hihat -> hh (hihat)
  //   fm    -> fm2op (2-op FM synth, max_voices=2)
  //   arp   -> arp (pulse through lowpass)
  //
  // Bus ID mappings (set in Go):
  //   kick  -> 0 (master out)
  //   snare -> 10 (reverb bus)
  //   hihat -> 0 (master out)
  //   fm    -> 10 (reverb bus)
  //   arp   -> 10 (reverb bus)
  //
  // Voice management:
  //   - Adapter assigns unique node IDs (2000+) and tracks active nodes per event
  //   - All voice synths added to group 100 with addAction 0 (head)
  //   - max_voices parameter (defaults to 1) limits polyphony per event
  //   - Voice stealing: oldest active node freed when limit exceeded
  //

  "SuperCollider setup complete".postln;
  "Server listening on port 57110 for direct server commands".postln;
});
