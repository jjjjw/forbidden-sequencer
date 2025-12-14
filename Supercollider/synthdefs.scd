// SynthDefs for Forbidden Sequencer
// These synthdefs respond to MIDI input from the sequencer

// FDN reverb with Hadamard diffusion and Hadamard feedback
SynthDef(\fdnReverb, {
  arg in=0, out=0,
      size=0.5, feedback=0.9, hpass=200, lpass=16000, wet=0.5,
      earlyMix=0.3;  // Mix level for early reflections

  var input, dry, snd, processed, wet_sig;
  var er1, er2;  // Early reflections
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

  // DIFFUSION: 4 stages of 8 delays with Hadamard mixing
  snd = DelayC.ar(snd, 0.05,
    [0.0020, 0.0039, 0.0060, 0.0081, 0.0104, 0.0131, 0.0154, 0.0185] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;

  snd = DelayC.ar(snd, 0.05,
    [0.0216, 0.0239, 0.0272, 0.0296, 0.0315, 0.0329, 0.0375, 0.0396] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;
  er1 = snd;  // Capture early reflections

  snd = DelayC.ar(snd, 0.1,
    [0.0405, 0.0462, 0.0518, 0.0585, 0.0606, 0.0686, 0.0712, 0.0795] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;

  snd = DelayC.ar(snd, 0.2,
    [0.0873, 0.0974, 0.1022, 0.1113, 0.1231, 0.1366, 0.1449, 0.1539] * size);
  snd = snd * hadamard.flop;
  snd = snd.sum;
  er2 = snd;  // Capture early reflection

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
  processed = processed + (Splay.ar(er1) * earlyMix) + (Splay.ar(er2) * earlyMix);

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

// SynthDef for 2-operator FM synthesis (fm2op)
SynthDef(\fm2op, {
  arg freq = 440, amp = 1, len = 1, out = 0, modRatio = 2.0, modIndex = 1.0;
  var carrier, modulator, modFreq, env, sig;

  // Modulator frequency (ratio relative to carrier)
  modFreq = freq * modRatio;

  // Amplitude envelope (percussive)
  env = EnvGen.kr(Env.perc(0.001, len - 0.001), doneAction: Done.freeSelf);

  // Modulator: sine wave scaled by modulation index
  modulator = SinOsc.ar(modFreq) * modFreq * modIndex;

  // Carrier: frequency modulated by the modulator
  carrier = SinOsc.ar(freq + modulator);

  // Apply envelope and amplitude
  sig = carrier * env * amp;

  Out.ar(out, sig ! 2);
}).add;

"SynthDefs loaded successfully".postln;
