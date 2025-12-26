// Markov Triggers Pattern - Task-based implementation with Markov chains
// Controls 5 voices: kick, snare, hihat, fm1, fm2
// Receives OSC control messages from Go TUI on port 57120

(
// Load Markov chain library
(thisProcess.nowExecutingPath.dirname.dirname +/+ "lib/markov.scd").load;

// Global state dictionary
~markovTrig = ~markovTrig ?? ();

// Pattern state
~markovTrig.baseEventDur = 0.125; // duration of each event slot in seconds
~markovTrig.phraseLength = 16; // number of events in phrase
~markovTrig.debugMode = false;

// Markov chains for each voice
~markovTrig.kickChain = ~newMarkovChain.value(42);
~markovTrig.snareChain = ~newMarkovChain.value(55);
~markovTrig.hihatChain = ~newMarkovChain.value(43);

// Probability state (0.0 to 1.0)
~markovTrig.kickProb = 0.5;
~markovTrig.snareProb = 0.5;
~markovTrig.hihatProb = 0.5;
~markovTrig.fm1Prob = 0.3;
~markovTrig.fm2Prob = 0.3;

// Snare trigger state
~markovTrig.tickInPhrase = 0;
~markovTrig.snareTriggerTick = 12; // snare triggers at tick 12 in phrase
~markovTrig.willSnareTrigger = false;
~markovTrig.snareTriggerChecked = false;

// FM state (melodic minor scale: 0, 2, 3, 5, 7, 9, 11)
~markovTrig.melodicMinor = [0, 2, 3, 5, 7, 9, 11];
~markovTrig.fm1Root = 60; // C4
~markovTrig.fm2Root = 72; // C5
~markovTrig.fm1Octaves = 2;
~markovTrig.fm2Octaves = 2;

// Initialize Markov chains with default probabilities
~markovTrig.updateKickChain = {
	~markovTrig.kickChain.setTransition(\playing, \playing, ~markovTrig.kickProb);
	~markovTrig.kickChain.setTransition(\playing, \silent, 1.0 - ~markovTrig.kickProb);
	~markovTrig.kickChain.setTransition(\silent, \silent, 1.0 - ~markovTrig.kickProb);
	~markovTrig.kickChain.setTransition(\silent, \playing, ~markovTrig.kickProb);
};

~markovTrig.updateSnareChain = {
	~markovTrig.snareChain.setTransition(\trigger, \trigger, ~markovTrig.snarePro);
	~markovTrig.snareChain.setTransition(\trigger, \no_trigger, 1.0 - ~markovTrig.snarePro);
	~markovTrig.snareChain.setTransition(\no_trigger, \no_trigger, 1.0 - ~markovTrig.snareProb);
	~markovTrig.snareChain.setTransition(\no_trigger, \trigger, ~markovTrig.snarePro);
};

~markovTrig.updateHihatChain = {
	~markovTrig.hihatChain.setTransition(\playing, \playing, ~markovTrig.hihatPro);
	~markovTrig.hihatChain.setTransition(\playing, \silent, 1.0 - ~markovTrig.hihatPro);
	~markovTrig.hihatChain.setTransition(\silent, \silent, 1.0 - ~markovTrig.hihatProb);
	~markovTrig.hihatChain.setTransition(\silent, \playing, ~markovTrig.hihatPro);
};

// Initialize chains
~markovTrig.updateKickChain.value;
~markovTrig.updateSnareChain.value;
~markovTrig.updateHihatChain.value;

// Single main task - handles timing and triggers all voices
~markovTrig.mainTask = Task({
	var eventDur, synthLen, state;
	var scaleLength, maxDegree, degree, midiNote, durationTicks;
	var ratios, modRatio, modIndex;

	ratios = [0.5, 1.0, 1.5, 2.0, 3.0, 4.0, 5.0, 7.0];

	inf.do {
		// Use baseEventDur directly (since phraseDur = baseEventDur * phraseLength)
		eventDur = ~markovTrig.baseEventDur;
		synthLen = eventDur * 0.75; // 75% of event duration
		// === SNARE ===
		// At phrase start, decide if snare will trigger this phrase
		if(~markovTrig.tickInPhrase == 0, {
			state = ~markovTrig.snareChain.nextState();
			~markovTrig.willSnareTrigger = (state == \trigger);
			~markovTrig.snareTriggerChecked = true;

			if(~markovTrig.debugMode, {
				"[markov_trig] Snare decision: %".format(~markovTrig.willSnareTrigger).postln;
			});
		});

		// Fire snare at trigger tick if decided to trigger
		if(~markovTrig.tickInPhrase == ~markovTrig.snareTriggerTick, {
			if(~markovTrig.willSnareTrigger, {
				s.bind {
					Synth(\cp, [
						\amp, 0.7,
						\len, synthLen,
						\out, 10 // reverb bus
					], target: 100);
				};
			});
		});

		// === KICK ===
		// If snare will trigger and we're at or past the trigger tick, stay silent
		if(~markovTrig.willSnareTrigger and: { ~markovTrig.tickInPhrase >= ~markovTrig.snareTriggerTick }, {
			// Silent
			if(~markovTrig.debugMode, {
				"[markov_trig] Kick: silent (snare active)".postln;
			});
		}, {
			// Use Markov chain to decide
			state = ~markovTrig.kickChain.nextState();

			if(~markovTrig.debugMode, {
				"[markov_trig] Kick: state=%".format(state).postln;
			});

			if(state == \playing, {
				s.bind {
					Synth(\bd, [
						\freq, 50,
						\amp, 0.8,
						\len, synthLen,
						\out, 0
					], target: 100);
				};
			});
		});

		// === HIHAT ===
		state = ~markovTrig.hihatChain.nextState();

		if(~markovTrig.debugMode, {
			"[markov_trig] Hihat: state=%".format(state).postln;
		});

		if(state == \playing, {
  		if(~markovTrig.debugMode, {
  			"[markov_trig] Hihat: playing".postln;
  		});
			s.bind {
				Synth(\hh, [
					\amp, 0.6,
					\len, synthLen,
					\out, 0
				], target: 100);
			};
		});

		// === FM1 ===
		if(1.0.rand < ~markovTrig.fm1Prob, {
			// Random pitch from scale
			scaleLength = ~markovTrig.melodicMinor.size;
			maxDegree = scaleLength * ~markovTrig.fm1Octaves;
			degree = maxDegree.rand;
			midiNote = ~markovTrig.fm1Root + ~markovTrig.melodicMinor[degree % scaleLength] + ((degree / scaleLength).floor * 12);

			// Random duration (0.75 to 2.0 ticks)
			durationTicks = 0.75 + (1.25.rand);

			// Random FM parameters
			modRatio = ratios.choose;
			modIndex = 0.1 + (2.9.rand);

			s.bind {
				Synth(\fm2op, [
					\midi_note, midiNote,
					\amp, 0.5,
					\modRatio, modRatio,
					\modIndex, modIndex,
					\out, 10, // reverb bus
					\dur, durationTicks * synthLen
				], target: 100);
			};
		});

		// === FM2 ===
		if(1.0.rand < ~markovTrig.fm2Prob, {
			// Random pitch from scale
			scaleLength = ~markovTrig.melodicMinor.size;
			maxDegree = scaleLength * ~markovTrig.fm2Octaves;
			degree = maxDegree.rand;
			midiNote = ~markovTrig.fm2Root + ~markovTrig.melodicMinor[degree % scaleLength] + ((degree / scaleLength).floor * 12);

			// Random duration (0.75 to 2.0 ticks)
			durationTicks = 0.75 + (1.25.rand);

			// Random FM parameters
			modRatio = ratios.choose;
			modIndex = 0.1 + (2.9.rand);

			s.bind {
				Synth(\fm2op, [
					\midi_note, midiNote,
					\amp, 0.4,
					\modRatio, modRatio,
					\modIndex, modIndex,
					\out, 10, // reverb bus
					\dur, durationTicks * synthLen
				], target: 100);
			};
		});

		// Yield and advance tick
		eventDur.yield;
		~markovTrig.tickInPhrase = (~markovTrig.tickInPhrase + 1) % ~markovTrig.phraseLength;
	};
}, SystemClock);

// OSC Responders

// Play/Pause/Resume/Stop
OSCdef(\markovTrigPlay, {
	"[markov_trig] Playing (from start)".postln;
	~markovTrig.tickInPhrase = 0;
	~markovTrig.willSnareTrigger = false;
	~markovTrig.snareTriggerChecked = false;
	~markovTrig.kickChain.resetState();
	~markovTrig.snareChain.resetState();
	~markovTrig.hihatChain.resetState();
	// Set initial states to ensure chains start active
	~markovTrig.kickChain.setState(\playing);
	~markovTrig.snareChain.setState(\trigger);
	~markovTrig.hihatChain.setState(\playing);
	~markovTrig.mainTask.reset;
	~markovTrig.mainTask.start;
}, '/pattern/markov_trig/play');

OSCdef(\markovTrigPause, {
	"[markov_trig] Paused".postln;
	~markovTrig.mainTask.pause;
}, '/pattern/markov_trig/pause');

OSCdef(\markovTrigResume, {
	"[markov_trig] Resumed".postln;
	~markovTrig.mainTask.resume;
}, '/pattern/markov_trig/resume');

OSCdef(\markovTrigStop, {
	"[markov_trig] Stopped (reset to start)".postln;
	~markovTrig.mainTask.stop;
	~markovTrig.mainTask.reset;
	~markovTrig.tickInPhrase = 0;
	~markovTrig.willSnareTrigger = false;
	~markovTrig.snareTriggerChecked = false;
}, '/pattern/markov_trig/stop');

// Reset to defaults
OSCdef(\markovTrigReset, {
	// Stop task
	~markovTrig.mainTask.stop;

	// Reset state
	~markovTrig.baseEventDur = 0.125;
	~markovTrig.phraseLength = 16;
	~markovTrig.debugMode = false;
	~markovTrig.kickProb = 0.5;
	~markovTrig.snareProb = 0.5;
	~markovTrig.hihatProb = 0.5;
	~markovTrig.fm1Prob = 0.3;
	~markovTrig.fm2Prob = 0.3;
	~markovTrig.tickInPhrase = 0;
	~markovTrig.willSnareTrigger = false;
	~markovTrig.snareTriggerChecked = false;

	// Rebuild chains with default probabilities
	~markovTrig.updateKickChain.value;
	~markovTrig.updateSnareChain.value;
	~markovTrig.updateHihatChain.value;
	~markovTrig.kickChain.resetState();
	~markovTrig.snareChain.resetState();
	~markovTrig.hihatChain.resetState();

	"[markov_trig] Reset".postln;
}, '/pattern/markov_trig/reset');

// Base event duration control
OSCdef(\markovTrigBaseEventDur, { |msg|
	var phraseDur;
	~markovTrig.baseEventDur = msg[1].asFloat;
	if(~markovTrig.debugMode, {
		phraseDur = ~markovTrig.baseEventDur * ~markovTrig.phraseLength;
		"[markov_trig] Base event dur: %s, Phrase dur: %s".format(~markovTrig.baseEventDur, phraseDur).postln;
	});
}, '/pattern/markov_trig/base_event_dur');

// Phrase length control
OSCdef(\markovTrigPhraseLength, { |msg|
	var phraseDur;
	~markovTrig.phraseLength = msg[1].asInteger;
	if(~markovTrig.debugMode, {
		phraseDur = ~markovTrig.baseEventDur * ~markovTrig.phraseLength;
		"[markov_trig] Phrase length: %, Phrase dur: %s".format(~markovTrig.phraseLength, phraseDur).postln;
	});
}, '/pattern/markov_trig/phrase_length');

// Probability controls
OSCdef(\markovTrigKickProb, { |msg|
	~markovTrig.kickProb = msg[1].asFloat.clip(0.0, 1.0);
	~markovTrig.updateKickChain.value;
	if(~markovTrig.debugMode, {
		"[markov_trig] Kick probability: %".format(~markovTrig.kickProb).postln;
	});
}, '/pattern/markov_trig/kick/prob');

OSCdef(\markovTrigSnareProb, { |msg|
	~markovTrig.snareProb = msg[1].asFloat.clip(0.0, 1.0);
	~markovTrig.updateSnareChain.value;
	if(~markovTrig.debugMode, {
		"[markov_trig] Snare probability: %".format(~markovTrig.snareProb).postln;
	});
}, '/pattern/markov_trig/snare/prob');

OSCdef(\markovTrigHihatProb, { |msg|
	~markovTrig.hihatProb = msg[1].asFloat.clip(0.0, 1.0);
	~markovTrig.updateHihatChain.value;
	if(~markovTrig.debugMode, {
		"[markov_trig] Hihat probability: %".format(~markovTrig.hihatProb).postln;
	});
}, '/pattern/markov_trig/hihat/prob');

OSCdef(\markovTrigFm1Prob, { |msg|
	~markovTrig.fm1Prob = msg[1].asFloat.clip(0.0, 1.0);
	if(~markovTrig.debugMode, {
		"[markov_trig] FM1 probability: %".format(~markovTrig.fm1Prob).postln;
	});
}, '/pattern/markov_trig/fm1/prob');

OSCdef(\markovTrigFm2Prob, { |msg|
	~markovTrig.fm2Prob = msg[1].asFloat.clip(0.0, 1.0);
	if(~markovTrig.debugMode, {
		"[markov_trig] FM2 probability: %".format(~markovTrig.fm2Prob).postln;
	});
}, '/pattern/markov_trig/fm2/prob');

// Debug toggle
OSCdef(\markovTrigDebug, { |msg|
	~markovTrig.debugMode = msg[1].asInteger == 1;
	"[markov_trig] Debug: %".format(~markovTrig.debugMode).postln;
}, '/pattern/markov_trig/debug');

"[markov_trig] Pattern loaded and ready. Listening for OSC on port 57120".postln;
"[markov_trig] Send /pattern/markov_trig/play to start".postln;
)
