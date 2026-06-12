<script>
  import { onMount, createEventDispatcher } from "svelte";
  import { Connect, GetClientConfigs } from "../../api/connection.js";
  import CustomSelect from "../../components/ui/CustomSelect.svelte";
  import { errorMessage } from "../../utils/errors.js";

  const dispatch = createEventDispatcher();

  let configs = [];
  let loading = true;
  let error = "";
  let selectedConfig = "";
  let connecting = false;

  onMount(async () => {
    try {
      configs = await GetClientConfigs();
      if (!configs || configs.length === 0) {
        error = "No sliver profiles found in ~/.sliver-client/configs";
        loading = false;
        return;
      }

      // Auto-connect if there's only one config
      if (configs.length === 1) {
        connect(configs[0]);
      } else {
        selectedConfig = configs[0];
        loading = false;
      }
    } catch (e) {
      error = errorMessage(e, "Failed to load configs: ");
      loading = false;
    }
  });

  async function connect(profile) {
    if (connecting) return;
    connecting = true;
    error = "";
    try {
      await Connect(profile);
      dispatch("connected", profile);
    } catch (e) {
      error = errorMessage(e, "Connection failed: ");
      connecting = false;
    }
  }
</script>

<div class="connection-screen">
  <div class="connection-box">
    <div class="logo">
      <img src="/wails.png" alt="Sliver Logo" />
      <h2>Sliver GUI</h2>
    </div>

    {#if loading}
      <div class="status">Loading profiles...</div>
    {:else if error}
      <div class="error">{error}</div>
      <button class="btn btn-primary" on:click={() => window.location.reload()}
        >Retry</button
      >
    {:else}
      <div class="form-group">
        <label for="teamserver-profile">Select Teamserver Profile</label>
        <CustomSelect
          id="teamserver-profile"
          bind:value={selectedConfig}
          options={configs}
          disabled={connecting}
        />
      </div>

      <button
        class="btn btn-primary w-100"
        on:click={() => connect(selectedConfig)}
        disabled={connecting}
      >
        {connecting ? "Connecting..." : "Connect"}
      </button>
    {/if}
  </div>
</div>

<style>
  .connection-screen {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: var(--bg-color);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 10000;
  }
  .connection-box {
    background-color: var(--panel-bg);
    border: 1px solid var(--panel-border);
    border-radius: 8px;
    padding: 30px;
    width: 350px;
    max-width: 90%;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
    text-align: center;
  }
  .logo {
    margin-bottom: 25px;
  }
  .logo img {
    width: 80px;
    height: 80px;
    margin-bottom: 10px;
  }
  .logo h2 {
    margin: 0;
    color: var(--text-color);
    font-weight: 300;
  }
  .form-group {
    text-align: left;
    margin-bottom: 20px;
  }
  label {
    display: block;
    margin-bottom: 8px;
    color: var(--text-muted);
    font-size: 0.9em;
  }
  .w-100 {
    width: 100%;
    padding: 12px;
    font-size: 1.1em;
  }
  .status {
    color: var(--text-muted);
  }
  .error {
    color: var(--danger-color);
    background: rgba(255, 74, 74, 0.1);
    padding: 10px;
    border-radius: 4px;
    margin-bottom: 20px;
    text-align: left;
    font-size: 0.9em;
    word-break: break-word;
  }
</style>
