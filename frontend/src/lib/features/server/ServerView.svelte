<script>
  import {
    DeleteImplantBuild,
    DeleteProfile,
    GenerateImplantFromProfile,
    listImplantBuilds,
    listProfiles,
    RegenerateImplant,
  } from '../../api/server.js';
  import ListenersPanel from './ListenersPanel.svelte';
  import LootPanel from './LootPanel.svelte';
  import CredentialsPanel from './CredentialsPanel.svelte';
  import OperatorsPanel from './OperatorsPanel.svelte';
  import EventsLog from './EventsLog.svelte';
  import GeneratePanel from './GeneratePanel.svelte';
  import SessionConsole from '../agents/SessionConsole.svelte';
  import { dialog } from '../../stores/dialogStore.js';
  import { errorMessage } from '../../utils/errors.js';

  export let eventHistory = [];

  const tabs = [
    { id: 'console', label: 'Server Console', icon: 'fa-terminal' },
    { id: 'listeners', label: 'Listeners & Jobs', icon: 'fa-headphones' },
    { id: 'implants', label: 'Implants', icon: 'fa-industry' },
    { id: 'profiles', label: 'Profiles', icon: 'fa-sliders-h' },
    { id: 'loot', label: 'Loot', icon: 'fa-download' },
    { id: 'creds', label: 'Credentials', icon: 'fa-key' },
    { id: 'operators', label: 'Operators', icon: 'fa-users' },
    { id: 'events', label: 'Events', icon: 'fa-stream' }
  ];
  export let active = 'listeners';

  // --- Implant builds ---
  let builds = [];
  let buildsErr = '';
  let buildSearchQuery = '';
  let showGen = false;
  let showGenProfile = false;
  async function loadBuilds() {
    buildsErr = '';
    try {
      builds = await listImplantBuilds();
    } catch (e) { buildsErr = errorMessage(e); }
  }
  async function delBuild(name) {
    if (!(await dialog.confirm(`Delete build "${name}"?`, 'Confirm Delete'))) return;
    try { await DeleteImplantBuild(name); loadBuilds(); } catch (e) { buildsErr = errorMessage(e); }
  }
  async function regen(name) {
    try { await RegenerateImplant(name); } catch (e) { buildsErr = errorMessage(e); }
  }

  // --- Profiles ---
  let profiles = [];
  let profErr = '';
  let profSuccess = '';
  async function loadProfiles() {
    profErr = '';
    try {
      profiles = await listProfiles();
    } catch (e) { profErr = errorMessage(e); }
  }
  async function delProfile(name) {
    if (!(await dialog.confirm(`Delete profile "${name}"?`, 'Confirm Delete'))) return;
    try { await DeleteProfile(name); loadProfiles(); } catch (e) { profErr = errorMessage(e); }
  }
  async function generateProfile(id, name, format) {
    try {
      profErr = '';
      profSuccess = 'Generating...';
      const path = await GenerateImplantFromProfile(id, name, format || 0);
      if (path) {
        profSuccess = "Saved to " + path;
      } else {
        profSuccess = "";
      }
    } catch (e) { 
      profErr = errorMessage(e);
      profSuccess = '';
    }
  }

  function fmtFormat(f) {
    return ({ 0: 'shared lib', 1: 'shellcode', 2: 'executable', 3: 'service', 4: 'third-party' })[f] ?? f;
  }

  $: if (active === 'implants') loadBuilds();
  $: if (active === 'profiles') loadProfiles();
</script>

