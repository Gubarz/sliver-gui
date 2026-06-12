<script>
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import DataTable from '../../components/ui/DataTable.svelte';
  import {
    agentRemoteAddress,
    isAgentOnline,
    pivotParentMap,
    shortAgentID,
  } from '../../utils/agents.js';
  export let data = [];
  export let pivotGraph = null;
  export const title = "Sessions table";
  export let selectedId = null;
  export let filterable = true;

  const dispatch = createEventDispatcher();

  // Tick once a second so "last checkin" stays current between event-driven
  // list refreshes. Display only — no server calls.
  let now = Math.floor(Date.now() / 1000);
  const ticker = setInterval(() => { now = Math.floor(Date.now() / 1000); }, 1000);
  onDestroy(() => clearInterval(ticker));

  function fmtCheckin(ts, nowSec) {
    if (!ts) return "-";
    const s = nowSec - ts;
    if (s < 2) return "just now";
    if (s < 60) return `${s}s ago`;
    if (s < 3600) return `${Math.floor(s / 60)}m ago`;
    if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
    return `${Math.floor(s / 86400)}d ago`;
  }

  let notes = {};

  onMount(() => {
    try {
      notes = JSON.parse(localStorage.getItem('sliver_agent_notes')) || {};
    } catch {}
  });

  function saveNote(id, text) {
    notes[id] = text;
    localStorage.setItem('sliver_agent_notes', JSON.stringify(notes));
  }

  function handleRightClick(e, session) {
    dispatch('contextmenu', { event: e, session });
  }
  function handleRowClick(session) {
    dispatch('select', session.ID);
  }
  function handleRowDblClick(session) {
    dispatch('interact', session.ID);
  }

  function getOsIcon(osName) {
    let lower = (osName || "").toLowerCase();
    if (lower.includes("win")) return "fab fa-windows";
    if (lower.includes("linux")) return "fab fa-linux";
    if (lower.includes("darwin") || lower.includes("mac")) return "fab fa-apple";
    return "fas fa-desktop";
  }

  function getTypeStr(item) {
    return item._kind || (item.NextCheckin !== undefined ? "beacon" : "session");
  }

  function getPrivilegeClass(username) {
    const lower = (username || "").toLowerCase();
    const account = lower.split(/[\\/]/).pop() || "";
    const isComputerAccount = account.endsWith("$");
    if (
      account === "root" ||
      lower.includes("system") ||
      lower.includes("admin") ||
      isComputerAccount
    ) {
      return "priv-high";
    }
    return "priv-normal";
  }

  $: pivotParents = pivotParentMap(pivotGraph);
  $: normalizedData = data.map((agent) => {
    const online = isAgentOnline(agent, now);
    return {
      ...agent,
      Note: notes[agent.ID] || "",
      ImplantName: agent.Name || "-",
      RemoteHost: agentRemoteAddress(agent, pivotParents, data),
      LastCheckin: agent.LastCheckin ?? agent.lastCheckin ?? 0,
      Online: online,
      Status: online ? "Online" : "Offline",
    };
  });

  let tableColumns = [
    { key: "Status", label: "Status", width: 42 },
    { key: "ID", label: "Agent ID", width: 72 },
    { key: "ImplantName", label: "Implant Name", width: 96 },
    { key: "_kind", label: "Type", width: 56 },
    { key: "Transport", label: "Transport", width: 64 },
    { key: "RemoteHost", label: "Remote Address", width: 96 },
    { key: "Hostname", label: "Computer", width: 90 },
    { key: "Username", label: "User", width: 90 },
    { key: "OS", label: "OS", width: 32 },
    { key: "Filename", label: "Process", width: 96 },
    { key: "PID", label: "PID", width: 44 },
    { key: "LastCheckin", label: "Last Checkin", width: 78 },
    { key: "Note", label: "Note", width: 100 }
  ];
</script>

