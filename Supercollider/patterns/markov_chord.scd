// Markov Chord Pattern - Single task implementation with section switching
// Alternates between chord section (4-voice FM chord) and percussion section (kick/snare/hihat)
// All notes in a chord are triggered within one bind callback
// Receives OSC control messages from Go TUI on port 57120

(
// Load Markov chain library
(thisProcess.nowExecutingPath.dirname.dirname +/+ "lib/markov.scd").load;

// Global state dictionary
~markovChord = ~markovChord ?? ();

// Pattern state
~markovChord.baseEventDur = 0.125; // duration of each event slot in seconds
~markovChord.phraseLength = 16; // number of events in phrase
~markovChord.phrasesPerSection = 2; // phrases before switching sections
~markovChord.debugMode = false;

// Section state
~markovChord.currentSection = \chord; // \chord or \percussion
~markovChord.phraseCounter = 0;
~markovChord.tickInPhrase = 0;
~markovChord.chordPlayed = false; // track if chord was played this section

// Melodic state (melodic minor scale: 0, 2, 3, 5, 7, 9, 11)
~markovChord.melodicMinor = [0, 2, 3, 5, 7, 9, 11];
~markovChord.rootNote = 53; // F3

// Markov chains for percussion
~markovChord.kickChain = ~newMarkovChain.value(42);
~markovChord.snareChain = ~newMarkovChain.value(84);
~markovChord.hihatChain = ~newMarkovChain.value(126);

// Initialize Markov chains with probabilities
~markovChord.kickChain.setTransition(\playing, \playing, 0.6);
~markovChord.kickChain.setTransition(\playing, \silent, 0.4);
~markovChord.kickChain.setTransition(\silent, \silent, 0.4);
~markovChord.kickChain.setTransition(\silent, \playing, 0.6);

~markovChord.snareChain.setTransition(\playing, \playing, 0.4);
~markovChord.snareChain.setTransition(\playing, \silent, 0.6);
~markovChord.snareChain.setTransition(\silent, \silent, 0.6);
~markovChord.snareChain.setTransition(\silent, \playing, 0.4);

~markovChord.hihatChain.setTransition(\playing, \playing, 0.8);
~markovChord.hihatChain.setTransition(\playing, \silent, 0.2);
~markovChord.hihatChain.setTransition(\silent, \silent, 0.2);
~markovChord.hihatChain.setTransition(\silent, \playing, 0.8);

// Main task - combines all pattern logic
~markovChord.mainTask = Task({
	var eventDur, synthLen, chordDegrees, modIndices, midiNote, phraseDur, state;

	chordDegrees = [0, 2, 4, 6]; // I, iii, V, vii in melodic minor
	modIndices = [0.3, 0.4, 0.35, 0.38];

	inf.do {
		// Use baseEventDur directly (since phraseDur = baseEventDur * phraseLength)
		eventDur = ~markovChord.baseEventDur;
		synthLen = eventDur * 0.75; // 75% of event duration
		phraseDur = ~markovChord.baseEventDur * ~markovChord.phraseLength;

		// At phrase start, check if we should switch sections
		if(~markovChord.tickInPhrase == 0, {
			~markovChord.phraseCounter = ~markovChord.phraseCounter + 1;

			if(~markovChord.phraseCounter >= ~markovChord.phrasesPerSection, {
				// Switch sections
				if(~markovChord.currentSection == \chord, {
					~markovChord.currentSection = \percussion;
					if(~markovChord.debugMode, {
						"[markov_chord] Switched to percussion section".postln;
					});
				}, {
					~markovChord.currentSection = \chord;
					~markovChord.chordPlayed = false; // reset for new chord section
					if(~markovChord.debugMode, {
						"[markov_chord] Switched to chord section".postln;
					});
				});

				~markovChord.phraseCounter = 0;
			});
		});

		// Handle section-specific playback
		if(~markovChord.currentSection == \chord, {
			// Only play chord once per section, at phrase start
			if((~markovChord.tickInPhrase == 0) and: { ~markovChord.chordPlayed.not }, {
				~markovChord.chordPlayed = true;

				// Play 4-note chord - all notes in one bind callback
				s.bind {
					chordDegrees.do { |degree, i|
						midiNote = ~markovChord.rootNote + ~markovChord.melodicMinor[degree];
						Synth(\fm2op, [
							\freq, midiNote.midicps,
							\amp, 0.6 / 4, // divide by 4 for 4 voices
							\modRatio, 1.0, // unison for smooth warm tone
							\modIndex, modIndices[i],
							\out, 10, // reverb bus
							\len, phraseDur // full phrase duration
						], target: 100);
					};
				};
			});
		}, {
			// Percussion section - play Markov-based drums
			state = ~markovChord.kickChain.nextState();
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

			state = ~markovChord.snareChain.nextState();
			if(state == \playing, {
				s.bind {
					Synth(\cp, [
						\amp, 0.7,
						\len, synthLen,
						\out, 10 // reverb bus
					], target: 100);
				};
			});

			state = ~markovChord.hihatChain.nextState();
			if(state == \playing, {
				s.bind {
					Synth(\hh, [
						\amp, 0.6,
						\len, synthLen,
						\out, 0
					], target: 100);
				};
			});
		});

		// Yield and advance tick
		eventDur.yield;
		~markovChord.tickInPhrase = (~markovChord.tickInPhrase + 1) % ~markovChord.phraseLength;
	};
}, SystemClock);

// OSC Responders

// Play/Pause/Resume/Stop
OSCdef(\markovChordPlay, {
	"[markov_chord] Playing (from start)".postln;
	~markovChord.currentSection = \chord;
	~markovChord.phraseCounter = 0;
	~markovChord.tickInPhrase = 0;
	~markovChord.chordPlayed = false;
	~markovChord.kickChain.resetState();
	~markovChord.snareChain.resetState();
	~markovChord.hihatChain.resetState();
	~markovChord.mainTask.reset;
	~markovChord.mainTask.start;
}, '/pattern/markov_chord/play');

OSCdef(\markovChordPause, {
	"[markov_chord] Paused".postln;
	~markovChord.mainTask.pause;
}, '/pattern/markov_chord/pause');

OSCdef(\markovChordResume, {
	"[markov_chord] Resumed".postln;
	~markovChord.mainTask.resume;
}, '/pattern/markov_chord/resume');

OSCdef(\markovChordStop, {
	"[markov_chord] Stopped (reset to start)".postln;
	~markovChord.mainTask.stop;
	~markovChord.mainTask.reset;
	~markovChord.currentSection = \chord;
	~markovChord.phraseCounter = 0;
	~markovChord.tickInPhrase = 0;
	~markovChord.chordPlayed = false;
}, '/pattern/markov_chord/stop');

// Reset to defaults
OSCdef(\markovChordReset, {
	// Stop task
	~markovChord.mainTask.stop;

	// Reset state
	~markovChord.baseEventDur = 0.125;
	~markovChord.phraseLength = 16;
	~markovChord.phrasesPerSection = 2;
	~markovChord.debugMode = false;
	~markovChord.currentSection = \chord;
	~markovChord.phraseCounter = 0;
	~markovChord.tickInPhrase = 0;
	~markovChord.chordPlayed = false;
	~markovChord.rootNote = 53;

	// Reset Markov chains
	~markovChord.kickChain.resetState();
	~markovChord.snareChain.resetState();
	~markovChord.hihatChain.resetState();

	"[markov_chord] Reset".postln;
}, '/pattern/markov_chord/reset');

// Base event duration control
OSCdef(\markovChordBaseEventDur, { |msg|
	var phraseDur;
	~markovChord.baseEventDur = msg[1].asFloat;
	if(~markovChord.debugMode, {
		phraseDur = ~markovChord.baseEventDur * ~markovChord.phraseLength;
		"[markov_chord] Base event dur: %s, Phrase dur: %s".format(~markovChord.baseEventDur, phraseDur).postln;
	});
}, '/pattern/markov_chord/base_event_dur');

// Phrase length control
OSCdef(\markovChordPhraseLength, { |msg|
	var phraseDur;
	~markovChord.phraseLength = msg[1].asInteger;
	if(~markovChord.debugMode, {
		phraseDur = ~markovChord.baseEventDur * ~markovChord.phraseLength;
		"[markov_chord] Phrase length: %, Phrase dur: %s".format(~markovChord.phraseLength, phraseDur).postln;
	});
}, '/pattern/markov_chord/phrase_length');

// Phrases per section control
OSCdef(\markovChordPhrasesPerSection, { |msg|
	~markovChord.phrasesPerSection = msg[1].asInteger;
	if(~markovChord.debugMode, {
		"[markov_chord] Phrases per section: %".format(~markovChord.phrasesPerSection).postln;
	});
}, '/pattern/markov_chord/phrases_per_section');

// Root note control
OSCdef(\markovChordRootNote, { |msg|
	~markovChord.rootNote = msg[1].asInteger;
	if(~markovChord.debugMode, {
		"[markov_chord] Root note: %".format(~markovChord.rootNote).postln;
	});
}, '/pattern/markov_chord/root_note');

// Debug toggle
OSCdef(\markovChordDebug, { |msg|
	~markovChord.debugMode = msg[1].asInteger == 1;
	"[markov_chord] Debug: %".format(~markovChord.debugMode).postln;
}, '/pattern/markov_chord/debug');

"[markov_chord] Pattern loaded and ready. Listening for OSC on port 57120".postln;
"[markov_chord] Send /pattern/markov_chord/play to start".postln;
)
