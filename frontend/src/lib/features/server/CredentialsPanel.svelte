<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { AddCredential, listCredentials, RemoveCredential } from '../../api/server.js';
  import PanelShell from '../../components/ui/PanelShell.svelte';
  import { dialog } from '../../stores/dialogStore.js';
  import { errorMessage } from '../../utils/errors.js';

  export let embedded = false;
  const dispatch = createEventDispatcher();
  function close() { dispatch('close'); }

  let creds = [];
  let error = "";
  let loading = true;
  let filter = "";

  async function refresh() {
    error = "";
    try {
      creds = await listCredentials();
    } catch (e) {
      error = errorMessage(e);
    } finally {
      loading = false;
    }
  }

  $: shown = creds.filter((c) => {
    if (!filter) return true;
    const hay = `${c.Username || c.username || ''} ${c.Plaintext || c.plaintext || ''} ${c.Hash || c.hash || ''}`.toLowerCase();
    return hay.includes(filter.toLowerCase());
  });

  // --- admin: add / delete ---
  let showAdd = false;
  let nu = "", np = "", nh = "", ncol = "";
  async function add() {
    error = "";
    try {
      await AddCredential(nu, np, nh, ncol);
      nu = np = nh = ncol = "";
      showAdd = false;
      await refresh();
    } catch (e) { error = errorMessage(e); }
  }
  async function remove(id) {
    if (!id || !(await dialog.confirm('Delete this credential?', 'Confirm Delete'))) return;
    error = "";
    try { await RemoveCredential(id); await refresh(); }
    catch (e) { error = errorMessage(e); }
  }

  onMount(refresh);
</script>

<PanelShell title="Credentials" icon="fa-key" {embedded} on:close={close}>
  <svelte:fragment slot="actions">
    <input class="filter" aria-label="Filter credentials" placeholder="filter…" bind:value={filter} />
    <button class="btn btn-primary" on:click={() => showAdd = !showAdd}><i class="fas fa-plus"></i> Add</button>
    <button class="btn" on:click={refresh}>Refresh</button>
  </svelte:fragment>

  {#if error}<div class="error">{error}</div>{/if}
  {#if showAdd}
    <div class="add-row">
      <input aria-label="Username" placeholder="username" bind:value={nu} />
      <input aria-label="Plaintext" placeholder="plaintext" bind:value={np} />
      <input aria-label="Hash" placeholder="hash" bind:value={nh} />
      <input aria-label="Collection" placeholder="collection" bind:value={ncol} />
      <button class="btn btn-primary" on:click={add}>Save</button>
      <button class="btn" on:click={() => showAdd = false}>Cancel</button>
    </div>
  {/if}
  <table class="data-table">
    <thead><tr><th>Username</th><th>Plaintext</th><th>Hash</th><th>Type</th><th>Collection</th><th></th></tr></thead>
    <tbody>
      {#if loading}
        <tr><td colspan="6" class="muted">Loading…</td></tr>
      {:else if shown.length === 0}
        <tr><td colspan="6" class="muted">No credentials.</td></tr>
      {:else}
        {#each shown as c}
          <tr>
            <td>{c.Username || c.username || '-'}</td>
            <td class="mono">{c.Plaintext || c.plaintext || '-'}</td>
            <td class="mono hash">{c.Hash || c.hash || '-'}</td>
            <td>{c.HashType || c.hashType || '-'}</td>
            <td>{c.Collection || c.collection || '-'}</td>
            <td><button class="btn btn-danger" on:click={() => remove(c.ID || c.id)}>Delete</button></td>
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>
</PanelShell>

<style>
  .filter { background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color); padding: 5px 10px; border-radius: 3px; font-family: var(--font-mono); outline: none; }
  .error { color: var(--danger-color); margin: 12px 0; font-size: 0.9em; }
  .add-row { display: flex; gap: 8px; margin-bottom: 12px; flex-wrap: wrap; }
  .add-row input { background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color); padding: 6px 10px; border-radius: 3px; outline: none; flex: 1; min-width: 120px; }
  .muted { color: var(--text-muted); text-align: center; padding: 14px; }
  .hash { max-width: 280px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
