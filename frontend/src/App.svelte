<script lang="ts">
  import Settings from './components/Settings.svelte';
  import { Stop, Play } from '../wailsjs/go/main/App';

  enum View {
    Main = 'main',
    Settings = 'settings'
  }

  let stopped = true;
  let currentView: View = View.Main;

  async function togglePlayback() {
    try {
      if (stopped) {
        await Play();
        stopped = false;
      } else {
        await Stop();
        stopped = true;
      }
    } catch (err) {
      console.error('Failed to toggle playback:', err);
    }
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
        <button class="px-8 py-2 text-lg cursor-pointer" on:click={togglePlayback}>
          {stopped ? 'Play' : 'Stop'}
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
