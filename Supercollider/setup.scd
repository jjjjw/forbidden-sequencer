// Setup script for Forbidden Sequencer MIDI integration
// This script initializes MIDI input and maps MIDI notes to SynthDefs
// Uses monophonic voices that retrigger on each event

// Boot server if not already running
s.waitForBoot({

  // Load SynthDefs first
  (thisProcess.nowExecutingPath.dirname +/+ "synthdefs.scd").load;

  s.sync; // Wait for SynthDefs to load

  // Create audio bus for reverb (or reuse existing)
  if (~reverbBus.isNil, {
    ~reverbBus = Bus.audio(s, 2);
    "Created reverb bus: %".format(~reverbBus.index).postln;
  }, {
    "Reusing existing reverb bus: %".format(~reverbBus.index).postln;
  });

  // Free existing reverb synth if it exists
  if (~reverb.notNil, {
    ~reverb.free;
  });

  // Create reverb synth at the tail (processes after sound sources)
  ~reverb = Synth.tail(s, \allpassReverb, [
    \in, ~reverbBus,
    \out, 0,
    \size, 0.7,
    \decay, 0.6,
    \wet, 0.5,
    \hpass, 150,
    \lpass, 9000
  ]);

  "Reverb initialized on bus %".format(~reverbBus.index).postln;

  // Initialize MIDI
  MIDIClient.init;

  "MIDI initialized. Available sources:".postln;
  MIDIClient.sources.do({ arg endpoint, i;
    "  [%] %".format(i, endpoint.name).postln;
  });

  // Connect only to MIDI source at index 0
  MIDIIn.connect(0, MIDIClient.sources[0]);
  "Connected to MIDI source 0: %".format(MIDIClient.sources[0].name).postln;

  // MIDI note mapping:
  // Note 36 -> kick (bd)
  // Note 37 -> snare (cp)
  // Note 42 -> hihat (hh)

  // Monophonic voice groups
  ~kickGroup = Group.new;
  ~snareGroup = Group.new;
  ~hihatGroup = Group.new;

  // MIDIdef for bass drum (kick) - monophonic, retriggers
  MIDIdef.noteOn(\kick, { arg vel, num, chan, src;
    var amp = vel / 127.0;
    var len = 0.3;
    // Free all existing kick synths, then create new one
    ~kickGroup.freeAll;
    Synth(\bd, [\amp, amp, \len, len], ~kickGroup);
    "kick: vel=%".format(vel).postln;
  }, noteNum: 36);

  // MIDIdef for snare/clap - monophonic, retriggers, routed to reverb bus
  MIDIdef.noteOn(\snare, { arg vel, num, chan, src;
    var amp = vel / 127.0;
    var len = 0.3;
    // Free all existing snare synths, then create new one
    ~snareGroup.freeAll;
    Synth(\cp, [\amp, amp, \len, len, \out, ~reverbBus], ~snareGroup);
    "clap: vel=%".format(vel).postln;
  }, noteNum: 37);

  // MIDIdef for hihat - monophonic, retriggers, routed to reverb bus
  MIDIdef.noteOn(\hihat, { arg vel, num, chan, src;
    var amp = vel / 127.0;
    var len = 0.15;
    // Free all existing hihat synths, then create new one
    ~hihatGroup.freeAll;
    Synth(\hh, [\amp, amp, \len, len, \out, ~reverbBus], ~hihatGroup);
    "hihat: vel=%".format(vel).postln;
  }, noteNum: 42);

  "MIDI responders configured successfully".postln;
  "Listening for MIDI notes: 36 (kick), 37 (snare/clap), 42 (hihat)".postln;
  "Each voice retriggers on new MIDI events".postln;
});
