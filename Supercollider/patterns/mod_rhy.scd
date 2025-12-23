// Modulated Rhythm Pattern - Routine-based implementation
// Controls two rhythmic patterns (kick and hihat) with ritardando distribution
// Receives OSC control messages from Go TUI on port 57120

(
// Global state dictionary
~modRhy = ~modRhy ?? ();

// Pattern state
~modRhy.phraseLength = 16; // ticks per phrase
~modRhy.tickDur = 0.125; // duration of one tick in seconds (125ms = 120 BPM with 4 ticks per beat)

// Kick state
~modRhy.kickCurve = 2.0;
~modRhy.kickEvents = 8;
~modRhy.kickSubdiv = 1;
~modRhy.kickQuantized = true;
~modRhy.kickDist = nil; // will hold array of event positions
~modRhy.kickTask = nil;
~modRhy.kickPaused = true;

// Hihat state
~modRhy.hihatCurve = 1.5;
~modRhy.hihatEvents = 6;
~modRhy.hihatSubdiv = 1;
~modRhy.hihatQuantized = true;
~modRhy.hihatDist = nil; // will hold array of event positions
~modRhy.hihatTask = nil;
~modRhy.hihatPaused = true;

// Ritardando distribution function
// Returns array of event positions (in ticks, as floats for sub-tick precision)
~ritardando = { |numEvents, phraseLength, curve|
	var positions = Array.new(numEvents);

	numEvents.do { |i|
		var t, curved, pos;

		// Normalize position 0.0 to 1.0
		t = if(numEvents == 1, { 0 }, { i / (numEvents - 1) });

		// Apply curve (events get further apart)
		curved = t.pow(curve);

		// Map to phrase length
		pos = curved * (phraseLength - 1);
		positions.add(pos);
	};

	positions;
};

// Generate kick distribution
~modRhy.generateKickDist = {
	~modRhy.kickDist = ~ritardando.value(
		~modRhy.kickEvents,
		~modRhy.phraseLength,
		~modRhy.kickCurve
	);
	"Kick distribution generated: %".format(~modRhy.kickDist).postln;
};

// Generate hihat distribution
~modRhy.generateHihatDist = {
	~modRhy.hihatDist = ~ritardando.value(
		~modRhy.hihatEvents,
		~modRhy.phraseLength,
		~modRhy.hihatCurve
	);
	"Hihat distribution generated: %".format(~modRhy.hihatDist).postln;
};

// Check if event should fire at this tick
~modRhy.shouldFire = {
	arg tickInPhrase, distribution, quantized;
	var shouldFire = false;

	if(quantized, {
		// Quantized mode: round positions to integers and check
		distribution.do { |pos|
			if(pos.round.asInteger == tickInPhrase, {
				shouldFire = true;
			});
		};
	}, {
		// Non-quantized mode: check if position falls within this tick
		distribution.do { |pos|
			if((pos.floor.asInteger == tickInPhrase), {
				shouldFire = true;
			});
		};
	});

	shouldFire;
};

// Get sub-tick offset for non-quantized mode
~modRhy.getOffset = {
	arg tickInPhrase, distribution;
	var offset = 0.0;

	distribution.do { |pos|
		if(pos.floor.asInteger == tickInPhrase, {
			offset = pos - pos.floor;
		});
	};

	offset;
};

// Kick routine
~modRhy.kickRoutine = Routine {
	var tickInPhrase = 0;

	inf.do {
		if(~modRhy.kickPaused.not, {
			var shouldFire, offset, waitTime;

			shouldFire = ~modRhy.shouldFire.(tickInPhrase, ~modRhy.kickDist, ~modRhy.kickQuantized);

			if(shouldFire, {
				offset = if(~modRhy.kickQuantized, { 0.0 }, {
					~modRhy.getOffset.(tickInPhrase, ~modRhy.kickDist);
				});

				// Schedule synth(s) based on subdivision
				~modRhy.kickSubdiv.do { |i|
					var subdiv_offset = offset + (i / ~modRhy.kickSubdiv);
					var delay = subdiv_offset * ~modRhy.tickDur;

					SystemClock.sched(delay, {
						Synth(\bd, [
							\freq, 50,
							\amp, 0.8,
							\len, 0.75 / ~modRhy.kickSubdiv,
							\out, 0
						], target: 100); // synths group
					});
				};
			});
		});

		// Advance tick
		tickInPhrase = (tickInPhrase + 1) % ~modRhy.phraseLength;

		// Wait for next tick
		~modRhy.tickDur.yield;
	};
};

// Hihat routine
~modRhy.hihatRoutine = Routine {
	var tickInPhrase = 0;

	inf.do {
		if(~modRhy.hihatPaused.not, {
			var shouldFire, offset, waitTime;

			shouldFire = ~modRhy.shouldFire.(tickInPhrase, ~modRhy.hihatDist, ~modRhy.hihatQuantized);

			if(shouldFire, {
				offset = if(~modRhy.hihatQuantized, { 0.0 }, {
					~modRhy.getOffset.(tickInPhrase, ~modRhy.hihatDist);
				});

				// Schedule synth(s) based on subdivision
				~modRhy.hihatSubdiv.do { |i|
					var subdiv_offset = offset + (i / ~modRhy.hihatSubdiv);
					var delay = subdiv_offset * ~modRhy.tickDur;

					SystemClock.sched(delay, {
						Synth(\hh, [
							\amp, 0.6,
							\len, 0.75 / ~modRhy.hihatSubdiv,
							\out, 0
						], target: 100); // synths group
					});
				};
			});
		});

		// Advance tick
		tickInPhrase = (tickInPhrase + 1) % ~modRhy.phraseLength;

		// Wait for next tick
		~modRhy.tickDur.yield;
	};
};

// Initialize distributions
~modRhy.generateKickDist.value;
~modRhy.generateHihatDist.value;

// Create tasks
~modRhy.kickTask = Task(~modRhy.kickRoutine, TempoClock.default);
~modRhy.hihatTask = Task(~modRhy.hihatRoutine, TempoClock.default);

// OSC Responders for control from Go TUI

// Play/Stop
OSCdef(\modRhyPlay, {
	"[mod_rhy] Playing".postln;
	~modRhy.kickPaused = false;
	~modRhy.hihatPaused = false;
	~modRhy.kickTask.resume;
	~modRhy.hihatTask.resume;
}, '/pattern/mod_rhy/play');

OSCdef(\modRhyStop, {
	"[mod_rhy] Stopped".postln;
	~modRhy.kickPaused = true;
	~modRhy.hihatPaused = true;
	~modRhy.kickTask.pause;
	~modRhy.hihatTask.pause;
}, '/pattern/mod_rhy/stop');

// Kick controls
OSCdef(\modRhyKickCurve, { |msg|
	~modRhy.kickCurve = msg[1].asFloat;
	~modRhy.generateKickDist.value;
	"[mod_rhy] Kick curve: %".format(~modRhy.kickCurve).postln;
}, '/pattern/mod_rhy/kick/curve');

OSCdef(\modRhyKickEvents, { |msg|
	~modRhy.kickEvents = msg[1].asInteger;
	~modRhy.generateKickDist.value;
	"[mod_rhy] Kick events: %".format(~modRhy.kickEvents).postln;
}, '/pattern/mod_rhy/kick/events');

OSCdef(\modRhyKickSubdiv, { |msg|
	~modRhy.kickSubdiv = msg[1].asInteger;
	"[mod_rhy] Kick subdivision: %".format(~modRhy.kickSubdiv).postln;
}, '/pattern/mod_rhy/kick/subdivision');

OSCdef(\modRhyKickQuantized, { |msg|
	~modRhy.kickQuantized = msg[1].asInteger == 1;
	"[mod_rhy] Kick quantized: %".format(~modRhy.kickQuantized).postln;
}, '/pattern/mod_rhy/kick/quantized');

// Hihat controls
OSCdef(\modRhyHihatCurve, { |msg|
	~modRhy.hihatCurve = msg[1].asFloat;
	~modRhy.generateHihatDist.value;
	"[mod_rhy] Hihat curve: %".format(~modRhy.hihatCurve).postln;
}, '/pattern/mod_rhy/hihat/curve');

OSCdef(\modRhyHihatEvents, { |msg|
	~modRhy.hihatEvents = msg[1].asInteger;
	~modRhy.generateHihatDist.value;
	"[mod_rhy] Hihat events: %".format(~modRhy.hihatEvents).postln;
}, '/pattern/mod_rhy/hihat/events');

OSCdef(\modRhyHihatSubdiv, { |msg|
	~modRhy.hihatSubdiv = msg[1].asInteger;
	"[mod_rhy] Hihat subdivision: %".format(~modRhy.hihatSubdiv).postln;
}, '/pattern/mod_rhy/hihat/subdivision');

OSCdef(\modRhyHihatQuantized, { |msg|
	~modRhy.hihatQuantized = msg[1].asInteger == 1;
	"[mod_rhy] Hihat quantized: %".format(~modRhy.hihatQuantized).postln;
}, '/pattern/mod_rhy/hihat/quantized');

"[mod_rhy] Pattern loaded and ready. Listening for OSC on port 57120".postln;
"[mod_rhy] Send /pattern/mod_rhy/play to start".postln;
)
