// Granular reverb with FDN feedback
// Input → (+feedback) → Record to Buffer → Grains → FDN → Output
//          ↑                                           ↓
//          └──────────── Feedback ─────────────────────┘

s.waitForBoot({

  // Allocate 2 mono buffers for live recording (1 second each)
  if (~grainBufferL.notNil, {
    ~grainBufferL.free;
  });
  if (~grainBufferR.notNil, {
    ~grainBufferR.free;
  });
  ~grainBufferL = Buffer.alloc(s, s.sampleRate * 1, 1);
  ~grainBufferR = Buffer.alloc(s, s.sampleRate * 1, 1);
  "Allocated 2 mono grain buffers: % seconds each".format(1).postln;

  s.sync;

  // Create audio bus for routing input
  if (~grainInputBus.isNil, {
    ~grainInputBus = Bus.audio(s, 2);
    "Created grain input bus: %".format(~grainInputBus.index).postln;
  });

});

// Granular reverb with FDN feedback
(
SynthDef(\granularReverbFDN, {
  arg in=0, out=0, bufferL, bufferR,
      grainSize=0.05, grainRate=20,
      feedback=0.85, fbAmt=0.5,
      hpass=200, lpass=12000,
      wet=0.5, amp=1.0;

  var input, dry, inputWithFb, inputL, inputR;
  var grainsL, grainsR, grains;
  var trig1, trig2, pos1, pos2;
  var snd, processed, wet_sig, fbSignal;
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

  // Add feedback to input
  fbSignal = LocalIn.ar(2);
  inputWithFb = input + (fbSignal * fbAmt);

  // Split to L/R for recording
  inputL = inputWithFb[0];
  inputR = inputWithFb[1];

  // Record each channel to its own mono buffer (continuous recording with looping)
  RecordBuf.ar(inputL, bufferL, loop: 1);
  RecordBuf.ar(inputR, bufferR, loop: 1);

  // 2 triggers per channel, offset by half phase
  trig1 = Impulse.ar(grainRate);
  trig2 = Impulse.ar(grainRate, 0.5);

  // Random positions in buffer for grain playback (normalized 0-1)
  pos1 = TRand.ar(0, 1, trig1);
  pos2 = TRand.ar(0, 1, trig2);

  // Left channel: 2 grains reading from left mono buffer
  grainsL = GrainBuf.ar(1, trig1, grainSize, bufferL, 1, pos1, 2, 0)
          + GrainBuf.ar(1, trig2, grainSize, bufferL, 1, pos2, 2, 0);

  // Right channel: 2 grains reading from right mono buffer
  grainsR = GrainBuf.ar(1, trig1, grainSize, bufferR, 1, pos1, 2, 0)
          + GrainBuf.ar(1, trig2, grainSize, bufferR, 1, pos2, 2, 0);

  // Combine to stereo
  grains = [grainsL, grainsR];

  // === FDN Processing ===

  // Expand stereo to 4 channels
  snd = grains.wrapExtend(4);

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

  // Send FDN output back to grain buffer input
  LocalOut.ar(processed);

  // Mix dry and wet
  wet_sig = (dry * (1 - wet)) + (processed * wet);
  wet_sig = wet_sig * amp;

  Out.ar(out, wet_sig);

}).add;
);

// Simple pulse oscillator to test with
(
SynthDef(\pulseTest, {
  arg out=0, freq=200, width=0.5, amp=0.3;

  var sig;

  sig = Pulse.ar(freq, width) * amp;
  sig = sig ! 2; // Convert to stereo

  Out.ar(out, sig);

}).add;
);

// Start the effect chain
(
// Pulse oscillator -> granular input bus
~pulse = Synth(\pulseTest, [
  \out, ~grainInputBus.index,
  \freq, 200,
  \width, 0.5,
  \amp, 0.3
]);

// Granular reverb with FDN feedback -> output
~granularReverb = Synth.after(~pulse, \granularReverbFDN, [
  \in, ~grainInputBus.index,
  \out, 0,
  \bufferL, ~grainBufferL,
  \bufferR, ~grainBufferR,
  \grainSize, 0.05,
  \grainRate, 20,
  \feedback, 0.85,
  \fbAmt, 0.3,
  \hpass, 200,
  \lpass, 12000,
  \wet, 0.5,
  \amp, 0.8
]);
)

// Change pulse frequency
// ~pulse.set(\freq, 300);
// ~pulse.set(\freq, 150);

// Change grain parameters
// ~granularReverb.set(\grainSize, 0.1);
// ~granularReverb.set(\grainRate, 40);

// Change FDN parameters
// ~granularReverb.set(\feedback, 0.9);
// ~granularReverb.set(\fbAmt, 0.5);  // Amount of FDN output fed back to input
// ~granularReverb.set(\wet, 0.7);

// Stop
// ~pulse.free; ~granularReverb.free;
