<script>
  import { AnsiUp } from 'ansi_up';
  import { onDestroy } from 'svelte';

  import {
    CancelBeaconTask,
    GetBeaconTaskOutput,
    listBeaconTasks,
  } from '../../api/agents.js';
  import { dialog } from '../../stores/dialogStore.js';
  import { errorMessage } from '../../utils/errors.js';
  import { startPolling } from '../../utils/polling.js';

  export let beaconID = '';
  export let active = true;

  let tasks = [];
  let error = '';
  let loading = true;
  let expandedTask = '';
  let expandedContent = '';
  let cancelingTask = '';
  let lastBeacon = null;
  let stopPoll = null;
  const ansiUp = new AnsiUp();
  ansiUp.use_classes = false;

  $: if (beaconID !== lastBeacon) {
    lastBeacon = beaconID;
    loading = true;
    tasks = [];
    expandedTask = '';
    refresh();
  }

  $: if (active && !stopPoll) {
    stopPoll = startPolling(refresh, 3000, { immediate: true });
  } else if (!active && stopPoll) {
    stopPoll();
    stopPoll = null;
  }

  onDestroy(() => stopPoll?.());

  async function refresh() {
    if (!beaconID) return;
    error = '';
    try {
      const next = await listBeaconTasks(beaconID);
      tasks = next.slice().sort((a, b) => (b.CreatedAt || 0) - (a.CreatedAt || 0));
    } catch (cause) {
      error = errorMessage(cause);
    } finally {
      loading = false;
    }
  }

  async function toggleTask(task) {
    if (expandedTask === task.ID) {
      expandedTask = '';
      return;
    }
    expandedTask = task.ID;
    expandedContent = 'Loading...';
    try {
      expandedContent = await GetBeaconTaskOutput(task.ID);
    } catch (cause) {
      expandedContent = errorMessage(cause, 'Failed to load task: ');
    }
  }

  async function cancelTask(event, task) {
    event.stopPropagation();
    const confirmed = await dialog.confirm(`Cancel pending task ${shortID(task.ID)}?`);
    if (!confirmed) return;

    cancelingTask = task.ID;
    try {
      await CancelBeaconTask(task.ID);
      await refresh();
    } catch (cause) {
      await dialog.alert(errorMessage(cause, 'Could not cancel task: '));
    } finally {
      cancelingTask = '';
    }
  }

  function shortID(id) {
    return String(id || '').split('-')[0];
  }

  function formatTime(value) {
    return value ? new Date(value * 1000).toLocaleString() : '-';
  }
</script>

<div class="task-panel">
  <div class="task-header">
    <span>Beacon Tasks</span>
    <button class="btn" type="button" on:click={refresh}>Refresh</button>
  </div>

  <div class="task-body">
    {#if error}
      <div class="message error">{error}</div>
    {:else if loading}
      <div class="message">Loading tasks...</div>
    {:else if tasks.length === 0}
      <div class="message">No tasks have been queued for this beacon.</div>
    {:else}
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Message Type</th>
            <th>State</th>
            <th>Created</th>
            <th>Sent</th>
            <th>Completed</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each tasks as task (task.ID)}
            <tr
              class="clickable"
              class:selected={expandedTask === task.ID}
              on:click={() => toggleTask(task)}
            >
              <td class="mono">{shortID(task.ID)}</td>
              <td>{String(task.Description || '').replace(/Req$/, '') || '-'}</td>
              <td class="state state-{String(task.State || '').toLowerCase()}">
                {task.State || '-'}
              </td>
              <td>{formatTime(task.CreatedAt)}</td>
              <td>{formatTime(task.SentAt)}</td>
              <td>{formatTime(task.CompletedAt)}</td>
              <td class="actions">
                {#if String(task.State || '').toLowerCase() === 'pending'}
                  <button
                    class="btn danger"
                    type="button"
                    disabled={cancelingTask === task.ID}
                    on:click={(event) => cancelTask(event, task)}
                  >
                    {cancelingTask === task.ID ? 'Canceling...' : 'Cancel'}
                  </button>
                {/if}
              </td>
            </tr>
            {#if expandedTask === task.ID}
              <tr class="expanded-row">
                <td colspan="7">
                  <!-- ansi_up escapes task output before producing its own span markup. -->
                  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                  <pre class="task-output mono">{@html ansiUp.ansi_to_html(expandedContent)}</pre>
                </td>
              </tr>
            {/if}
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<style>
  .task-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-color);
  }
  .task-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 15px;
    border-bottom: 1px solid var(--border-color);
    background: var(--header-bg);
  }
  .task-body {
    flex: 1;
    overflow: auto;
  }
  .message {
    padding: 16px;
    color: var(--text-muted);
  }
  .message.error {
    color: var(--danger-color);
  }
  .clickable {
    cursor: pointer;
  }
  .state {
    font-weight: 600;
  }
  .state-completed {
    color: var(--success-color);
  }
  .state-sent,
  .state-pending {
    color: #d6a23e;
  }
  .state-canceled {
    color: var(--text-muted);
  }
  .actions {
    width: 1%;
    white-space: nowrap;
  }
  .btn.danger {
    color: var(--danger-color);
  }
  .expanded-row td {
    padding: 0;
    background: var(--toolbar-bg);
    border-bottom: 2px solid var(--accent-color);
  }
  .task-output {
    max-height: 420px;
    margin: 0;
    padding: 15px;
    overflow: auto;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    color: var(--text-color);
    font-size: 0.9em;
  }
</style>
