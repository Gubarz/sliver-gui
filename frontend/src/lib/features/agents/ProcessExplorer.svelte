<script>
  import { createEventDispatcher } from "svelte";
  import { KillProcess, listProcesses } from "../../api/agents.js";
  import { dialog } from "../../stores/dialogStore.js";
  import { errorMessage } from "../../utils/errors.js";
  import ContextMenu from "./ContextMenu.svelte";
  import DataTable from "../../components/ui/DataTable.svelte";

  export let sessionID = "";

  const dispatch = createEventDispatcher();

  let processes = [];
  let loading = true;
  let error = "";
  let isTreeView = false;
  let isFullView = false;
  // Full View is per-agent: each session tracks its own opsec-gated setting, so
  // enabling it on one agent never turns it on (or re-enumerates) for another.
  const fullViewBySession = new Map();

  // Context menu state
  let showContextMenu = false;
  let contextMenuX = 0;
  let contextMenuY = 0;
  let contextMenuTargetPid = null;
  let contextMenuTargetName = "";

  const processMenuActions = [
    { id: "migrate", icon: "fas fa-exchange-alt", label: "Migrate Into" },
    { id: "procdump", icon: "fas fa-syringe", label: "Dump Memory" },
    { id: "getsystem", icon: "fas fa-user-shield", label: "Get System (via this process)" },
    { id: "kill", icon: "fas fa-skull", label: "Kill Process" },
  ];

  const baseColumns = [
    { key: "PidStr", label: "PID", width: 120 },
    { key: "PpidStr", label: "PPID", width: 80 },
    { key: "ExecutableStr", label: "Executable", width: 250 },
  ];
  const fullColumns = [
    ...baseColumns.map((c) => ({ ...c })),
    { key: "OwnerStr", label: "Owner", width: 170 },
    { key: "ArchStr", label: "Arch", width: 80 },
    { key: "SessionStr", label: "Session", width: 80 },
  ];
  // Full View reveals deeper (and noisier) process detail.
  $: tableColumns = isFullView ? fullColumns : baseColumns;

  // Reload when the active session changes (instance is reused across sessions).
  let lastSession = null;
  $: if (sessionID !== lastSession) {
    lastSession = sessionID;
    // Restore this agent's Full View setting (defaults off → opsec-safe basic ps).
    isFullView = fullViewBySession.get(sessionID) || false;
    loadProcesses();
  }

  // Add string versions for sorting/filtering
  $: normalizedProcesses = processes.map((p) => ({
    ...p,
    PidStr: String(p.Pid ?? p.pid ?? 0),
    PpidStr: String(p.Ppid ?? p.ppid ?? 0),
    ExecutableStr: p.Executable ?? p.executable ?? "",
    OwnerStr: p.Owner ?? p.owner ?? "",
    ArchStr: p.Architecture ?? p.architecture ?? "",
    SessionStr: String(p.SessionID ?? p.sessionID ?? ""),
  }));

  $: displayProcesses = isTreeView
    ? buildProcessTree(normalizedProcesses)
    : normalizedProcesses;

  function buildProcessTree(procs) {
    const tree = {};
    const rootNodes = [];
    const pidMap = new Set();

    const normalized = procs.map((p) => ({
      ...p,
      p_id: p.Pid || p.pid,
      pp_id: p.Ppid || p.ppid,
    }));
    normalized.forEach((p) => pidMap.add(p.p_id));

    normalized.forEach((p) => {
      if (p.pp_id === 0 || !pidMap.has(p.pp_id)) {
        rootNodes.push(p);
      } else {
        if (!tree[p.pp_id]) tree[p.pp_id] = [];
        tree[p.pp_id].push(p);
      }
    });

    const flattened = [];
    function traverse(node, depth) {
      flattened.push({ ...node, _depth: depth });
      if (tree[node.p_id]) {
        tree[node.p_id]
          .sort((a, b) => a.p_id - b.p_id)
          .forEach((child) => traverse(child, depth + 1));
      }
    }

    rootNodes
      .sort((a, b) => a.p_id - b.p_id)
      .forEach((root) => traverse(root, 0));
    return flattened;
  }

  async function loadProcesses() {
    loading = true;
    error = "";
    try {
      processes = await listProcesses(sessionID, isFullView);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      loading = false;
    }
  }

  // Full View surfaces owner/arch/session — deeper enumeration that isn't
  // opsec-safe, so confirm before enabling. Turning it back off is instant.
  async function toggleFullView() {
    if (isFullView) {
      isFullView = false;
      fullViewBySession.set(sessionID, false);
      await loadProcesses();
      return;
    }
    const sid = sessionID; // capture: the user may switch agents during the dialog
    const ok = await dialog.confirm(
      "Full View performs deeper process enumeration (owner, architecture, session) " +
        "on this agent. This is NOT opsec-safe and may trigger EDR. Continue?",
      "Opsec Warning",
    );
    if (!ok) return;
    fullViewBySession.set(sid, true);
    if (sid !== sessionID) return; // switched away; the setting is saved for that agent
    isFullView = true;
    await loadProcesses();
  }

  async function killProcess(pid) {
    if (
      !(await dialog.confirm(
        `Are you sure you want to kill PID ${pid}?`,
        "Confirm Kill",
      ))
    )
      return;
    try {
      await KillProcess(sessionID, pid);
      await loadProcesses();
    } catch (err) {
      await dialog.alert(
        errorMessage(err, "Failed to kill process: "),
        "Kill Error",
      );
    }
  }

  function handleRightClick(event, proc) {
    showContextMenu = true;
    contextMenuX = event.clientX;
    contextMenuY = event.clientY;
    contextMenuTargetPid = proc.Pid || proc.pid;
    contextMenuTargetName = proc.Executable || proc.executable || proc.ExecutableStr || "";
  }

  function handleMenuSelect(event) {
    showContextMenu = false;
    const pid = contextMenuTargetPid;
    const name = contextMenuTargetName;
    if (!pid) return;
    switch (event.detail.action) {
      case "kill":
        killProcess(pid);
        break;
      case "migrate":
        // Run in the agent's console so the (async) result is visible there.
        dispatch("command", { cmd: `migrate -p ${pid}` });
        break;
      case "procdump":
        dispatch("command", { cmd: `procdump -p ${pid}` });
        break;
      case "getsystem":
        // getsystem hosts in a named process, not a PID.
        if (name) dispatch("command", { cmd: `getsystem -p ${name}` });
        break;
    }
  }

  function hideContextMenu() {
    showContextMenu = false;
  }
