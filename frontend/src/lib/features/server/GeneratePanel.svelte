<script>
  import { createEventDispatcher } from 'svelte';
  import { GenerateImplant, SaveProfile } from '../../api/server.js';
  import CustomSelect from '../../components/ui/CustomSelect.svelte';
  import PanelShell from '../../components/ui/PanelShell.svelte';
  import { errorMessage } from '../../utils/errors.js';

  const dispatch = createEventDispatcher();
  function close() { dispatch('close'); }
  export let isProfile = false;

  let goos = "windows";
  let goarch = "amd64";
  let format = "exe";
  let c2url = "mtls://";
  let name = "";
  let isBeacon = false;
  let interval = 60;

  let generating = false;
  let error = "";
  let result = "";

  async function generate() {
    generating = true;
    error = "";
    result = "";
    try {
      if (isProfile) {
        if (!name) throw new Error("A profile name is required.");
        await SaveProfile(name, goos, goarch, format, c2url, isBeacon, Number(interval));
        result = "Profile saved: " + name;
        close();
      } else {
        const path = await GenerateImplant(goos, goarch, format, c2url, name, isBeacon, Number(interval));
        result = path ? "Saved to " + path : "Cancelled.";
      }
    } catch (e) {
      error = errorMessage(e);
    } finally {
      generating = false;
    }
  }
</script>

<PanelShell
  title={isProfile ? 'Create Profile' : 'Generate Implant'}
  icon={isProfile ? 'fa-sliders-h' : 'fa-industry'}
  width="520px"
  bodyPadding="18px 20px"
  on:close={close}
>
  <div class="grid">
    <div class="field">
      <label for="generate-os">OS:</label>
      <CustomSelect id="generate-os" bind:value={goos} options={['windows', 'linux', 'darwin']} />
    </div>
    <div class="field">
      <label for="generate-arch">Arch:</label>
      <CustomSelect id="generate-arch" bind:value={goarch} options={['amd64', '386', 'arm64']} />
    </div>
    <div class="field">
      <label for="generate-format">Format:</label>
      <CustomSelect id="generate-format" bind:value={format} options={['exe', 'shared', 'shellcode', 'service']} />
    </div>
  </div>

  <label class="full">C2 URL
    <input type="text" bind:value={c2url} placeholder="mtls://10.0.0.1:443" />
  </label>

  <label class="full">Name {#if !isProfile}(optional){/if}
    <input type="text" bind:value={name} placeholder={isProfile ? "profile name" : "auto-generated if blank"} />
  </label>

  <label class="checkbox">
    <input type="checkbox" bind:checked={isBeacon} /> Beacon
    {#if isBeacon}
      <span class="interval">interval (s): <input aria-label="Beacon interval in seconds" type="number" bind:value={interval} style="width:70px" /></span>
    {/if}
  </label>

  {#if error}<div class="error">{error}</div>{/if}
  {#if result}<div class="ok">{result}</div>{/if}

  <div class="actions">
    <button class="btn btn-primary" on:click={generate} disabled={generating}>
      {#if generating}
        {isProfile ? 'Saving...' : 'Generating… (this can take a minute)'}
      {:else}
        {isProfile ? 'Save Profile' : 'Generate'}
      {/if}
    </button>
  </div>
</PanelShell>

<style>
  .grid { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 10px; margin-bottom: 12px; }
  label { display: flex; flex-direction: column; font-size: 0.85em; color: var(--text-muted); gap: 4px; }
  label.full { margin-bottom: 12px; }
  label.checkbox { flex-direction: row; align-items: center; gap: 8px; color: var(--text-color); margin-bottom: 12px; }
  .interval { color: var(--text-muted); }
  input[type=text], input[type=number] {
    background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color);
    padding: 6px 10px; border-radius: 3px; font-family: var(--font-mono); outline: none;
  }
  .error { color: var(--danger-color); margin-bottom: 10px; font-size: 0.9em; }
  .ok { color: var(--success-color); margin-bottom: 10px; font-size: 0.9em; word-break: break-all; }
  .actions { text-align: right; }
</style>
