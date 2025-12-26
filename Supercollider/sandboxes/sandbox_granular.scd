// Simple granular synthesis
// Load a buffer from file and generate 50 grains at varying frequencies

s.waitForBoot({

  // Load buffer from audio file
  // Change this path to your audio file
  ~grainBuffer = Buffer.read(s, "/Users/jj/Downloads/2023-04-17-breaks/amen.wav");

  s.sync;

  "Buffer loaded: % frames, % channels, % seconds".format(
    ~grainBuffer.numFrames,
    ~grainBuffer.numChannels,
    ~grainBuffer.duration
  ).postln;

});

// Simple granular synth with 50 grains at varying frequencies
(
SynthDef(\simpleGranular, {
  arg out=0, buffer, amp=0.5;

  var grains, triggers, positions, rates;

  // Create 50 triggers with random timing offsets
  triggers = Array.fill(50, { |i|
    Impulse.ar(Rand(5, 20), Rand(0, 1.0))
  });

  // Random positions in buffer for each grain
  positions = Array.fill(50, { |i|
    TRand.ar(0, 1, triggers[i])
  });

  // Varying playback rates (frequency variations)
  // Range from 0.5 (half speed/octave down) to 2.0 (double speed/octave up)
  rates = Array.fill(50, { |i|
    Rand(0.5, 2.0)
  });

  // Generate 50 grains
  grains = Array.fill(50, { |i|
    GrainBuf.ar(
      1,                    // mono output per grain
      triggers[i],          // trigger
      Rand(0.02, 0.1),      // grain size (20-100ms)
      buffer,               // buffer to read from
      rates[i],             // playback rate
      positions[i],         // position in buffer
      2                     // interpolation
    )
  });

  // Mix all grains and output as stereo
  grains = Mix(grains) * amp;

  Out.ar(out, Pan2.ar(grains, 0));

}).add;
)

// Play the granular synth
~granularSynth = Synth(\simpleGranular, [\buffer, ~grainBuffer, \amp, 0.3]);

// Stop
// ~granularSynth.free;
