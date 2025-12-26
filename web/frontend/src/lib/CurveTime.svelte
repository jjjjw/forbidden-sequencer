<script>
	import { sendOSC } from './osc.js';

	// Pattern state
	let baseEventDur = 0.125;
	let phraseEvents = 16;
	let isPlaying = false;
	let debug = false;

	// Kick state
	let kickCurve = 1.5;
	let kickEvents = 8;
	let kickOffset = 0;

	// Hihat state
	let hihatCurve = 1.5;
	let hihatEvents = 8;
	let hihatOffset = 0;

	// Computed
	$: phraseDur = baseEventDur * phraseEvents;

	// Playback controls
	function togglePlay() {
		if (isPlaying) {
			sendOSC('/pattern/curve_time/stop');
			isPlaying = false;
		} else {
			sendOSC('/pattern/curve_time/play');
			isPlaying = true;
		}
	}

	// Parameter updates
	function updateBaseEventDur() {
		sendOSC('/pattern/curve_time/base_event_dur', parseFloat(baseEventDur));
	}

	function updatePhraseEvents() {
		sendOSC('/pattern/curve_time/phrase_events', parseInt(phraseEvents));
	}

	function updateKickCurve() {
		sendOSC('/pattern/curve_time/kick/curve', parseFloat(kickCurve));
	}

	function updateKickEvents() {
		sendOSC('/pattern/curve_time/kick/events', parseInt(kickEvents));
	}

	function updateKickOffset() {
		sendOSC('/pattern/curve_time/kick/offset', parseInt(kickOffset));
	}

	function updateHihatCurve() {
		sendOSC('/pattern/curve_time/hihat/curve', parseFloat(hihatCurve));
	}

	function updateHihatEvents() {
		sendOSC('/pattern/curve_time/hihat/events', parseInt(hihatEvents));
	}

	function updateHihatOffset() {
		sendOSC('/pattern/curve_time/hihat/offset', parseInt(hihatOffset));
	}

	function toggleDebug() {
		debug = !debug;
		sendOSC('/pattern/curve_time/debug', debug ? 1 : 0);
	}
</script>

<div class="max-w-4xl mx-auto p-8">
	<h2 class="text-3xl font-bold mb-6">Curve Time</h2>

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
						>Phrase Events: {phraseEvents}</span
					>
					<input
						type="range"
						bind:value={phraseEvents}
						on:input={updatePhraseEvents}
						min="16"
						max="32"
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

	<!-- Kick controls -->
	<div class="bg-gray-100 rounded-lg p-6 mb-6">
		<h3 class="text-xl font-semibold mb-4 text-gray-700">Kick</h3>

		<div class="space-y-4">
			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Curve: {kickCurve.toFixed(1)}</span
					>
					<input
						type="range"
						bind:value={kickCurve}
						on:input={updateKickCurve}
						min="0.5"
						max="2.0"
						step="0.1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block">Events: {kickEvents}</span>
					<input
						type="range"
						bind:value={kickEvents}
						on:input={updateKickEvents}
						min="1"
						max="16"
						step="1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Offset: {kickOffset > 0 ? '+' : ''}{kickOffset}</span
					>
					<input
						type="range"
						bind:value={kickOffset}
						on:input={updateKickOffset}
						min={-phraseEvents + 1}
						max={phraseEvents - 1}
						step="1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>
		</div>
	</div>

	<!-- Hihat controls -->
	<div class="bg-gray-100 rounded-lg p-6">
		<h3 class="text-xl font-semibold mb-4 text-gray-700">Hihat</h3>

		<div class="space-y-4">
			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Curve: {hihatCurve.toFixed(1)}</span
					>
					<input
						type="range"
						bind:value={hihatCurve}
						on:input={updateHihatCurve}
						min="0.5"
						max="2.0"
						step="0.1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block">Events: {hihatEvents}</span>
					<input
						type="range"
						bind:value={hihatEvents}
						on:input={updateHihatEvents}
						min="1"
						max="16"
						step="1"
						class="w-full h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer slider"
					/>
				</label>
			</div>

			<div>
				<label class="block">
					<span class="text-sm font-medium text-gray-700 mb-2 block"
						>Offset: {hihatOffset > 0 ? '+' : ''}{hihatOffset}</span
					>
					<input
						type="range"
						bind:value={hihatOffset}
						on:input={updateHihatOffset}
						min={-phraseEvents + 1}
						max={phraseEvents - 1}
						step="1"
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
