<script>
  import CurveTime from "./lib/CurveTime.svelte";
  import MarkovTrig from "./lib/MarkovTrig.svelte";
  import MarkovChord from "./lib/MarkovChord.svelte";

  let currentPattern = "curve-time";

  const patterns = [
    { id: "curve-time", name: "Curve Time", component: CurveTime },
    { id: "markov-trig", name: "Markov Triggers", component: MarkovTrig },
    { id: "markov-chord", name: "Markov Chord", component: MarkovChord },
  ];

  $: currentComponent = patterns.find((p) => p.id === currentPattern).component;
</script>

<div class="min-h-screen bg-white">
  <!-- Header -->
  <header class="bg-gray-800 text-white shadow-lg">
    <div class="max-w-7xl mx-auto px-4 py-6">
      <h1 class="text-xl font-bold">Forbidden Sequencer</h1>
    </div>
  </header>

  <!-- Tab Navigation -->
  <nav class="bg-gray-100 border-b border-gray-200">
    <div class="max-w-7xl mx-auto px-4">
      <div class="flex space-x-1">
        {#each patterns as pattern}
          <button
            on:click={() => (currentPattern = pattern.id)}
            class="px-6 py-4 font-medium transition-colors border-b-2 {currentPattern ===
            pattern.id
              ? 'border-blue-500 text-blue-600 bg-white'
              : 'border-transparent text-gray-600 hover:text-gray-800 hover:bg-gray-50'}"
          >
            {pattern.name}
          </button>
        {/each}
      </div>
    </div>
  </nav>

  <!-- Controller Content -->
  <main class="py-8">
    <svelte:component this={currentComponent} />
  </main>

  <!-- Footer -->
  <footer class="bg-gray-100 border-t border-gray-200 mt-12">
    <div class="max-w-7xl mx-auto px-4 py-6 text-center text-sm text-gray-600">
      <p>
        Make sure SuperCollider is running with patterns loaded and the OSC
        bridge is running on port 8080
      </p>
    </div>
  </footer>
</div>
