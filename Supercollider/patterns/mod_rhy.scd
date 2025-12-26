// Modulated Rhythm Pattern - Routine-based implementation
// Controls two rhythmic patterns (kick and hihat) with ritardando distribution
// Receives OSC control messages from Go TUI on port 57120

(
// Global state dictionary
~modRhy = ~modRhy ?? ();

// Pattern state
~modRhy.phraseDur = 2.0; // phrase duration in seconds
~modRhy.debugMode = false; // debug logging toggle

// Kick state
~modRhy.kickCurve = 1.5;
~modRhy.kickEvents = 8;
~modRhy.kickOffset = 0; // offset in event positions (active)
~modRhy.kickOffsetBuffer = nil; // buffered offset (pending)
~modRhy.kickPhraseEvents = 16; // number of event positions for kick (active)
~modRhy.kickPhraseEventsBuffer = nil; // buffered phrase events for kick (pending)
~modRhy.kickDurations = nil; // array of durations between events (active)
~modRhy.kickDurationsBuffer = nil; // buffered durations (pending)

// Hihat state
~modRhy.hihatCurve = 1.5;
~modRhy.hihatEvents = 8;
~modRhy.hihatOffset = 0; // offset in event positions (active)
~modRhy.hihatOffsetBuffer = nil; // buffered offset (pending)
~modRhy.hihatPhraseEvents = 16; // number of event positions for hihat (active)
~modRhy.hihatPhraseEventsBuffer = nil; // buffered phrase events for hihat (pending)
~modRhy.hihatDurations = nil; // array of durations between events (active)
~modRhy.hihatDurationsBuffer = nil; // buffered durations (pending)

// Duration generator for kick - writes to buffer
~modRhy.generateKickDurations = {
	var numEvents, positions;

	numEvents = ~modRhy.kickPhraseEventsBuffer ?? ~modRhy.kickPhraseEvents;
	positions = Array.new(numEvents);

	// Calculate event positions using curve
	numEvents.do { |i|
		var t, curved, pos;
		t = i / (numEvents - 1); // 0.0 to 1.0
		curved = t.pow(~modRhy.kickCurve);
		pos = curved * ~modRhy.phraseDur;
		positions.add(pos);
	};

	// Convert positions to durations (time between events)
	~modRhy.kickDurationsBuffer = Array.new(numEvents);
	numEvents.do { |i|
		var dur;
		dur = if(i < (numEvents - 1), {
			positions[i + 1] - positions[i]
		}, {
			~modRhy.phraseDur - positions[i] // last event to end of phrase
		});
		~modRhy.kickDurationsBuffer.add(dur);
	};

	if(~modRhy.debugMode, {
		"[mod_rhy] Kick durations buffered: %".format(~modRhy.kickDurationsBuffer).postln;
	});
};

// Duration generator for hihat - writes to buffer
~modRhy.generateHihatDurations = {
	var numEvents, positions;

	numEvents = ~modRhy.hihatPhraseEventsBuffer ?? ~modRhy.hihatPhraseEvents;
	positions = Array.new(numEvents);

	// Calculate event positions using curve
	numEvents.do { |i|
		var t, curved, pos;
		t = i / (numEvents - 1); // 0.0 to 1.0
		curved = t.pow(~modRhy.hihatCurve);
		pos = curved * ~modRhy.phraseDur;
		positions.add(pos);
	};

	// Convert positions to durations (time between events)
	~modRhy.hihatDurationsBuffer = Array.new(numEvents);
	numEvents.do { |i|
		var dur;
		dur = if(i < (numEvents - 1), {
			positions[i + 1] - positions[i]
		}, {
			~modRhy.phraseDur - positions[i] // last event to end of phrase
		});
		~modRhy.hihatDurationsBuffer.add(dur);
	};

	if(~modRhy.debugMode, {
		"[mod_rhy] Hihat durations buffered: %".format(~modRhy.hihatDurationsBuffer).postln;
	});
};

// Kick task - loops through phrase positions
~modRhy.kickTask = Task({
	var pos = 0, dur, offsetPos;

	inf.do {
		// At phrase start (pos 0), apply buffered parameters if available
		if(pos == 0, {
			if(~modRhy.kickPhraseEventsBuffer.notNil, {
				~modRhy.kickPhraseEvents = ~modRhy.kickPhraseEventsBuffer;
				~modRhy.kickPhraseEventsBuffer = nil;
				if(~modRhy.debugMode, {
					"[mod_rhy] Kick phrase events applied: %".format(~modRhy.kickPhraseEvents).postln;
				});
			});
			if(~modRhy.kickOffsetBuffer.notNil, {
				~modRhy.kickOffset = ~modRhy.kickOffsetBuffer;
				~modRhy.kickOffsetBuffer = nil;
				if(~modRhy.debugMode, {
					"[mod_rhy] Kick offset applied: %".format(~modRhy.kickOffset).postln;
				});
			});
			if(~modRhy.kickDurationsBuffer.notNil, {
				~modRhy.kickDurations = ~modRhy.kickDurationsBuffer;
				~modRhy.kickDurationsBuffer = nil;
				if(~modRhy.debugMode, {
					"[mod_rhy] Kick durations applied".postln;
				});
			});
		});

		// Get duration for this position
		dur = ~modRhy.kickDurations[pos];

		// Calculate position relative to offset
		offsetPos = (pos - ~modRhy.kickOffset + ~modRhy.kickPhraseEvents) % ~modRhy.kickPhraseEvents;

		// Fire synth if this position is within the active event window
		if(offsetPos < ~modRhy.kickEvents, {
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

		// Next position, wrap at kickPhraseEvents
		pos = (pos + 1) % ~modRhy.kickPhraseEvents;
	};
}, SystemClock);

// Hihat task - loops through phrase positions
~modRhy.hihatTask = Task({
	var pos = 0, dur, offsetPos;

	inf.do {
		// At phrase start (pos 0), apply buffered parameters if available
		if(pos == 0, {
			if(~modRhy.hihatPhraseEventsBuffer.notNil, {
				~modRhy.hihatPhraseEvents = ~modRhy.hihatPhraseEventsBuffer;
				~modRhy.hihatPhraseEventsBuffer = nil;
				if(~modRhy.debugMode, {
					"[mod_rhy] Hihat phrase events applied: %".format(~modRhy.hihatPhraseEvents).postln;
				});
			});
			if(~modRhy.hihatOffsetBuffer.notNil, {
				~modRhy.hihatOffset = ~modRhy.hihatOffsetBuffer;
				~modRhy.hihatOffsetBuffer = nil;
				if(~modRhy.debugMode, {
					"[mod_rhy] Hihat offset applied: %".format(~modRhy.hihatOffset).postln;
				});
			});
			if(~modRhy.hihatDurationsBuffer.notNil, {
				~modRhy.hihatDurations = ~modRhy.hihatDurationsBuffer;
				~modRhy.hihatDurationsBuffer = nil;
				if(~modRhy.debugMode, {
					"[mod_rhy] Hihat durations applied".postln;
				});
			});
		});

		// Get duration for this position
		dur = ~modRhy.hihatDurations[pos];

		// Calculate position relative to offset
		offsetPos = (pos - ~modRhy.hihatOffset + ~modRhy.hihatPhraseEvents) % ~modRhy.hihatPhraseEvents;

		// Fire synth if this position is within the active event window
		if(offsetPos < ~modRhy.hihatEvents, {
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

		// Next position, wrap at hihatPhraseEvents
		pos = (pos + 1) % ~modRhy.hihatPhraseEvents;
	};
}, SystemClock);

// OSC Responders for control from Go TUI

// Play/Pause/Resume/Stop
OSCdef(\modRhyPlay, {
	"[mod_rhy] Playing (from start)".postln;
	~modRhy.kickTask.reset;
	~modRhy.hihatTask.reset;
	~modRhy.kickTask.start;
	~modRhy.hihatTask.start;
}, '/pattern/mod_rhy/play');

OSCdef(\modRhyPause, {
	"[mod_rhy] Paused".postln;
	~modRhy.kickTask.pause;
	~modRhy.hihatTask.pause;
}, '/pattern/mod_rhy/pause');

OSCdef(\modRhyResume, {
	"[mod_rhy] Resumed".postln;
	~modRhy.kickTask.resume;
	~modRhy.hihatTask.resume;
}, '/pattern/mod_rhy/resume');

OSCdef(\modRhyStop, {
	"[mod_rhy] Stopped (reset to start)".postln;
	~modRhy.kickTask.stop;
	~modRhy.hihatTask.stop;
	~modRhy.kickTask.reset;
	~modRhy.hihatTask.reset;
}, '/pattern/mod_rhy/stop');

// Reset to defaults
OSCdef(\modRhyReset, {
	// Stop tasks
	~modRhy.kickTask.stop;
	~modRhy.hihatTask.stop;

	// Reset pattern state
	~modRhy.phraseDur = 2.0;
	~modRhy.debugMode = false;

	// Reset kick state
	~modRhy.kickCurve = 1.5;
	~modRhy.kickEvents = 8;
	~modRhy.kickOffset = 0;
	~modRhy.kickOffsetBuffer = nil;
	~modRhy.kickPhraseEvents = 16;
	~modRhy.kickPhraseEventsBuffer = nil;
	~modRhy.kickDurationsBuffer = nil;

	// Reset hihat state
	~modRhy.hihatCurve = 1.5;
	~modRhy.hihatEvents = 8;
	~modRhy.hihatOffset = 0;
	~modRhy.hihatOffsetBuffer = nil;
	~modRhy.hihatPhraseEvents = 16;
	~modRhy.hihatPhraseEventsBuffer = nil;
	~modRhy.hihatDurationsBuffer = nil;

	// Regenerate and apply default durations
	~modRhy.generateKickDurations.value;
	~modRhy.kickDurations = ~modRhy.kickDurationsBuffer;
	~modRhy.kickDurationsBuffer = nil;

	~modRhy.generateHihatDurations.value;
	~modRhy.hihatDurations = ~modRhy.hihatDurationsBuffer;
	~modRhy.hihatDurationsBuffer = nil;

	"[mod_rhy] Reset".postln;
}, '/pattern/mod_rhy/reset');

// Phrase duration control
OSCdef(\modRhyPhraseDur, { |msg|
	~modRhy.phraseDur = msg[1].asFloat;
	// Regenerate both duration buffers when phrase duration changes
	~modRhy.generateKickDurations.value;
	~modRhy.generateHihatDurations.value;
	if(~modRhy.debugMode, {
		"[mod_rhy] Phrase duration: %s".format(~modRhy.phraseDur).postln;
	});
}, '/pattern/mod_rhy/phrase_dur');

// Phrase events control
OSCdef(\modRhyPhraseEvents, { |msg|
	~modRhy.kickPhraseEventsBuffer = msg[1].asInteger;
	~modRhy.hihatPhraseEventsBuffer = msg[1].asInteger;
	// Regenerate both duration buffers when phrase events changes
	~modRhy.generateKickDurations.value;
	~modRhy.generateHihatDurations.value;
	if(~modRhy.debugMode, {
		"[mod_rhy] Phrase events: %".format(msg[1].asInteger).postln;
	});
}, '/pattern/mod_rhy/phrase_events');

// Debug toggle
OSCdef(\modRhyDebug, { |msg|
	~modRhy.debugMode = msg[1].asInteger == 1;
	"[mod_rhy] Debug: %".format(~modRhy.debugMode).postln;
}, '/pattern/mod_rhy/debug');

// Kick controls
OSCdef(\modRhyKickCurve, { |msg|
	~modRhy.kickCurve = msg[1].asFloat;
	~modRhy.generateKickDurations.value;
	if(~modRhy.debugMode, {
		"[mod_rhy] Kick curve: %".format(~modRhy.kickCurve).postln;
	});
}, '/pattern/mod_rhy/kick/curve');

OSCdef(\modRhyKickEvents, { |msg|
	~modRhy.kickEvents = msg[1].asInteger;
	if(~modRhy.debugMode, {
		"[mod_rhy] Kick events: %".format(~modRhy.kickEvents).postln;
	});
}, '/pattern/mod_rhy/kick/events');

OSCdef(\modRhyKickOffset, { |msg|
	~modRhy.kickOffsetBuffer = msg[1].asInteger;
	if(~modRhy.debugMode, {
		"[mod_rhy] Kick offset: %".format(~modRhy.kickOffsetBuffer).postln;
	});
}, '/pattern/mod_rhy/kick/offset');

// Hihat controls
OSCdef(\modRhyHihatCurve, { |msg|
	~modRhy.hihatCurve = msg[1].asFloat;
	~modRhy.generateHihatDurations.value;
	if(~modRhy.debugMode, {
		"[mod_rhy] Hihat curve: %".format(~modRhy.hihatCurve).postln;
	});
}, '/pattern/mod_rhy/hihat/curve');

OSCdef(\modRhyHihatEvents, { |msg|
	~modRhy.hihatEvents = msg[1].asInteger;
	if(~modRhy.debugMode, {
		"[mod_rhy] Hihat events: %".format(~modRhy.hihatEvents).postln;
	});
}, '/pattern/mod_rhy/hihat/events');

OSCdef(\modRhyHihatOffset, { |msg|
	~modRhy.hihatOffsetBuffer = msg[1].asInteger;
	if(~modRhy.debugMode, {
		"[mod_rhy] Hihat offset: %".format(~modRhy.hihatOffsetBuffer).postln;
	});
}, '/pattern/mod_rhy/hihat/offset');

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
