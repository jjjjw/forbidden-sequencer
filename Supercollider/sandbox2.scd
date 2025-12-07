// Granular reverb with FDN diffusion
// Based on random grain shuffling + regenerated processing
// Input → Buffer → Granular playback → FDN diffusion/feedback

s.waitForBoot({

  // Load SynthDefs
  (thisProcess.nowExecutingPath.dirname +/+ "synthdefs.scd").load;

  s.sync;

  // Create reverb bus
  if (~reverbBus.isNil, {
    ~reverbBus = Bus.audio(s, 2, 10);
    "Created reverb bus: %".format(~reverbBus.index).postln;
  }, {
    "Reusing existing reverb bus: %".format(~reverbBus.index).postln;
  });

  // Allocate buffer for granular recording (4 seconds, stereo)
  if (~grainBuffer.notNil, {
    ~grainBuffer.free;
  });
  ~grainBuffer = Buffer.alloc(s, s.sampleRate * 4, 2);
  "Allocated grain buffer: % seconds, 2 channels".format(4).postln;

  s.sync;

  // Free existing reverb synth if it exists
  if (~granularReverb.notNil, {
    ~granularReverb.free;
  });

  // Create granular reverb synth
  ~granularReverb = Synth.tail(s, \granularFdnReverb, [
    \in, ~reverbBus.index,
    \out, 0,
    \buffer, ~grainBuffer,
    \grainSize, 0.05,
    \grainRate, 20,
    \feedback, 0.85,
    \wet, 0.5,
    \hpass, 200,
    \lpass, 16000
  ]);

  "Granular reverb initialized on bus %".format(~reverbBus.index).postln;

  // Define helper function: Recreate reverb with new parameters
  ~recreateGranularReverb = {
    arg grainSize=0.05, grainRate=20, feedback=0.85, wet=0.5, hpass=200, lpass=16000;

    "--- Recreating granular reverb ---".postln;

    if (~granularReverb.notNil, {
      ~granularReverb.free;
      "Destroyed granular reverb synth".postln;
    });

    ~granularReverb = Synth.tail(s, \granularFdnReverb, [
      \in, ~reverbBus.index,
      \out, 0,
      \buffer, ~grainBuffer,
      \grainSize, grainSize,
      \grainRate, grainRate,
      \feedback, feedback,
      \wet, wet,
      \hpass, hpass,
      \lpass, lpass
    ]);

    "Recreated granular reverb: grainSize=%, grainRate=%, feedback=%, wet=%".format(
      grainSize, grainRate, feedback, wet
    ).postln;
  };

  "Sandbox2 setup complete - use ~recreateGranularReverb.value() to test".postln;
});

(
SynthDef(\granularFdnReverb, {
  arg in=0, out=0, buffer,
      grainSize=0.05, grainRate=20,
      feedback=0.85, hpass=200, lpass=16000, wet=0.5;

  var input, dry, rec, snd, grains, processed, wet_sig;
  var hadamard;
  var lfo1, lfo2, lfo3, lfo4, lfo5, lfo6, lfo7, lfo8;
  var trig1, trig2, pos1, pos2;
  var bufDur, grainL, grainR;

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

  // Record input to buffer (continuous recording with looping)
  RecordBuf.ar(input, buffer, loop: 1);

  // Buffer duration in seconds
  bufDur = BufDur.kr(buffer);

  // Granular synthesis: 2 grains per channel
  // Triggers at specified grain rate with slight randomization
  trig1 = Impulse.ar(grainRate);
  trig2 = Impulse.ar(grainRate, 0.5); // Offset by half phase

  // Random positions in buffer for grain playback (normalized 0-1)
  pos1 = TRand.ar(0, 1, trig1);
  pos2 = TRand.ar(0, 1, trig2);

  // Stereo grains: 2 grains reading from stereo buffer
  grains = GrainBuf.ar(2, trig1, grainSize, buffer, 1, pos1, 2)
         + GrainBuf.ar(2, trig2, grainSize, buffer, 1, pos2, 2);

  grains = grains * 0.5; // Scale down since we're adding 2 grains

  // Expand stereo to 8 channels for FDN processing
  snd = grains.wrapExtend(8);

  // FEEDBACK: 8 delays with Hadamard matrix mixing
  snd = snd + LocalIn.ar(8);

  // Slow LFOs for subtle modulation
  lfo1 = SinOsc.kr(0.07) * 0.0005;
  lfo2 = SinOsc.kr(0.11) * 0.0005;
  lfo3 = SinOsc.kr(0.13) * 0.0005;
  lfo4 = SinOsc.kr(0.17) * 0.0005;
  lfo5 = SinOsc.kr(0.19) * 0.0005;
  lfo6 = SinOsc.kr(0.23) * 0.0005;
  lfo7 = SinOsc.kr(0.29) * 0.0005;
  lfo8 = SinOsc.kr(0.31) * 0.0005;

  // Main delay lines (100-200ms range)
  snd = DelayC.ar(snd, 0.25,
    [0.113, 0.127, 0.137, 0.149, 0.157, 0.167, 0.179, 0.191]
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

~recreateGranularReverb.(grainSize: 0.05, grainRate: 20, feedback: 0.85);
)

// Trigger hi-hat (routed to reverb bus 10)
Synth(\hh, [\len, 0.2, \out, ~reverbBus.index, \amp, 0.8]);
