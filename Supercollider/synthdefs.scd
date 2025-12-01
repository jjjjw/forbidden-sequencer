// SynthDefs for Forbidden Sequencer
// These synthdefs respond to MIDI input from the sequencer

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