<div class="server-view">
  <div class="subtabs" role="tablist" aria-label="Server sections">
    {#each tabs as t}
      <button
        type="button"
        class="subtab"
        class:active={active === t.id}
        role="tab"
        aria-selected={active === t.id}
        on:click={() => active = t.id}
      >
        <i class="fas {t.icon}"></i> {t.label}
      </button>
    {/each}
  </div>

  <div class="section">
    {#if active === 'console'}
      <div class="pane" style="display: flex; flex-direction: column; height: 100%; padding: 0;">
        <SessionConsole sessionID="" />
      </div>
    {:else if active === 'listeners'}
      <ListenersPanel embedded />
    {:else if active === 'loot'}
      <LootPanel embedded />
    {:else if active === 'creds'}
      <CredentialsPanel embedded />
    {:else if active === 'operators'}
      <OperatorsPanel embedded />
    {:else if active === 'events'}
      <EventsLog events={eventHistory} />

    {:else if active === 'implants'}
      <div class="pane">
        <div class="pane-head">
          <h3>Implant Builds</h3>
          <div>
            <input type="text" class="input" placeholder="Search builds..." bind:value={buildSearchQuery} style="width: 200px;" />
            <button class="btn" on:click={loadBuilds}>Refresh</button>
            <button class="btn btn-primary" on:click={() => showGen = true}><i class="fas fa-plus"></i> Generate</button>
          </div>
        </div>
        {#if buildsErr}<div class="err">{buildsErr}</div>{/if}
        <table class="data-table">
          <thead><tr><th>Name</th><th>OS / Arch</th><th>Format</th><th>Type</th><th></th></tr></thead>
          <tbody>
            {#if builds.length === 0}
              <tr><td colspan="5" class="muted">No builds yet. Click Generate to create one.</td></tr>
            {:else}
              {#each builds.filter(b => b.name.toLowerCase().includes(buildSearchQuery.toLowerCase())) as b}
                <tr>
                  <td class="mono">{b.name}</td>
                  <td>{(b.GOOS || b.goos || '?')}/{(b.GOARCH || b.goarch || '?')}</td>
                  <td>{fmtFormat(b.Format ?? b.format)}</td>
                  <td>{(b.IsBeacon ?? b.isBeacon) ? 'beacon' : 'session'}</td>
                  <td class="row-actions">
                    <button class="btn" on:click={() => regen(b.name)}>Download</button>
                    <button class="btn btn-danger" on:click={() => delBuild(b.name)}>Delete</button>
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>

    {:else if active === 'profiles'}
      <div class="pane">
        <div class="pane-head">
          <h3>Implant Profiles</h3>
          <div>
            <button class="btn" on:click={loadProfiles}>Refresh</button>
            <button class="btn btn-primary" on:click={() => showGenProfile = true}><i class="fas fa-plus"></i> New Profile</button>
          </div>
        </div>
        {#if profErr}<div class="err">{profErr}</div>{/if}
        {#if profSuccess}<div class="ok">{profSuccess}</div>{/if}
        <table class="data-table">
          <thead><tr><th>Name</th><th>OS / Arch</th><th>Format</th><th>Type</th><th></th></tr></thead>
          <tbody>
            {#if profiles.length === 0}
              <tr><td colspan="5" class="muted">No saved profiles.</td></tr>
            {:else}
              {#each profiles as p}
                <tr>
                  <td class="mono">{p.Name || p.name}</td>
                  <td>{(p.Config?.GOOS || '?')}/{(p.Config?.GOARCH || '?')}</td>
                  <td>{fmtFormat(p.Config?.Format)}</td>
                  <td>{p.Config?.IsBeacon ? 'beacon' : 'session'}</td>
                  <td class="row-actions">
                    <button class="btn" on:click={() => generateProfile(p.Config?.ID || p.config?.id, p.Name || p.name, p.Config?.Format || 0)}>Generate</button>
                    <button class="btn btn-danger" on:click={() => delProfile(p.Name || p.name)}>Delete</button>
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</div>

{#if showGen}
  <GeneratePanel on:close={() => { showGen = false; loadBuilds(); }} />
{/if}
{#if showGenProfile}
  <GeneratePanel isProfile={true} on:close={() => { showGenProfile = false; loadProfiles(); }} />
{/if}

<style>
  .server-view { display: flex; flex-direction: column; height: 100%; background: var(--bg-color); }
  .subtabs { display: flex; gap: 2px; background: var(--toolbar-bg); border-bottom: 1px solid var(--panel-border); padding: 0 6px; flex-wrap: wrap; }
  .subtab { padding: 9px 16px; cursor: pointer; color: var(--text-muted); border-top: 2px solid transparent; font-size: 0.88em; white-space: nowrap; }
  .subtab { border-right: 0; border-bottom: 0; border-left: 0; background: transparent; font: inherit; }
  .subtab:hover { color: var(--text-color); }
  .subtab.active { color: var(--accent-color); background: var(--panel-bg); border-top: 2px solid var(--accent-color); }
  .subtab i { margin-right: 5px; }
  .section { flex: 1; overflow: auto; position: relative; }
  .pane { padding: 0 20px 16px 20px; }
  .pane-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; padding-top: 16px; }
  .pane-head h3 { margin: 0; color: var(--text-color); }
  .pane-head > div { display: flex; gap: 8px; }
  .err { color: var(--danger-color); margin-bottom: 12px; font-size: 0.9em; }
  .muted { color: var(--text-muted); text-align: center; padding: 16px; }
  .row-actions { display: flex; gap: 6px; }
</style>
