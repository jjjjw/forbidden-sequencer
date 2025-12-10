// Sandbox for testing reverb
// Evaluate this file to set up reverb
// Reverb inspiration: https://signalsmith-audio.co.uk/writing/2021/lets-write-a-reverb/
// TODO: Tune the delay times more
// TODO: Invert polarity in diffusor

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
  ~reverb = Synth.tail(s, \fdnReverb, [
    \in, ~reverbBus.index,
    \out, 0,
    \size, 0.5,
    \feedback, 0.9,
    \wet, 0.5,
    \hpass, 200,
    \lpass, 16000
  ]);

  "Reverb initialized on bus %".format(~reverbBus.index).postln;

  // Define helper function: Recreate reverb with new parameters
  ~recreateReverb = {
    arg size=0.5, feedback=0.9, wet=0.5, hpass=200, lpass=16000;

    "--- Recreating reverb ---".postln;

    // Free existing reverb
    if (~reverb.notNil, {
      ~reverb.free;
      "Destroyed reverb synth".postln;
    });

    // Create new reverb with specified parameters
    ~reverb = Synth.tail(s, \fdnReverb, [
      \in, ~reverbBus.index,
      \out, 0,
      \size, size,
      \feedback, feedback,
      \wet, wet,
      \hpass, hpass,
      \lpass, lpass
    ]);

    "Recreated reverb: size=%, feedback=%, wet=%, hpass=%, lpass=%".format(
      size, feedback, wet, hpass, lpass
    ).postln;
  };

  "Sandbox setup complete - use ~recreateReverb.value() to test".postln;
});

(
SynthDef(\fdnReverb, {
  arg in=0, out=0,
      size=0.5, feedback=0.9, hpass=200, lpass=16000, wet=0.5,
      earlyMix=0.3;  // Mix level for early reflections

  var input, dry, snd, processed, wet_sig;
  var step2, step4;  // Early reflections
  var hadamard;
  var lfo1, lfo2, lfo3, lfo4, lfo5, lfo6, lfo7, lfo8;

  // 8x8 Hadamard matrix (normalized by 1/sqrt(8) ≈ 0.35355)
  hadamard = [
    [ 1,  1,  1,  1,  1,  1,  1,  1],
    [ 1, -1,  1, -1,  1, -1,  1, -1],
    [ 1,  1, -1, -1,  1,  1, -1, -1],
    [ 1, -1, -1,  1,  1, -1, -1,  1],
    [ 1,  1,  1,  1, -1, -1, -1, -1],
    [ 1, -1,  1, -1, -1,  1, -1,  1],
    [ 1,  1, -1, -1, -1, -1,  1,  1],
    [ 1, -1, -1,  1, -1,  1,  1, -1]
  ] * (1/8).sqrt;

  // Input (stereo)
  input = In.ar(in, 2);
  dry = input;

  // Expand stereo to 8 channels
  snd = input.wrapExtend(8);

  // DIFFUSION: 4 steps of 8 delays with Hadamard mixing
  // Step 1
  snd = DelayC.ar(snd, 0.05,
    [0.0020, 0.0039, 0.0060, 0.0081, 0.0104, 0.0131, 0.0154, 0.0185] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;

  // Step 2
  snd = DelayC.ar(snd, 0.05,
    [0.0216, 0.0239, 0.0272, 0.0296, 0.0315, 0.0329, 0.0375, 0.0396] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;
  step2 = snd;  // Capture early reflections from step 2

  // Step 3
  snd = DelayC.ar(snd, 0.1,
    [0.0405, 0.0462, 0.0518, 0.0585, 0.0606, 0.0686, 0.0712, 0.0795] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;

  // Step 4
  snd = DelayC.ar(snd, 0.2,
    [0.0873, 0.0974, 0.1022, 0.1113, 0.1231, 0.1366, 0.1449, 0.1539] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;
  step4 = snd;  // Capture early reflections from step 4

  // FEEDBACK: 8 delays with Hadamard matrix mixing
  snd = snd + LocalIn.ar(8);

  // Slow LFOs for subtle modulation (different frequencies for each delay line)
  lfo1 = SinOsc.kr(0.07) * 0.0005;
  lfo2 = SinOsc.kr(0.11) * 0.0005;
  lfo3 = SinOsc.kr(0.13) * 0.0005;
  lfo4 = SinOsc.kr(0.17) * 0.0005;
  lfo5 = SinOsc.kr(0.19) * 0.0005;
  lfo6 = SinOsc.kr(0.23) * 0.0005;
  lfo7 = SinOsc.kr(0.29) * 0.0005;
  lfo8 = SinOsc.kr(0.31) * 0.0005;

  // Main delay lines
  snd = DelayC.ar(snd, 0.25,
    ([0.1017, 0.1195, 0.1292, 0.1464, 0.1537, 0.1741, 0.1873, 0.1902] * size)
    + [lfo1, lfo2, lfo3, lfo4, lfo5, lfo6, lfo7, lfo8]
    - ControlDur.ir
  );

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

  // Add early reflections from step 2 and step 4
  processed = processed + (Splay.ar(step2) * earlyMix) + (Splay.ar(step4) * earlyMix);

  // Mix dry and wet
  wet_sig = (dry * (1 - wet)) + (processed * wet);

  Out.ar(out, wet_sig);
}).add;

~recreateReverb.(size: 0.7, feedback: 0.98);
)

// Trigger clap (routed to reverb bus 10)
Synth(\hh, [\len, 0.2,\out, ~reverbBus.index, \amp, 0.8]);