<DataTable data={normalizedData} columns={tableColumns} {filterable} compact defaultSortKey="Hostname" let:rows let:columns>
  {#each rows as item}
    <tr
      on:contextmenu|preventDefault={(e) => handleRightClick(e, item)}
      on:click={() => handleRowClick(item)}
      on:dblclick={() => handleRowDblClick(item)}
      class="{getPrivilegeClass(item.Username)}"
      class:selected={selectedId === item.ID}
    >
      <td style="width: {columns[0].width}px; max-width: {columns[0].width}px;">
        <span class="status" class:online={item.Online} title={item.Online ? 'Online' : 'Offline'}>
          <span class="status-dot"></span>
        </span>
      </td>
      <td style="width: {columns[1].width}px; max-width: {columns[1].width}px;" title={item.ID}>{shortAgentID(item.ID)}</td>
      <td style="width: {columns[2].width}px; max-width: {columns[2].width}px;" title={item.ImplantName}>{item.ImplantName}</td>
      <td style="width: {columns[3].width}px; max-width: {columns[3].width}px;"><span class="type type-{getTypeStr(item)}">{getTypeStr(item)}</span></td>
      <td style="width: {columns[4].width}px; max-width: {columns[4].width}px;" title={item.Transport || "-"}>{item.Transport || "-"}</td>
      <td style="width: {columns[5].width}px; max-width: {columns[5].width}px;" title={item.RemoteHost || "-"}>{item.RemoteHost || "-"}</td>
      <td style="width: {columns[6].width}px; max-width: {columns[6].width}px;" title={item.Hostname}>{item.Hostname}</td>
      <td style="width: {columns[7].width}px; max-width: {columns[7].width}px;" title={item.Username}>{item.Username}</td>
      <td style="width: {columns[8].width}px; max-width: {columns[8].width}px;" title={item.OS}><i class="{getOsIcon(item.OS)}"></i></td>
      <td style="width: {columns[9].width}px; max-width: {columns[9].width}px;" title={item.Filename || item.ProcessName || "-"}>{item.Filename || item.ProcessName || "-"}</td>
      <td style="width: {columns[10].width}px; max-width: {columns[10].width}px;">{item.PID}</td>
      <td class="mono" style="width: {columns[11].width}px; max-width: {columns[11].width}px;" title={item.LastCheckin ? new Date(item.LastCheckin * 1000).toLocaleString() : '-'}>{fmtCheckin(item.LastCheckin, now)}</td>
      <td style="width: {columns[12].width}px; max-width: {columns[12].width}px;" on:dblclick|stopPropagation>
        <input 
          class="note-input" 
          type="text" 
          placeholder="Add note..."
          bind:value={notes[item.ID]} 
          on:input={() => saveNote(item.ID, notes[item.ID])}
        />
      </td>
    </tr>
  {/each}
  {#if rows.length === 0}
    <tr>
      <td colspan={columns.length} style="text-align: center; color: var(--text-muted); padding: 20px;">
        {data.length === 0 ? 'No active agents found' : 'No agents match the filter'}
      </td>
    </tr>
  {/if}
</DataTable>

<style>

  /* Privilege row coloring */
  tr.priv-high td {
    color: #ff6b6b; /* Red/orange text for high privilege */
  }
  tr.priv-high:hover td {
    background: rgba(255, 107, 107, 0.1);
  }

  .type { padding: 1px 7px; border-radius: 10px; font-size: 0.8em; }
  .type-session { background: rgba(70,160,90,0.18); color: var(--success-color); }
  .type-beacon { background: rgba(214,162,62,0.18); color: #d6a23e; }
  .status {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--text-muted);
  }
  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--danger-color);
  }
  .status.online {
    color: var(--success-color);
  }
  .status.online .status-dot {
    background: var(--success-color);
    box-shadow: 0 0 6px var(--success-color);
  }

  .note-input {
    width: 100%;
    background: transparent;
    border: none;
    color: inherit;
    font-family: inherit;
    font-size: inherit;
    outline: none;
    padding: 2px 4px;
    border-radius: 3px;
  }
  .note-input:hover, .note-input:focus {
    background: rgba(255,255,255,0.1);
  }
</style>
