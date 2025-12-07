// SynthDefs for Forbidden Sequencer
// These synthdefs respond to MIDI input from the sequencer

// FDN reverb with Hadamard diffusion and Hadamard feedback
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

// Allpass reverb
SynthDef(\allpassReverb, {
  arg in=0, out=0,
      size=0.5, decay=0.5, wet=0.5,
      hpass=200, lpass=8000;

  var input, processed, dry, wet_sig;
  var max_delay, base_time;
  var lfo0, lfo1, lfo2;
  var fb, sig, lfoL, lfoR, lfoMod, hp, lp;

  // Input
  input = In.ar(in, 2);
  dry = input;

  // LFO generation
  lfo0 = SinOsc.ar(0.9128) * 11;
  lfo1 = SinOsc.ar(1.1341) * 9;
  lfo2 = SinOsc.ar(1.0) * 10;

  // Timing calculations
  max_delay = 0.3;
  base_time = size * (1/150); // 0.006666667 seconds

  // Process stereo with LocalIn/LocalOut (shared feedback bus)
  fb = LocalIn.ar(2);
  sig = input + (fb * decay);

  // Stereo LFO modulation (different per channel)
  lfoL = [lfo0, lfo1, lfo2];
  lfoR = [lfo2, lfo0, lfo1];
  lfoMod = [lfoL, lfoR].flop;  // [[lfo0,lfo2], [lfo1,lfo0], [lfo2,lfo1]]

  sig = AllpassC.ar(sig, max_delay, (base_time * 3) + (lfoMod[0] * 0.001), 0.5);
  sig = AllpassC.ar(sig, max_delay, (base_time * 7) + (lfoMod[1] * 0.001), 0.5);
  sig = AllpassC.ar(sig, max_delay, (base_time * 11) + (lfoMod[2] * 0.001), 0.5);
  sig = AllpassC.ar(sig, max_delay, (base_time * 19) + (lfoMod[0] * 0.001), 0.5);
  sig = AllpassC.ar(sig, max_delay, (base_time * 23) + (lfoMod[1] * 0.001), 0.5);

  hp = HPF.ar(sig, hpass);
  lp = LPF.ar(hp, lpass);

  sig = AllpassC.ar(lp, max_delay, (base_time * 31) + (lfoMod[2] * 0.001), 0.5);

  LocalOut.ar(sig);
  processed = sig;

  // Mix dry and wet
  wet_sig = (dry * (1 - wet)) + (processed * wet);

  Out.ar(out, wet_sig);
}).add;

// SynthDef for high-hat (hh)
SynthDef(\hh, {
  arg freq = 440, len = 1, amp = 1, out = 0;
  var sig = BPF.ar(Hasher.ar(Sweep.ar), [7500, 7000, 8000], 0.3).tanh * 0.33;
  sig = sig * EnvGen.ar(Env.perc(0.001, len - 0.001), doneAction: Done.freeSelf);
  sig = sig.sum * amp;
  Out.ar(out, sig ! 2);
}).add;

// SynthDef for clap (cp)
SynthDef(\cp, {
  arg freq = 440, len = 1, amp = 1, out = 0;
  var sig, left, right;
  sig = BPF.ar(Hasher.ar(Sweep.ar), [1320, 1100, 1420], 0.1) * 10.dbamp;
  left = sig * Env([0, 1, 0, 1, 0, 1, 0], [Rand(0.001, 0.01), 0.01, 0.001, 0.01, 0.001, 0.08]).ar(Done.freeSelf);
  right = sig * Env([0, 1, 0, 1, 0, 1, 0], [Rand(0.001, 0.01), 0.01, 0.001, 0.01, 0.001, 0.08]).ar;
  sig = [left, right].tanh * amp;
  Out.ar(out, sig);
}).add;

// SynthDef for bass drum (bd)
SynthDef(\bd, {
  arg freq = 50, len = 1, amp = 1, out = 0, ratio = 7, sweep = 0.05;
  var fCurve = EnvGen.kr(Env([freq * ratio, freq], [sweep], -5)),
  env = EnvGen.kr(Env([1, 0.8, 0], [len * 0.7, len * 0.3], -4), doneAction: Done.freeSelf),
  sig = SinOsc.ar(fCurve, 0.5pi).tanh * env * amp;
  Out.ar(out, sig ! 2);
}).add;

"SynthDefs loaded successfully".postln;
