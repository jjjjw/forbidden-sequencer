// Live granular effect
// Records input into a buffer and plays back with 2 grains per channel

s.waitForBoot({

  // Allocate 2 mono buffers for live recording (1 second each)
  if (~liveGrainBufferL.notNil, {
    ~liveGrainBufferL.free;
  });
  if (~liveGrainBufferR.notNil, {
    ~liveGrainBufferR.free;
  });
  ~liveGrainBufferL = Buffer.alloc(s, s.sampleRate * 1, 1);
  ~liveGrainBufferR = Buffer.alloc(s, s.sampleRate * 1, 1);
  "Allocated 2 mono grain buffers: % seconds each".format(1).postln;

  s.sync;

  // Create audio bus for routing
  if (~granularBus.isNil, {
    ~granularBus = Bus.audio(s, 2);
    "Created granular bus: %".format(~granularBus.index).postln;
  });

});

// Live granular effect (2 grains per channel)
(
SynthDef(\liveGranular, {
  arg in=0, out=0, bufferL, bufferR,
      grainSize=0.05, grainRate=20, amp=1.0;

  var input, inputL, inputR;
  var grainsL, grainsR, grains;
  var trig1, trig2, pos1, pos2;

  // Input (stereo)
  input = In.ar(in, 2);
  inputL = input[0];
  inputR = input[1];

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
  grains = [grainsL, grainsR] * amp;

  Out.ar(out, grains);

}).add;
)

// Simple pulse oscillator to test with
(
SynthDef(\pulseTest, {
  arg out=0, freq=200, width=0.5, amp=0.3;

  var sig;

  sig = Pulse.ar(freq, width) * amp;
  sig = sig ! 2; // Convert to stereo

  Out.ar(out, sig);

}).add;
)

// Start the effect chain
(
// Pulse oscillator -> granular bus
~pulse = Synth(\pulseTest, [
  \out, ~granularBus.index,
  \freq, 200,
  \width, 0.5,
  \amp, 0.3
]);

// Granular effect -> output
~granular = Synth.after(~pulse, \liveGranular, [
  \in, ~granularBus.index,
  \out, 0,
  \bufferL, ~liveGrainBufferL,
  \bufferR, ~liveGrainBufferR,
  \grainSize, 0.05,
  \grainRate, 20,
  \amp, 0.8
]);
)

// Change pulse frequency
~pulse.set(\freq, 300);
~pulse.set(\freq, 150);

// Change grain parameters
~granular.set(\grainSize, 0.1);
~granular.set(\grainRate, 40);

// Stop
// ~pulse.free; ~granular.free;
