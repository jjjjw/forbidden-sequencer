// SynthDefs for Forbidden Sequencer
// These synthdefs respond to MIDI input from the sequencer

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
