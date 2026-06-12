<script>
  import { onMount, createEventDispatcher } from "svelte";
  import { listOperators } from "../../api/server.js";
  import PanelShell from "../../components/ui/PanelShell.svelte";
  import { errorMessage } from "../../utils/errors.js";
  import { startPolling } from "../../utils/polling.js";

  export let embedded = false;
  const dispatch = createEventDispatcher();
  function close() {
    dispatch("close");
  }

  let operators = [];
  let loading = true;
  let error = "";

  async function fetchOperators() {
    try {
      operators = await listOperators();
      error = "";
    } catch (e) {
      error = errorMessage(e, "Failed to load operators: ");
    } finally {
      loading = false;
    }
  }

  onMount(() => startPolling(fetchOperators, 5000));
</script>

<PanelShell title="Operators" icon="fa-users" width="800px" {embedded} on:close={close}>
  {#if error}
    <div style="color: var(--danger-color); margin-bottom: 15px;">{error}</div>
  {/if}

  {#if loading}
    <p>Loading operators...</p>
  {:else if operators.length === 0}
    <p>No operators connected.</p>
  {:else}
    <table class="data-table">
      <thead><tr><th>Status</th><th>Name</th></tr></thead>
      <tbody>
        {#each operators as op}
          <tr>
            <td>
              {#if op.Online}
                <span style="color: var(--success-color)">● Online</span>
              {:else}
                <span style="color: var(--danger-color)">○ Offline</span>
              {/if}
            </td>
            <td>{op.Name || "-"}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</PanelShell>
