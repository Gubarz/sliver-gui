<script>
  import { onMount, createEventDispatcher } from "svelte";
  import { GetPivots } from "../../api/server.js";
  import PanelShell from "../../components/ui/PanelShell.svelte";
  import { errorMessage } from "../../utils/errors.js";
  import { startPolling } from "../../utils/polling.js";

  export let embedded = false;

  const dispatch = createEventDispatcher();
  function close() {
    dispatch("close");
  }

  let pivotGraph = null;
  let loading = true;
  let error = "";

  async function fetchPivots() {
    try {
      pivotGraph = await GetPivots();
      error = "";
    } catch (e) {
      error = errorMessage(e, "Failed to load pivots: ");
    } finally {
      loading = false;
    }
  }

  onMount(() => startPolling(fetchPivots, 5000));

  // Flatten the tree for simple tabular display
  function flattenGraph(nodes, depth = 0, result = []) {
    if (!nodes) return result;
    for (const node of nodes) {
      result.push({ ...node, _depth: depth });
      flattenGraph(node.Children || node.children || [], depth + 1, result);
    }
    return result;
  }

  $: flatNodes = flattenGraph(
    pivotGraph?.Children || pivotGraph?.children || [],
  );
</script>

<PanelShell title="Active Pivots" icon="fa-project-diagram" width="800px" bodyPadding="20px" {embedded} on:close={close}>
  {#if error}
    <div style="color: var(--danger-color); margin-bottom: 15px;">{error}</div>
  {/if}

  {#if loading}
    <p>Loading pivots...</p>
  {:else if flatNodes.length === 0}
    <p>No active pivots.</p>
  {:else}
    <table class="data-table">
      <thead>
        <tr><th>Tree</th><th>Name</th><th>Peer ID</th><th>Session ID</th></tr>
      </thead>
      <tbody>
        {#each flatNodes as node}
          <tr>
            <td class="mono">
              <span style="display:inline-block; width: {node._depth * 20}px"></span>
              {#if node._depth > 0}↳{/if}
              {node._depth === 0 ? "root" : "peer"}
            </td>
            <td>{node.Name || node.name || "-"}</td>
            <td class="mono">{node.PeerID || node.peerID || "-"}</td>
            <td class="mono">
              {(node.Session?.ID || node.session?.id || node.Session?.id || "-").substring(0, 8)}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</PanelShell>
