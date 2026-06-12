<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { KillJob, listJobs, StartListener } from '../../api/server.js';
  import CustomSelect from '../../components/ui/CustomSelect.svelte';
  import PanelShell from '../../components/ui/PanelShell.svelte';
  import { errorMessage } from '../../utils/errors.js';
  import { startPolling } from '../../utils/polling.js';

  export let embedded = false;
  const dispatch = createEventDispatcher();
  function close() { dispatch('close'); }

  let jobs = [];
  let error = "";
  let loading = true;

  // Start-listener form.
  let proto = "mtls";
  let host = "0.0.0.0";
  let port = 443;
  let domains = "";
  let starting = false;

  const DEFAULT_PORTS = { mtls: 443, http: 80, https: 443, dns: 53 };

  async function refresh() {
    error = "";
    try {
      jobs = await listJobs();
    } catch (e) {
      error = errorMessage(e);
    } finally {
      loading = false;
    }
  }

  function onProtoChange() {
    port = DEFAULT_PORTS[proto] ?? port;
  }

  async function start() {
    starting = true;
    error = "";
    try {
      await StartListener(proto, host, Number(port), domains);
      await refresh();
    } catch (e) {
      error = errorMessage(e);
    } finally {
      starting = false;
    }
  }

  async function kill(id) {
    error = "";
    try {
      await KillJob(id);
      await refresh();
    } catch (e) {
      error = errorMessage(e);
    }
  }

  onMount(() => startPolling(refresh, 3000));

  $: isDNS = proto === 'dns';
</script>

<PanelShell title="Listeners & Jobs" icon="fa-headphones" bodyPadding="16px 20px" {embedded} on:close={close}>
  <div class="start-row">
    <div style="width: 100px;">
      <CustomSelect id="listener-protocol" bind:value={proto} options={['mtls', 'http', 'https', 'dns']} on:change={onProtoChange} />
    </div>
    <input class="host" aria-label="Listener host" type="text" bind:value={host} placeholder="Host" />
    <input class="port" aria-label="Listener port" type="number" bind:value={port} placeholder="Port" />
    {#if isDNS}
      <input class="domains" aria-label="DNS domains" type="text" bind:value={domains} placeholder="domains (comma-separated)" />
    {/if}
    <button class="btn btn-primary" on:click={start} disabled={starting}>
      {starting ? 'Starting…' : 'Start Listener'}
    </button>
  </div>

  {#if error}<div class="error">{error}</div>{/if}

  <table class="data-table">
    <thead>
      <tr><th>ID</th><th>Name</th><th>Protocol</th><th>Port</th><th>Description</th><th></th></tr>
    </thead>
    <tbody>
      {#if loading}
        <tr><td colspan="6" class="muted">Loading…</td></tr>
      {:else if jobs.length === 0}
        <tr><td colspan="6" class="muted">No active jobs.</td></tr>
      {:else}
        {#each jobs as j}
          <tr>
            <td class="mono">{j.ID ?? j.id}</td>
            <td>{j.Name ?? j.name}</td>
            <td>{j.Protocol ?? j.protocol}</td>
            <td class="mono">{j.Port ?? j.port}</td>
            <td>{j.Description ?? j.description}</td>
            <td><button class="btn btn-danger" on:click={() => kill(j.ID ?? j.id)}>Kill</button></td>
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>
</PanelShell>

<style>
  .start-row { display: flex; gap: 8px; margin-bottom: 14px; flex-wrap: wrap; }
  .start-row input {
    background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color);
    padding: 6px 10px; border-radius: 3px; font-family: var(--font-mono); outline: none;
  }
  .start-row .host { flex: 1; min-width: 120px; }
  .start-row .port { width: 90px; }
  .start-row .domains { flex: 1; min-width: 160px; }
  .error { color: var(--danger-color); margin-bottom: 12px; font-size: 0.9em; }
  .muted { color: var(--text-muted); text-align: center; padding: 14px; }
  .btn-danger { background: var(--danger-color); color: #fff; border: none; padding: 3px 10px; border-radius: 3px; cursor: pointer; font-size: 0.85em; }
</style>
