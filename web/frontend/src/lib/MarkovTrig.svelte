<script>
	import { sendOSC } from './osc.js';

	// Pattern state
	let baseEventDur = 0.125;
	let phraseLength = 16;
	let isPlaying = false;
	let debug = false;

	// Voice probabilities
	let kickProb = 0.5;
	let snareProb = 0.5;
	let hihatProb = 0.5;
	let fm1Prob = 0.3;
	let fm2Prob = 0.3;

	// Computed
	$: phraseDur = baseEventDur * phraseLength;

	// Playback controls
	function togglePlay() {
		if (isPlaying) {
			sendOSC('/pattern/markov_trig/stop');
			isPlaying = false;
		} else {
			sendOSC('/pattern/markov_trig/play');
			isPlaying = true;
		}
	}

	// Parameter updates
	function updateBaseEventDur() {
		sendOSC('/pattern/markov_trig/base_event_dur', parseFloat(baseEventDur));
	}

	function updatePhraseLength() {
		sendOSC('/pattern/markov_trig/phrase_length', parseInt(phraseLength));
	}

	function updateKickProb() {
		sendOSC('/pattern/markov_trig/kick/prob', parseFloat(kickProb));
	}

	function updateSnareProb() {
		sendOSC('/pattern/markov_trig/snare/prob', parseFloat(snareProb));
	}

	function updateHihatProb() {
		sendOSC('/pattern/markov_trig/hihat/prob', parseFloat(hihatProb));
	}

	function updateFm1Prob() {
		sendOSC('/pattern/markov_trig/fm1/prob', parseFloat(fm1Prob));
	}

	function updateFm2Prob() {
		sendOSC('/pattern/markov_trig/fm2/prob', parseFloat(fm2Prob));
	}

	function toggleDebug() {
		debug = !debug;
		sendOSC('/pattern/markov_trig/debug', debug ? 1 : 0);
	}
</script>

<div class="max-w-4xl mx-auto p-8">
	<h2 class="text-3xl font-bold mb-6">Markov Triggers</h2>

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

	<!-- Voice probabilities -->
	<div class="bg-gray-100 rounded-lg p-6">
		<h3 class="text-xl font-semibold mb-4 text-gray-700">Voice Probabilities</h3>

		<div class="space-y-4">
			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Kick: {(kickProb * 100).toFixed(0)}%</span
					>
					<input
						type="range"
						bind:value={kickProb}
						on:input={updateKickProb}
						min="0.0"
						max="1.0"
						step="0.1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Snare: {(snareProb * 100).toFixed(0)}%</span
					>
					<input
						type="range"
						bind:value={snareProb}
						on:input={updateSnareProb}
						min="0.0"
						max="1.0"
						step="0.1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Hihat: {(hihatProb * 100).toFixed(0)}%</span
					>
					<input
						type="range"
						bind:value={hihatProb}
						on:input={updateHihatProb}
						min="0.0"
						max="1.0"
						step="0.1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>FM1: {(fm1Prob * 100).toFixed(0)}%</span
					>
					<input
						type="range"
						bind:value={fm1Prob}
						on:input={updateFm1Prob}
						min="0.0"
						max="1.0"
						step="0.1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>FM2: {(fm2Prob * 100).toFixed(0)}%</span
					>
					<input
						type="range"
						bind:value={fm2Prob}
						on:input={updateFm2Prob}
						min="0.0"
						max="1.0"
						step="0.1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
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
