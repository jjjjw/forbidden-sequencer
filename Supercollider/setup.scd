// Setup script for Forbidden Sequencer SuperCollider integration
// This script initializes Groups, SynthDefs, and audio buses

// Boot server if not already running
s.waitForBoot({

  // Load SynthDefs first
  (thisProcess.nowExecutingPath.dirname +/+ "lib/synthdefs.scd").load;

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
  "SuperCollider setup complete".postln;
  // "Server listening on port 57110 for direct server commands".postln;
});
