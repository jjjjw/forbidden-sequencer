// Simple 4x4 FDN Reverb
// 4 delay lines with Hadamard mixing + 2 allpass filters

s.waitForBoot({

  // Load SynthDefs
  (thisProcess.nowExecutingPath.dirname +/+ "synthdefs.scd").load;

  s.sync;

  // Create reverb bus
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

  // Create reverb synth
  ~reverb = Synth.tail(s, \simpleFDN, [
    \in, ~reverbBus.index,
    \out, 0,
    \feedback, 0.85,
    \wet, 0.5,
    \hpass, 200,
    \lpass, 12000
  ]);

  "Simple FDN reverb initialized on bus %".format(~reverbBus.index).postln;

  // Define helper function: Recreate reverb with new parameters
  ~recreateReverb = {
    arg feedback=0.85, wet=0.5, hpass=200, lpass=12000;

    "--- Recreating reverb ---".postln;

    // Free existing reverb
    if (~reverb.notNil, {
      ~reverb.free;
      "Destroyed reverb synth".postln;
    });

    // Create new reverb with specified parameters
    ~reverb = Synth.tail(s, \simpleFDN, [
      \in, ~reverbBus.index,
      \out, 0,
      \feedback, feedback,
      \wet, wet,
      \hpass, hpass,
      \lpass, lpass
    ]);

    "Recreated reverb: feedback=%, wet=%, hpass=%, lpass=%".format(
      feedback, wet, hpass, lpass
    ).postln;
  };

  "Sandbox setup complete - use ~recreateReverb.value() to test".postln;
});

(
SynthDef(\simpleFDN, {
  arg in=0, out=0,
      feedback=0.85, hpass=200, lpass=12000, wet=0.5;

  var input, dry, snd, processed, wet_sig;
  var hadamard;

  // 4x4 Hadamard matrix (normalized by 1/sqrt(4) = 0.5)
  hadamard = [
    [ 1,  1,  1,  1],
    [ 1, -1,  1, -1],
    [ 1,  1, -1, -1],
    [ 1, -1, -1,  1]
  ] * 0.5;

  // Input (stereo)
  input = In.ar(in, 2);
  dry = input;

  // Expand stereo to 4 channels
  snd = input.wrapExtend(4);

  // FEEDBACK: 4 delays with Hadamard matrix mixing
  snd = snd + LocalIn.ar(4);

  snd = DelayC.ar(snd, 0.05, [0.0315, 0.0329, 0.0375, 0.0396]);

  snd = snd * feedback;

  // Filtering
  snd = HPF.ar(snd, hpass);
  snd = LPF.ar(snd, lpass);

  // Apply Hadamard matrix for feedback mixing
  snd = snd * hadamard.flop;
  snd = snd.sum;

  LocalOut.ar(snd);

  // Output as stereo with Splay
  processed = Splay.ar(snd);

  // 2 Allpass filters after the feedback loop
  processed = AllpassC.ar(processed, 0.1, 0.0051, 0.1);
  processed = AllpassC.ar(processed, 0.1, 0.0127, 0.1);

  // Mix dry and wet
  wet_sig = (dry * (1 - wet)) + (processed * wet);

  Out.ar(out, wet_sig);
}).add;

~recreateReverb.(feedback: 0.8);
)

// Trigger hi-hat (routed to reverb bus 10)
Synth(\hh, [\len, 0.2, \out, ~reverbBus.index, \amp, 0.8]);
