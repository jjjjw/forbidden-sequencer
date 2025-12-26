<script>
	import { sendOSC } from './osc.js';

	// Pattern state
	let baseEventDur = 0.125;
	let phraseLength = 16;
	let phrasesPerSection = 2;
	let rootNote = 53; // F3
	let isPlaying = false;
	let debug = false;

	// Computed
	$: phraseDur = baseEventDur * phraseLength;

	// Playback controls
	function togglePlay() {
		if (isPlaying) {
			sendOSC('/pattern/markov_chord/stop');
			isPlaying = false;
		} else {
			sendOSC('/pattern/markov_chord/play');
			isPlaying = true;
		}
	}

	// Parameter updates
	function updateBaseEventDur() {
		sendOSC('/pattern/markov_chord/base_event_dur', parseFloat(baseEventDur));
	}

	function updatePhraseLength() {
		sendOSC('/pattern/markov_chord/phrase_length', parseInt(phraseLength));
	}

	function updatePhrasesPerSection() {
		sendOSC('/pattern/markov_chord/phrases_per_section', parseInt(phrasesPerSection));
	}

	function updateRootNote() {
		sendOSC('/pattern/markov_chord/root_note', parseInt(rootNote));
	}

	function toggleDebug() {
		debug = !debug;
		sendOSC('/pattern/markov_chord/debug', debug ? 1 : 0);
	}

	// Helper function to get note name from MIDI number
	function getNoteName(midiNote) {
		const noteNames = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'A#', 'B'];
		const octave = Math.floor(midiNote / 12) - 1;
		const noteName = noteNames[midiNote % 12];
		return `${noteName}${octave}`;
	}
</script>

<div class="max-w-4xl mx-auto p-8">
	<h2 class="text-3xl font-bold mb-6">Markov Chord</h2>

	<!-- Playback controls -->
	<div class="flex gap-4 mb-8">
		<button
			on:click={togglePlay}
			class="px-6 py-3 rounded-lg font-semibold transition-colors {isPlaying
				? 'bg-green-500 hover:bg-green-600 text-white'
				: 'bg-gray-200 hover:bg-gray-300 text-gray-800'}"
		>
			{isPlaying ? '■ Stop' : '▶ Play'}
		</button>
		<button
			on:click={toggleDebug}
			class="px-6 py-3 rounded-lg font-semibold transition-colors {debug
				? 'bg-blue-500 hover:bg-blue-600 text-white'
				: 'bg-gray-200 hover:bg-gray-300 text-gray-800'}"
		>
			Debug
		</button>
	</div>

	<!-- Global parameters -->
	<div class="bg-gray-100 rounded-lg p-6 mb-6">
		<h3 class="text-xl font-semibold mb-4 text-gray-700">Global</h3>

		<div class="space-y-4">
			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Base Event Dur: {baseEventDur.toFixed(3)}s</span
					>
					<input
						type="range"
						bind:value={baseEventDur}
						on:input={updateBaseEventDur}
						min="0.025"
						max="1.0"
						step="0.005"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Phrase Length: {phraseLength}</span
					>
					<input
						type="range"
						bind:value={phraseLength}
						on:input={updatePhraseLength}
						min="4"
						max="64"
						step="1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div class="text-sm text-gray-600 mt-2">
				Phrase Duration: {phraseDur.toFixed(2)}s
			</div>
		</div>
	</div>

	<!-- Pattern parameters -->
	<div class="bg-gray-100 rounded-lg p-6">
		<h3 class="text-xl font-semibold mb-4 text-gray-700">Pattern</h3>

		<div class="space-y-4">
			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Root Note: {rootNote} ({getNoteName(rootNote)})</span
					>
					<input
						type="range"
						bind:value={rootNote}
						on:input={updateRootNote}
						min="0"
						max="127"
						step="1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Phrases per Section: {phrasesPerSection}</span
					>
					<input
						type="range"
						bind:value={phrasesPerSection}
						on:input={updatePhrasesPerSection}
						min="1"
						max="16"
						step="1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div class="text-sm text-gray-600 mt-2">
				Alternates between chord and percussion sections every {phrasesPerSection} phrase{phrasesPerSection >
				1
					? 's'
					: ''}
			</div>
		</div>
	</div>
</div>

<style>
	/* Custom slider thumb styling */
	.slider::-webkit-slider-thumb {
		-webkit-appearance: none;
		appearance: none;
		width: 20px;
		height: 20px;
		background: #374151;
		cursor: pointer;
		border-radius: 50%;
	}

	.slider::-moz-range-thumb {
		width: 20px;
		height: 20px;
		background: #374151;
		cursor: pointer;
		border-radius: 50%;
		border: none;
	}
</style>