</script>

<svelte:window on:click={hideContextMenu} />

<div class="tab-wrapper">
  <div class="tab-header">
    Process Explorer
    <button
      class="btn"
      class:btn-danger={isFullView}
      style="float: right; margin-top: -4px; margin-left: 10px;"
      title="Reveals owner/arch/session via deeper enumeration (not opsec-safe)"
      on:click={toggleFullView}
    >
      {isFullView ? "Basic View" : "Full View"}
    </button>
    <button
      class="btn"
      style="float: right; margin-top: -4px; margin-left: 10px;"
      on:click={() => (isTreeView = !isTreeView)}
    >
      {isTreeView ? "List View" : "Tree View"}
    </button>
    <button
      class="btn"
      style="float: right; margin-top: -4px;"
      on:click={loadProcesses}>Refresh</button
    >
  </div>

  <div class="tab-content">
    {#if loading}
      <div style="padding: 20px; text-align: center;">Loading processes...</div>
    {:else if error}
      <div
        style="padding: 20px; color: var(--danger-color); text-align: center;"
      >
        {error}
      </div>
    {:else}
      <DataTable
        data={displayProcesses}
        columns={tableColumns}
        defaultSortKey="PidStr"
        let:rows
        let:columns
      >
        {#each rows as p}
          <tr
            on:contextmenu|preventDefault={(e) => handleRightClick(e, p)}
          >
            <td
              class="mono"
              style="width: {columns[0].width}px; max-width: {columns[0]
                .width}px;"
              title={p.PidStr}
            >
              {#if isTreeView}
                <span
                  style="display: inline-block; margin-left: {p._depth * 20}px;"
                >
                  {#if p._depth > 0}<span
                      style="color: #666; margin-right: 5px;">&rdsh;</span
                    >{/if}
                  {p.PidStr}
                </span>
              {:else}
                {p.PidStr}
              {/if}
            </td>
            <td
              class="mono"
              style="width: {columns[1].width}px; max-width: {columns[1]
                .width}px;"
              title={p.PpidStr}>{p.PpidStr}</td
            >
            <td
              style="width: {columns[2].width}px; max-width: {columns[2]
                .width}px;"
              title={p.ExecutableStr}>{p.ExecutableStr}</td
            >
            {#if isFullView}
              <td
                style="width: {columns[3].width}px; max-width: {columns[3]
                  .width}px;"
                title={p.OwnerStr}>{p.OwnerStr}</td
              >
              <td
                class="mono"
                style="width: {columns[4].width}px; max-width: {columns[4]
                  .width}px;"
                title={p.ArchStr}>{p.ArchStr}</td
              >
              <td
                class="mono"
                style="width: {columns[5].width}px; max-width: {columns[5]
                  .width}px;"
                title={p.SessionStr}>{p.SessionStr}</td
              >
            {/if}
          </tr>
        {/each}
        {#if rows.length === 0}
          <tr>
            <td
              colspan={columns.length}
              style="text-align: center; color: var(--text-muted); padding: 20px;"
              >No processes found.</td
            >
          </tr>
        {/if}
      </DataTable>
    {/if}
  </div>
</div>

{#if showContextMenu}
  <ContextMenu
    x={contextMenuX}
    y={contextMenuY}
    coreActions={processMenuActions}
    footerActions={[]}
    dangerActions={[]}
    categories={[]}
    on:select={handleMenuSelect}
  />
{/if}

<style>
  .tab-wrapper {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--bg-color);
    border-radius: 4px;
  }
  .tab-header {
    background-color: var(--header-bg);
    padding: 10px 15px;
    font-size: 0.9em;
    border-bottom: 1px solid var(--border-color);
    margin-bottom: 10px;
  }
  .tab-content {
    flex: 1;
    overflow-y: auto;
  }
  tr:hover {
    background-color: rgba(255, 255, 255, 0.05);
  }
</style>
