// Modulated Rhythm Pattern - Routine-based implementation
// Controls two rhythmic patterns (kick and hihat) with ritardando distribution
// Receives OSC control messages from Go TUI on port 57120

(
// Global state dictionary
~modRhy = ~modRhy ?? ();

// Pattern state
~modRhy.phraseLength = 16; // ticks per phrase
~modRhy.tickDur = 0.125; // duration of one tick in seconds (125ms = 120 BPM with 4 ticks per beat)

// Kick task - fires on first half of phrase ticks
~modRhy.kickTask = Task({
	var tick = 0;

	inf.do {
		// Fire synth only on first half of phrase
		if(tick < (~modRhy.phraseLength / 2), {
		  s.bind {
  			Synth(\bd, [
  				\freq, 50,
  				\amp, 0.8,
  				\len, ~modRhy.tickDur * 0.75,
  				\out, 0
  			], target: 100);
      };
		});

		// Increment tick counter and wrap at phrase length
		tick = (tick + 1) % ~modRhy.phraseLength;

		// Yield tick duration in seconds (will reschedule for this many seconds later)
		~modRhy.tickDur.yield;
	};
}, SystemClock);

// Hihat task - fires on first half of phrase ticks
~modRhy.hihatTask = Task({
	var tick = 0;

	inf.do {
		// Fire synth only on first half of phrase
		if(tick < (~modRhy.phraseLength / 2), {
		  s.bind {
  			Synth(\hh, [
  				\amp, 0.6,
  				\len, ~modRhy.tickDur * 0.75,
  				\out, 0
  			], target: 100);
      };
		});

		// Increment tick counter and wrap at phrase length
		tick = (tick + 1) % ~modRhy.phraseLength;

		// Yield tick duration in seconds (will reschedule for this many seconds later)
		~modRhy.tickDur.yield;
	};
}, SystemClock);

// OSC Responders for control from Go TUI

// Play/Stop
OSCdef(\modRhyPlay, {
	"[mod_rhy] Playing".postln;
	~modRhy.kickTask.start;
	~modRhy.hihatTask.start;
}, '/pattern/mod_rhy/play');

OSCdef(\modRhyStop, {
	"[mod_rhy] Stopped".postln;
	~modRhy.kickTask.stop;
	~modRhy.hihatTask.stop;
}, '/pattern/mod_rhy/stop');

// Tick duration control
OSCdef(\modRhyTickDur, { |msg|
	~modRhy.tickDur = msg[1].asFloat;
	"[mod_rhy] Tick duration: %s".format(~modRhy.tickDur).postln;
}, '/pattern/mod_rhy/tick_dur');

"[mod_rhy] Pattern loaded and ready. Listening for OSC on port 57120".postln;
"[mod_rhy] Send /pattern/mod_rhy/play to start".postln;
)
