// Setup script for Forbidden Sequencer OSC integration
// This script initializes OSC responders and maps OSC messages to SynthDefs
// OSC bundles with timestamps are automatically scheduled by SuperCollider
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

  // OSC message format from Go sequencer (sent as timestamped bundles):
  // /trigger/<eventname> <freq> <velocity> <duration> <c> <d>
  // SuperCollider automatically schedules the OSCdef callback at the bundle timestamp
  // For example:
  // Bundle with timestamp -> /trigger/kick 60.0 0.8 0.3 0.0 0.0

  // Monophonic voice groups
  ~kickGroup = Group.new;
  ~snareGroup = Group.new;
  ~hihatGroup = Group.new;

  // Enable OSC listening on custom port 57121
  thisProcess.openUDPPort(57121);
  "UDP port 57121 opened for OSC listening".postln;

  // Free existing OSC responders
  OSCdef(\kick).free;
  OSCdef(\snare).free;
  OSCdef(\hihat).free;

  // OSCdef for bass drum (kick) - monophonic, retriggers
  OSCdef(\kick, { arg msg, time, addr, recvPort;
    var freq, vel, dur, c, d, amp, len;

    freq = msg[1]; // frequency (can use for pitch variations)
    vel = msg[2];  // velocity (0.0-1.0)
    dur = msg[3];  // duration in seconds
    c = msg[4];    // additional parameter (unused for now)
    d = msg[5];    // additional parameter (unused for now)

    amp = vel;
    len = dur;

    // SuperCollider automatically schedules this callback at the bundle timestamp
    ~kickGroup.freeAll;
    Synth(\bd, [\freq, 50, \amp, amp, \len, len], ~kickGroup);
  }, '/trigger/kick');

  // OSCdef for snare/clap - monophonic, retriggers, routed to reverb bus
  OSCdef(\snare, { arg msg, time, addr, recvPort;
    var freq, vel, dur, c, d, amp, len;
    freq = msg[1];
    vel = msg[2];
    dur = msg[3];
    c = msg[4];
    d = msg[5];

    amp = vel;
    len = dur;

    ~snareGroup.freeAll;
    Synth(\cp, [\amp, amp, \len, len, \out, ~reverbBus], ~snareGroup);
  }, '/trigger/snare');

  // OSCdef for hihat - monophonic, retriggers
  OSCdef(\hihat, { arg msg, time, addr, recvPort;
    var freq, vel, dur, c, d, amp, len;
    freq = msg[1];
    vel = msg[2];
    dur = msg[3];
    c = msg[4];
    d = msg[5];

    amp = vel;
    len = dur;

    ~hihatGroup.freeAll;
    Synth(\hh, [\amp, amp, \len, len], ~hihatGroup);
  }, '/trigger/hihat');

  "OSC responders configured successfully".postln;
  "Listening on port 57121 for timestamped OSC bundles:".postln;
  "  /trigger/kick <freq> <vel> <dur> <c> <d>".postln;
  "  /trigger/snare <freq> <vel> <dur> <c> <d>".postln;
  "  /trigger/hihat <freq> <vel> <dur> <c> <d>".postln;
});
