<script lang="ts">
  import Settings from './components/Settings.svelte';

  enum View {
    Main = 'main',
    Settings = 'settings'
  }

  let paused = true;
  let currentView: View = View.Main;

  function togglePause() {
    paused = !paused;
    // TODO: Call Go backend to pause/resume sequencer
  }

  function showSettings() {
    currentView = View.Settings;
  }

  function showMain() {
    currentView = View.Main;
  }
</script>

<main class="p-8 max-w-7xl mx-auto">
  <div class="flex justify-between items-center mb-8">
    <h1 class="text-3xl font-bold m-0">Forbidden Sequencer</h1>
    <div class="flex gap-2">
      {#if currentView === View.Main}
        <button class="px-8 py-2 text-lg cursor-pointer" on:click={togglePause}>
          {paused ? 'Resume' : 'Pause'}
        </button>
        <button class="px-8 py-2 text-lg cursor-pointer" on:click={showSettings}>Settings</button>
      {:else}
        <button class="px-8 py-2 text-lg cursor-pointer" on:click={showMain}>Back</button>
      {/if}
    </div>
  </div>

  {#if currentView === View.Settings}
    <Settings />
  {:else}
    <!-- TODO: Add pattern selection -->
  {/if}
</main>
