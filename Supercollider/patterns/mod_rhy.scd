// Modulated Rhythm Pattern - Routine-based implementation
// Controls two rhythmic patterns (kick and hihat) with ritardando distribution
// Receives OSC control messages from Go TUI on port 57120

(
// Global state dictionary
~modRhy = ~modRhy ?? ();

// Pattern state
~modRhy.phraseDur = 2.0; // phrase duration in seconds

// Kick state
~modRhy.kickCurve = 2.0;
~modRhy.kickEvents = 8;
~modRhy.kickDurations = nil; // array of 16 durations between events (active)
~modRhy.kickDurationsBuffer = nil; // buffered durations (pending)

// Hihat state
~modRhy.hihatCurve = 1.5;
~modRhy.hihatEvents = 8;
~modRhy.hihatDurations = nil; // array of 16 durations between events (active)
~modRhy.hihatDurationsBuffer = nil; // buffered durations (pending)

// Duration generator for kick (16 positions) - writes to buffer
~modRhy.generateKickDurations = {
	var positions = Array.new(16);

	// Calculate 16 event positions using curve
	16.do { |i|
		var t, curved, pos;
		t = i / 15; // 0.0 to 1.0
		curved = t.pow(~modRhy.kickCurve);
		pos = curved * ~modRhy.phraseDur;
		positions.add(pos);
	};

	// Convert positions to durations (time between events)
	~modRhy.kickDurationsBuffer = Array.new(16);
	16.do { |i|
		var dur;
		dur = if(i < 15, {
			positions[i + 1] - positions[i]
		}, {
			~modRhy.phraseDur - positions[i] // last event to end of phrase
		});
		~modRhy.kickDurationsBuffer.add(dur);
	};

	"[mod_rhy] Kick durations buffered: %".format(~modRhy.kickDurationsBuffer).postln;
};

// Duration generator for hihat (16 positions) - writes to buffer
~modRhy.generateHihatDurations = {
	var positions = Array.new(16);

	// Calculate 16 event positions using curve
	16.do { |i|
		var t, curved, pos;
		t = i / 15; // 0.0 to 1.0
		curved = t.pow(~modRhy.hihatCurve);
		pos = curved * ~modRhy.phraseDur;
		positions.add(pos);
	};

	// Convert positions to durations (time between events)
	~modRhy.hihatDurationsBuffer = Array.new(16);
	16.do { |i|
		var dur;
		dur = if(i < 15, {
			positions[i + 1] - positions[i]
		}, {
			~modRhy.phraseDur - positions[i] // last event to end of phrase
		});
		~modRhy.hihatDurationsBuffer.add(dur);
	};

	"[mod_rhy] Hihat durations buffered: %".format(~modRhy.hihatDurationsBuffer).postln;
};

// Kick task - loops through 16 positions
~modRhy.kickTask = Task({
	var pos = 0, dur;

	inf.do {
		// At phrase start, apply buffered durations if available
		if((pos == 0) && ~modRhy.kickDurationsBuffer.notNil, {
			~modRhy.kickDurations = ~modRhy.kickDurationsBuffer;
			~modRhy.kickDurationsBuffer = nil;
			"[mod_rhy] Kick durations applied".postln;
		});

		// Get duration for this position
		dur = ~modRhy.kickDurations[pos];

		// Fire synth if this position is active
		if(pos < ~modRhy.kickEvents, {
			s.bind {
				Synth(\bd, [
					\freq, 50,
					\amp, 0.8,
					\len, dur * 0.75,
					\out, 0
				], target: 100);
			};
		});

		// Yield duration for this position
		dur.yield;

		// Next position, wrap at 16
		pos = (pos + 1) % 16;
	};
}, SystemClock);

// Hihat task - loops through 16 positions
~modRhy.hihatTask = Task({
	var pos = 0, dur;

	inf.do {
		// At phrase start, apply buffered durations if available
		if((pos == 0) && ~modRhy.hihatDurationsBuffer.notNil, {
			~modRhy.hihatDurations = ~modRhy.hihatDurationsBuffer;
			~modRhy.hihatDurationsBuffer = nil;
			"[mod_rhy] Hihat durations applied".postln;
		});

		// Get duration for this position
		dur = ~modRhy.hihatDurations[pos];

		// Fire synth if this position is active
		if(pos < ~modRhy.hihatEvents, {
			s.bind {
				Synth(\hh, [
					\amp, 0.6,
					\len, dur * 0.75,
					\out, 0
				], target: 100);
			};
		});

		// Yield duration for this position
		dur.yield;

		// Next position, wrap at 16
		pos = (pos + 1) % 16;
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

// Phrase duration control
OSCdef(\modRhyPhraseDur, { |msg|
	~modRhy.phraseDur = msg[1].asFloat;
	// Regenerate both duration buffers when phrase duration changes
	~modRhy.generateKickDurations.value;
	~modRhy.generateHihatDurations.value;
	"[mod_rhy] Phrase duration: %s".format(~modRhy.phraseDur).postln;
}, '/pattern/mod_rhy/phrase_dur');

// Kick controls
OSCdef(\modRhyKickCurve, { |msg|
	~modRhy.kickCurve = msg[1].asFloat;
	~modRhy.generateKickDurations.value;
	"[mod_rhy] Kick curve: %".format(~modRhy.kickCurve).postln;
}, '/pattern/mod_rhy/kick/curve');

OSCdef(\modRhyKickEvents, { |msg|
	~modRhy.kickEvents = msg[1].asInteger;
	"[mod_rhy] Kick events: %".format(~modRhy.kickEvents).postln;
}, '/pattern/mod_rhy/kick/events');

// Hihat controls
OSCdef(\modRhyHihatCurve, { |msg|
	~modRhy.hihatCurve = msg[1].asFloat;
	~modRhy.generateHihatDurations.value;
	"[mod_rhy] Hihat curve: %".format(~modRhy.hihatCurve).postln;
}, '/pattern/mod_rhy/hihat/curve');

OSCdef(\modRhyHihatEvents, { |msg|
	~modRhy.hihatEvents = msg[1].asInteger;
	"[mod_rhy] Hihat events: %".format(~modRhy.hihatEvents).postln;
}, '/pattern/mod_rhy/hihat/events');

// Initialize durations (directly to active arrays for first time)
~modRhy.generateKickDurations.value;
~modRhy.kickDurations = ~modRhy.kickDurationsBuffer;
~modRhy.kickDurationsBuffer = nil;

~modRhy.generateHihatDurations.value;
~modRhy.hihatDurations = ~modRhy.hihatDurationsBuffer;
~modRhy.hihatDurationsBuffer = nil;

"[mod_rhy] Pattern loaded and ready. Listening for OSC on port 57120".postln;
"[mod_rhy] Send /pattern/mod_rhy/play to start".postln;
)
