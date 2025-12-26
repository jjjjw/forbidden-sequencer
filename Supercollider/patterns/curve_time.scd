// Curve Time Pattern - Routine-based implementation
// Controls two rhythmic patterns (kick and hihat) with ritardando distribution
// Receives OSC control messages from Go TUI on port 57120

(
// Global state dictionary
~curveTime = ~curveTime ?? ();

// Pattern state
~curveTime.baseEventDur = 0.125; // duration of each event slot in seconds
~curveTime.phraseEvents = 16; // number of event positions in phrase (set globally, not per-voice)
~curveTime.debugMode = false; // debug logging toggle

// Kick state
~curveTime.kickCurve = 1.5;
~curveTime.kickEvents = 8;
~curveTime.kickOffset = 0; // offset in event positions (active)
~curveTime.kickOffsetBuffer = nil; // buffered offset (pending)
~curveTime.kickPhraseEvents = 16; // number of event positions for kick (active)
~curveTime.kickPhraseEventsBuffer = nil; // buffered phrase events for kick (pending)
~curveTime.kickDurations = nil; // array of durations between events (active)
~curveTime.kickDurationsBuffer = nil; // buffered durations (pending)

// Hihat state
~curveTime.hihatCurve = 1.5;
~curveTime.hihatEvents = 8;
~curveTime.hihatOffset = 0; // offset in event positions (active)
~curveTime.hihatOffsetBuffer = nil; // buffered offset (pending)
~curveTime.hihatPhraseEvents = 16; // number of event positions for hihat (active)
~curveTime.hihatPhraseEventsBuffer = nil; // buffered phrase events for hihat (pending)
~curveTime.hihatDurations = nil; // array of durations between events (active)
~curveTime.hihatDurationsBuffer = nil; // buffered durations (pending)

// Duration generator for kick - writes to buffer
~curveTime.generateKickDurations = {
	var numEvents, positions, phraseDur;

	// Recalculate phraseDur from base values
	phraseDur = ~curveTime.baseEventDur * ~curveTime.phraseEvents;

	numEvents = ~curveTime.kickPhraseEventsBuffer ?? ~curveTime.kickPhraseEvents;
	positions = Array.new(numEvents);

	// Calculate event positions using curve
	numEvents.do { |i|
		var t, curved, pos;
		t = i / (numEvents - 1); // 0.0 to 1.0
		curved = t.pow(~curveTime.kickCurve);
		pos = curved * phraseDur;
		positions.add(pos);
	};

	// Convert positions to durations (time between events)
	~curveTime.kickDurationsBuffer = Array.new(numEvents);
	numEvents.do { |i|
		var dur;
		dur = if(i < (numEvents - 1), {
			positions[i + 1] - positions[i]
		}, {
			phraseDur - positions[i] // last event to end of phrase
		});
		~curveTime.kickDurationsBuffer.add(dur);
	};

	if(~curveTime.debugMode, {
		"[curve_time] Kick durations buffered: %".format(~curveTime.kickDurationsBuffer).postln;
	});
};

// Duration generator for hihat - writes to buffer
~curveTime.generateHihatDurations = {
	var numEvents, positions, phraseDur;

	// Recalculate phraseDur from base values
	phraseDur = ~curveTime.baseEventDur * ~curveTime.phraseEvents;

	numEvents = ~curveTime.hihatPhraseEventsBuffer ?? ~curveTime.hihatPhraseEvents;
	positions = Array.new(numEvents);

	// Calculate event positions using curve
	numEvents.do { |i|
		var t, curved, pos;
		t = i / (numEvents - 1); // 0.0 to 1.0
		curved = t.pow(~curveTime.hihatCurve);
		pos = curved * phraseDur;
		positions.add(pos);
	};

	// Convert positions to durations (time between events)
	~curveTime.hihatDurationsBuffer = Array.new(numEvents);
	numEvents.do { |i|
		var dur;
		dur = if(i < (numEvents - 1), {
			positions[i + 1] - positions[i]
		}, {
			phraseDur - positions[i] // last event to end of phrase
		});
		~curveTime.hihatDurationsBuffer.add(dur);
	};

	if(~curveTime.debugMode, {
		"[curve_time] Hihat durations buffered: %".format(~curveTime.hihatDurationsBuffer).postln;
	});
};

// Kick task - loops through phrase positions
~curveTime.kickTask = Task({
	var pos = 0, dur, offsetPos;

	inf.do {
		// At phrase start (pos 0), apply buffered parameters if available
		if(pos == 0, {
			if(~curveTime.kickPhraseEventsBuffer.notNil, {
				~curveTime.kickPhraseEvents = ~curveTime.kickPhraseEventsBuffer;
				~curveTime.kickPhraseEventsBuffer = nil;
				if(~curveTime.debugMode, {
					"[curve_time] Kick phrase events applied: %".format(~curveTime.kickPhraseEvents).postln;
				});
			});
			if(~curveTime.kickOffsetBuffer.notNil, {
				~curveTime.kickOffset = ~curveTime.kickOffsetBuffer;
				~curveTime.kickOffsetBuffer = nil;
				if(~curveTime.debugMode, {
					"[curve_time] Kick offset applied: %".format(~curveTime.kickOffset).postln;
				});
			});
			if(~curveTime.kickDurationsBuffer.notNil, {
				~curveTime.kickDurations = ~curveTime.kickDurationsBuffer;
				~curveTime.kickDurationsBuffer = nil;
				if(~curveTime.debugMode, {
					"[curve_time] Kick durations applied".postln;
				});
			});
		});

		// Get duration for this position
		dur = ~curveTime.kickDurations[pos];

		// Calculate position relative to offset
		offsetPos = (pos - ~curveTime.kickOffset + ~curveTime.kickPhraseEvents) % ~curveTime.kickPhraseEvents;

		// Fire synth if this position is within the active event window
		if(offsetPos < ~curveTime.kickEvents, {
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
		pos = (pos + 1) % ~curveTime.kickPhraseEvents;
	};
}, SystemClock);

// Hihat task - loops through phrase positions
~curveTime.hihatTask = Task({
	var pos = 0, dur, offsetPos;

	inf.do {
		// At phrase start (pos 0), apply buffered parameters if available
		if(pos == 0, {
			if(~curveTime.hihatPhraseEventsBuffer.notNil, {
				~curveTime.hihatPhraseEvents = ~curveTime.hihatPhraseEventsBuffer;
				~curveTime.hihatPhraseEventsBuffer = nil;
				if(~curveTime.debugMode, {
					"[curve_time] Hihat phrase events applied: %".format(~curveTime.hihatPhraseEvents).postln;
				});
			});
			if(~curveTime.hihatOffsetBuffer.notNil, {
				~curveTime.hihatOffset = ~curveTime.hihatOffsetBuffer;
				~curveTime.hihatOffsetBuffer = nil;
				if(~curveTime.debugMode, {
					"[curve_time] Hihat offset applied: %".format(~curveTime.hihatOffset).postln;
				});
			});
			if(~curveTime.hihatDurationsBuffer.notNil, {
				~curveTime.hihatDurations = ~curveTime.hihatDurationsBuffer;
				~curveTime.hihatDurationsBuffer = nil;
				if(~curveTime.debugMode, {
					"[curve_time] Hihat durations applied".postln;
				});
			});
		});

		// Get duration for this position
		dur = ~curveTime.hihatDurations[pos];

		// Calculate position relative to offset
		offsetPos = (pos - ~curveTime.hihatOffset + ~curveTime.hihatPhraseEvents) % ~curveTime.hihatPhraseEvents;

		// Fire synth if this position is within the active event window
		if(offsetPos < ~curveTime.hihatEvents, {
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
		pos = (pos + 1) % ~curveTime.hihatPhraseEvents;
	};
}, SystemClock);

// OSC Responders for control from Go TUI

// Play/Pause/Resume/Stop
OSCdef(\curveTimePlay, {
	"[curve_time] Playing (from start)".postln;
	~curveTime.kickTask.reset;
	~curveTime.hihatTask.reset;
	~curveTime.kickTask.start;
	~curveTime.hihatTask.start;
}, '/pattern/curve_time/play');

OSCdef(\curveTimePause, {
	"[curve_time] Paused".postln;
	~curveTime.kickTask.pause;
	~curveTime.hihatTask.pause;
}, '/pattern/curve_time/pause');

OSCdef(\curveTimeResume, {
	"[curve_time] Resumed".postln;
	~curveTime.kickTask.resume;
	~curveTime.hihatTask.resume;
}, '/pattern/curve_time/resume');

OSCdef(\curveTimeStop, {
	"[curve_time] Stopped (reset to start)".postln;
	~curveTime.kickTask.stop;
	~curveTime.hihatTask.stop;
	~curveTime.kickTask.reset;
	~curveTime.hihatTask.reset;
}, '/pattern/curve_time/stop');

// Reset to defaults
OSCdef(\curveTimeReset, {
	// Stop tasks
	~curveTime.kickTask.stop;
	~curveTime.hihatTask.stop;
	~curveTime.kickTask.reset;
	~curveTime.hihatTask.reset;

	// Reset pattern state
	~curveTime.baseEventDur = 0.125;
	~curveTime.phraseEvents = 16;
	~curveTime.debugMode = false;

	// Reset kick state
	~curveTime.kickCurve = 1.5;
	~curveTime.kickEvents = 8;
	~curveTime.kickOffset = 0;
	~curveTime.kickOffsetBuffer = nil;
	~curveTime.kickPhraseEvents = 16;
	~curveTime.kickPhraseEventsBuffer = nil;
	~curveTime.kickDurationsBuffer = nil;

	// Reset hihat state
	~curveTime.hihatCurve = 1.5;
	~curveTime.hihatEvents = 8;
	~curveTime.hihatOffset = 0;
	~curveTime.hihatOffsetBuffer = nil;
	~curveTime.hihatPhraseEvents = 16;
	~curveTime.hihatPhraseEventsBuffer = nil;
	~curveTime.hihatDurationsBuffer = nil;

	// Regenerate and apply default durations
	~curveTime.generateKickDurations.value;
	~curveTime.kickDurations = ~curveTime.kickDurationsBuffer;
	~curveTime.kickDurationsBuffer = nil;

	~curveTime.generateHihatDurations.value;
	~curveTime.hihatDurations = ~curveTime.hihatDurationsBuffer;
	~curveTime.hihatDurationsBuffer = nil;

	"[curve_time] Reset".postln;
}, '/pattern/curve_time/reset');

// Base event duration control
OSCdef(\curveTimeBaseEventDur, { |msg|
	var phraseDur;
	~curveTime.baseEventDur = msg[1].asFloat;
	// Regenerate both duration buffers when base event duration changes
	~curveTime.generateKickDurations.value;
	~curveTime.generateHihatDurations.value;
	if(~curveTime.debugMode, {
		phraseDur = ~curveTime.baseEventDur * ~curveTime.phraseEvents;
		"[curve_time] Base event dur: %s, Phrase dur: %s".format(~curveTime.baseEventDur, phraseDur).postln;
	});
}, '/pattern/curve_time/base_event_dur');

// Phrase events control
OSCdef(\curveTimePhraseEvents, { |msg|
	var phraseDur;
	~curveTime.phraseEvents = msg[1].asInteger;
	~curveTime.kickPhraseEventsBuffer = ~curveTime.phraseEvents;
	~curveTime.hihatPhraseEventsBuffer = ~curveTime.phraseEvents;
	// Regenerate both duration buffers when phrase events changes
	~curveTime.generateKickDurations.value;
	~curveTime.generateHihatDurations.value;
	if(~curveTime.debugMode, {
		phraseDur = ~curveTime.baseEventDur * ~curveTime.phraseEvents;
		"[curve_time] Phrase events: %, Phrase dur: %s".format(~curveTime.phraseEvents, phraseDur).postln;
	});
}, '/pattern/curve_time/phrase_events');

// Debug toggle
OSCdef(\curveTimeDebug, { |msg|
	~curveTime.debugMode = msg[1].asInteger == 1;
	"[curve_time] Debug: %".format(~curveTime.debugMode).postln;
}, '/pattern/curve_time/debug');

// Kick controls
OSCdef(\curveTimeKickCurve, { |msg|
	~curveTime.kickCurve = msg[1].asFloat;
	~curveTime.generateKickDurations.value;
	if(~curveTime.debugMode, {
		"[curve_time] Kick curve: %".format(~curveTime.kickCurve).postln;
	});
}, '/pattern/curve_time/kick/curve');

OSCdef(\curveTimeKickEvents, { |msg|
	~curveTime.kickEvents = msg[1].asInteger;
	if(~curveTime.debugMode, {
		"[curve_time] Kick events: %".format(~curveTime.kickEvents).postln;
	});
}, '/pattern/curve_time/kick/events');

OSCdef(\curveTimeKickOffset, { |msg|
	~curveTime.kickOffsetBuffer = msg[1].asInteger;
	if(~curveTime.debugMode, {
		"[curve_time] Kick offset: %".format(~curveTime.kickOffsetBuffer).postln;
	});
}, '/pattern/curve_time/kick/offset');

// Hihat controls
OSCdef(\curveTimeHihatCurve, { |msg|
	~curveTime.hihatCurve = msg[1].asFloat;
	~curveTime.generateHihatDurations.value;
	if(~curveTime.debugMode, {
		"[curve_time] Hihat curve: %".format(~curveTime.hihatCurve).postln;
	});
}, '/pattern/curve_time/hihat/curve');

OSCdef(\curveTimeHihatEvents, { |msg|
	~curveTime.hihatEvents = msg[1].asInteger;
	if(~curveTime.debugMode, {
		"[curve_time] Hihat events: %".format(~curveTime.hihatEvents).postln;
	});
}, '/pattern/curve_time/hihat/events');

OSCdef(\curveTimeHihatOffset, { |msg|
	~curveTime.hihatOffsetBuffer = msg[1].asInteger;
	if(~curveTime.debugMode, {
		"[curve_time] Hihat offset: %".format(~curveTime.hihatOffsetBuffer).postln;
	});
}, '/pattern/curve_time/hihat/offset');

// Initialize durations (directly to active arrays for first time)
~curveTime.generateKickDurations.value;
~curveTime.kickDurations = ~curveTime.kickDurationsBuffer;
~curveTime.kickDurationsBuffer = nil;

~curveTime.generateHihatDurations.value;
~curveTime.hihatDurations = ~curveTime.hihatDurationsBuffer;
~curveTime.hihatDurationsBuffer = nil;

"[curve_time] Pattern loaded and ready. Listening for OSC on port 57120".postln;
"[curve_time] Send /pattern/curve_time/play to start".postln;
)
