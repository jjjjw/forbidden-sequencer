<script lang="ts">
  let paused = true;

  function togglePause() {
    paused = !paused;
    // TODO: Call Go backend to pause/resume sequencer
  }
</script>

<main>
  <div class="header">
    <h1>Forbidden Sequencer</h1>
    <button on:click={togglePause}>
      {paused ? 'Resume' : 'Pause'}
    </button>
  </div>

  <!-- TODO: Add lanes and pattern selection -->
</main>

<style>
  main {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .header h1 {
    margin: 0;
  }

  .header button {
    padding: 0.5rem 2rem;
    font-size: 1.1rem;
    cursor: pointer;
  }
</style>
