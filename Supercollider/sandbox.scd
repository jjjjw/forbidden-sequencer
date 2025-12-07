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
      size=0.5, feedback=0.9, hpass=200, lpass=16000, wet=0.5;

  var input, dry, snd, processed, wet_sig;
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
    [0.0051, 0.0053, 0.0059, 0.0061, 0.0067, 0.0071, 0.0073, 0.0079] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;

  // Step 2
  snd = DelayC.ar(snd, 0.05,
    [0.0083, 0.0089, 0.0097, 0.0101, 0.0103, 0.0107, 0.0109, 0.0113] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;

  // Step 3
  snd = DelayC.ar(snd, 0.05,
    [0.0127, 0.0131, 0.0137, 0.0139, 0.0149, 0.0151, 0.0157, 0.0163] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;

  // Step 4
  snd = DelayC.ar(snd, 0.05,
    [0.0167, 0.0173, 0.0179, 0.0181, 0.0191, 0.0193, 0.0197, 0.0199] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;

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

  // Main delay lines: mutually incommensurate values for dense reflections
  // With subtle, slow modulation (100-200ms range)
  snd = DelayC.ar(snd, 0.25,
    ([0.113, 0.127, 0.137, 0.149, 0.157, 0.167, 0.179, 0.191] * size)
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

  // Mix dry and wet
  wet_sig = (dry * (1 - wet)) + (processed * wet);

  Out.ar(out, wet_sig);
}).add;

~recreateReverb.(size: 0.9, feedback: 0.9);
)

// Trigger clap (routed to reverb bus 10)
Synth(\hh, [\len, 0.2,\out, ~reverbBus.index, \amp, 0.8]);
