<script lang="ts">
  import { onMount } from 'svelte';
  import { GetMIDIPorts, GetCurrentMIDIPort, SetMIDIPort, GetChannelMappings, SetChannelMapping, RemoveChannelMapping } from '../../wailsjs/go/main/App';

  interface MIDIPort {
    Index: number;
    Name: string;
  }

  let availablePorts: MIDIPort[] = [];
  let currentPort = -1;
  let channelMappings: Record<string, number> = {};
  let newEventName = '';
  let newChannel = 0;
  let loading = true;
  let error = '';

  onMount(async () => {
    await loadSettings();
  });

  async function loadSettings() {
    try {
      loading = true;
      error = '';

      availablePorts = await GetMIDIPorts();
      currentPort = await GetCurrentMIDIPort();
      channelMappings = await GetChannelMappings();

      loading = false;
    } catch (err) {
      error = `Failed to load settings: ${err}`;
      loading = false;
    }
  }

  async function handlePortChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    const portIndex = parseInt(target.value);

    try {
      error = '';
      await SetMIDIPort(portIndex);
      currentPort = portIndex;
    } catch (err) {
      error = `Failed to set MIDI port: ${err}`;
    }
  }

  async function handleChannelChange(eventName: string, event: Event) {
    const target = event.target as HTMLInputElement;
    const channel = parseInt(target.value);

    try {
      error = '';
      await SetChannelMapping(eventName, channel);
      channelMappings[eventName] = channel;
    } catch (err) {
      error = `Failed to set channel mapping: ${err}`;
    }
  }

  async function addChannelMapping() {
    if (!newEventName.trim()) {
      error = 'Event name cannot be empty';
      return;
    }

    try {
      error = '';
      await SetChannelMapping(newEventName, newChannel);
      channelMappings[newEventName] = newChannel;
      channelMappings = channelMappings;
      newEventName = '';
      newChannel = 0;
    } catch (err) {
      error = `Failed to add channel mapping: ${err}`;
    }
  }

  async function removeChannelMapping(eventName: string) {
    try {
      error = '';
      await RemoveChannelMapping(eventName);
      delete channelMappings[eventName];
      channelMappings = channelMappings;
    } catch (err) {
      error = `Failed to remove channel mapping: ${err}`;
    }
  }
</script>

<div class="p-8 max-w-4xl">
  <h2 class="text-2xl font-bold mb-6">Settings</h2>

  {#if loading}
    <p class="text-gray-600">Loading...</p>
  {:else}
    {#if error}
      <div class="bg-red-100 text-red-700 p-3 rounded mb-4">{error}</div>
    {/if}

    <div class="space-y-6">
      <div>
        <h3 class="text-lg font-semibold mb-4">Adapters</h3>

        <div class="border rounded p-4">
          <h4 class="font-semibold mb-4">MIDI</h4>

          <div class="mb-4">
            <label class="block mb-2">MIDI Bus:</label>
            <select
              class="border rounded px-3 py-2 w-full"
              value={currentPort}
              on:change={handlePortChange}
            >
              {#each availablePorts as port}
                <option value={port.Index}>{port.Name}</option>
              {/each}
            </select>
          </div>

          <div class="border-t pt-4">
            <h5 class="font-semibold mb-3">Channel Mapping</h5>

            {#if Object.keys(channelMappings).length === 0}
              <p class="text-gray-500 text-sm mb-4">No channel mappings</p>
            {:else}
              <div class="space-y-2 mb-4">
                {#each Object.entries(channelMappings) as [eventName, channel]}
                  <div class="flex items-center gap-3">
                    <span class="flex-1">{eventName}</span>
                    <input
                      type="number"
                      min="0"
                      max="15"
                      value={channel}
                      on:change={(e) => handleChannelChange(eventName, e)}
                      class="border rounded px-2 py-1 w-20"
                    />
                    <button
                      class="px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700"
                      on:click={() => removeChannelMapping(eventName)}
                    >
                      Remove
                    </button>
                  </div>
                {/each}
              </div>
            {/if}

            <div class="flex gap-2">
              <input
                type="text"
                placeholder="Event name"
                bind:value={newEventName}
                class="border rounded px-3 py-2 flex-1"
              />
              <input
                type="number"
                min="0"
                max="15"
                bind:value={newChannel}
                class="border rounded px-2 py-2 w-20"
              />
              <button
                class="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
                on:click={addChannelMapping}
              >
                Add
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>
