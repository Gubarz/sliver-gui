<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { DownloadLoot, listLoot, RemoveLoot } from '../../api/server.js';
  import PanelShell from '../../components/ui/PanelShell.svelte';
  import { dialog } from '../../stores/dialogStore.js';
  import { errorMessage } from '../../utils/errors.js';

  export let embedded = false;
  const dispatch = createEventDispatcher();
  function close() { dispatch('close'); }

  let loot = [];
  let error = "";
  let loading = true;

  async function refresh() {
    error = "";
    try {
      loot = await listLoot();
    } catch (e) {
      error = errorMessage(e);
    } finally {
      loading = false;
    }
  }

  async function download(id) {
    error = "";
    try {
      await DownloadLoot(id);
    } catch (e) {
      error = errorMessage(e);
    }
  }

  async function remove(id) {
    if (!(await dialog.confirm('Delete this loot item?', 'Confirm Delete'))) return;
    error = "";
    try {
      await RemoveLoot(id);
      await refresh();
    } catch (e) {
      error = errorMessage(e);
    }
  }

  onMount(refresh);
</script>

<PanelShell title="Loot" icon="fa-download" width="640px" {embedded} on:close={close}>
  <svelte:fragment slot="actions">
    <button class="btn" on:click={refresh}>Refresh</button>
  </svelte:fragment>
  {#if error}<div class="error">{error}</div>{/if}
  <table class="data-table">
    <thead><tr><th>ID</th><th>Name</th><th>File Type</th><th></th></tr></thead>
    <tbody>
      {#if loading}
        <tr><td colspan="4" class="muted">Loading…</td></tr>
      {:else if loot.length === 0}
        <tr><td colspan="4" class="muted">No loot stored.</td></tr>
      {:else}
        {#each loot as l}
          <tr>
            <td class="mono">{(l.ID || l.id || '').substring(0, 8)}</td>
            <td>{l.Name || l.name}</td>
            <td>{l.FileType || l.fileType || '-'}</td>
            <td class="actions">
              <button class="btn" on:click={() => download(l.ID || l.id)}>Download</button>
              <button class="btn btn-danger" on:click={() => remove(l.ID || l.id)}>Delete</button>
            </td>
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>
</PanelShell>

<style>
  .error { color: var(--danger-color); margin: 12px 0; font-size: 0.9em; }
  .muted { color: var(--text-muted); text-align: center; padding: 14px; }
  .actions { display: flex; gap: 6px; }
</style>
