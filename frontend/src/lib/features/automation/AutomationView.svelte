<script>
  import { onMount } from 'svelte';

  import {
    ClearAutomationHistory,
    DeleteAutomationRule,
    GetAutomationHistory,
    ListAutomationRules,
    RunAutomationRule,
    SaveAutomationRule,
    SetAutomationRuleEnabled,
  } from '../../api/automation.js';
  import { onAutomationRun, onAutomationUpdated } from '../../api/runtime.js';
  import { dialog } from '../../stores/dialogStore.js';
  import { matchesAutomationTarget } from '../../utils/automation.js';
  import { errorMessage } from '../../utils/errors.js';

  export let sessions = [];
  export let beacons = [];

  const triggerOptions = [
    ['session-connected', 'Session connected'],
    ['beacon-registered', 'Beacon registered'],
    ['beacon-checkin', 'Beacon check-in'],
    ['interval', 'Recurring interval'],
    ['manual', 'Manual only'],
  ];

  let rules = [];
  let history = [];
  let selectedID = '';
  let draft = null;
  let selectedRun = null;
  let activePanel = 'rule';
  let manualTarget = '';
  let busy = false;
  let loadError = '';

  $: targets = [
    ...sessions.map((item) => ({ ...item, _kind: 'session' })),
    ...beacons.map((item) => ({ ...item, _kind: 'beacon' })),
  ];
  $: matchingTargets = draft
    ? targets.filter((target) => matchesAutomationTarget(target, draft))
    : targets;

  onMount(() => {
    refresh();
    const stopUpdated = onAutomationUpdated(refreshRules);
    const stopRun = onAutomationRun(refreshHistory);
    return () => {
      stopUpdated();
      stopRun();
    };
  });

  async function refresh() {
    await Promise.all([refreshRules(), refreshHistory()]);
  }

  async function refreshRules() {
    try {
      rules = await ListAutomationRules();
      if (selectedID) {
        const selected = rules.find((rule) => rule.id === selectedID);
        if (selected && !draft?.__dirty) draft = copyRule(selected);
      }
    } catch (error) {
      loadError = errorMessage(error);
    }
  }

  async function refreshHistory() {
    try {
      history = await GetAutomationHistory();
      if (selectedRun) {
        selectedRun = history.find((run) => run.id === selectedRun.id) || selectedRun;
      }
    } catch (error) {
      loadError = errorMessage(error);
    }
  }

  function blankRule() {
    return {
      id: '',
      name: 'New automation',
      description: '',
      enabled: true,
      trigger: 'manual',
      targetKind: 'any',
      filter: { os: '', arch: '', hostname: '', username: '', name: '' },
      executionMode: 'commands',
      commands: [''],
      script: `// target: { id, name, hostname, username, os, arch, kind }
// trigger: { type }
const whoami = sliver.run("whoami");
sliver.log("Agent:", target.name || target.hostname);
sliver.log(whoami);`,
      timeoutSeconds: 300,
      continueOnError: false,
      delaySeconds: 0,
      cooldownSeconds: 0,
      intervalSeconds: 60,
      maxRuns: 0,
      runCount: 0,
      __dirty: true,
    };
  }

  function copyRule(rule) {
    const copy = JSON.parse(JSON.stringify(rule));
    copy.filter ||= { os: '', arch: '', hostname: '', username: '', name: '' };
    copy.commands ||= [];
    copy.executionMode ||= 'commands';
    copy.script ||= '';
    copy.timeoutSeconds ||= 300;
    copy.__dirty = false;
    return copy;
  }

  function createRule() {
    selectedID = '';
    draft = blankRule();
    activePanel = 'rule';
    manualTarget = '';
  }

  function selectRule(rule) {
    selectedID = rule.id;
    draft = copyRule(rule);
    activePanel = 'rule';
    manualTarget = '';
  }

  function markDirty() {
    if (draft) draft.__dirty = true;
  }

  function updateFilter(key, value) {
    draft.filter = { ...draft.filter, [key]: value };
    markDirty();
  }

  function updateCommands(value) {
    draft.commands = value.split('\n');
    markDirty();
  }

  async function saveRule() {
    if (!draft || busy) return;
    busy = true;
    try {
      const payload = { ...draft };
      delete payload.__dirty;
      const saved = await SaveAutomationRule(payload);
      selectedID = saved.id;
      draft = copyRule(saved);
      await refreshRules();
    } catch (error) {
      await dialog.alert(`Could not save automation: ${errorMessage(error)}`);
    } finally {
      busy = false;
    }
  }

  async function deleteRule() {
    if (!draft?.id || !(await dialog.confirm(
      `Delete "${draft.name}"? Existing run history will be retained.`,
      'Delete Automation',
    ))) return;
    busy = true;
    try {
      await DeleteAutomationRule(draft.id);
      selectedID = '';
      draft = null;
      await refreshRules();
    } catch (error) {
      await dialog.alert(`Could not delete automation: ${errorMessage(error)}`);
    } finally {
      busy = false;
    }
  }

  async function toggleRule(rule) {
    try {
      await SetAutomationRuleEnabled(rule.id, !rule.enabled);
      await refreshRules();
    } catch (error) {
      await dialog.alert(`Could not update automation: ${errorMessage(error)}`);
    }
  }

  async function runNow() {
    if (!draft?.id) {
      await dialog.alert('Save the rule before running it.');
      return;
    }
    busy = true;
    try {
      await RunAutomationRule(draft.id, manualTarget);
      activePanel = 'history';
      setTimeout(refreshHistory, 150);
    } catch (error) {
      await dialog.alert(`Could not run automation: ${errorMessage(error)}`);
    } finally {
      busy = false;
    }
  }

  async function clearHistory() {
    if (!(await dialog.confirm('Clear all automation run history?', 'Clear History'))) return;
    try {
      await ClearAutomationHistory();
      history = [];
      selectedRun = null;
    } catch (error) {
      await dialog.alert(`Could not clear history: ${errorMessage(error)}`);
    }
  }

  function targetLabel(target) {
    return `${target.Name || target.Hostname || target.ID} (${target._kind})`;
  }

  function formatTime(value) {
    return value ? new Date(value).toLocaleString() : '-';
  }

  function triggerLabel(value) {
    return triggerOptions.find(([id]) => id === value)?.[1] || value;
  }
</script>

<div class="automation">
  <aside class="rules-sidebar">
    <div class="sidebar-header">
      <div>
        <h2>Automation</h2>
        <span>{rules.length} rule{rules.length === 1 ? '' : 's'}</span>
      </div>
      <button type="button" class="btn btn-primary" on:click={createRule}>
        <i class="fas fa-plus"></i> New
      </button>
    </div>

    <div class="rule-list">
      {#if rules.length === 0}
        <div class="empty-small">No automation rules yet.</div>
      {/if}
      {#each rules as rule (rule.id)}
        <button
          type="button"
          class="rule-card"
          class:selected={selectedID === rule.id}
          on:click={() => selectRule(rule)}
        >
          <span class="rule-state" class:on={rule.enabled}></span>
          <span class="rule-copy">
            <strong>{rule.name}</strong>
            <small>{triggerLabel(rule.trigger)} · {rule.targetKind || 'any'}</small>
          </span>
          <span
            class="mini-toggle"
            role="switch"
            tabindex="0"
            aria-checked={rule.enabled}
            title={rule.enabled ? 'Disable rule' : 'Enable rule'}
            on:click|stopPropagation={() => toggleRule(rule)}
            on:keydown|stopPropagation={(event) => (event.key === 'Enter' || event.key === ' ') && toggleRule(rule)}
          >{rule.enabled ? 'On' : 'Off'}</span>
        </button>
      {/each}
    </div>

    <button type="button" class="history-link" class:active={activePanel === 'history'} on:click={() => activePanel = 'history'}>
      <i class="fas fa-history"></i>
      <span>Run history</span>
      <strong>{history.length}</strong>
    </button>
  </aside>

  <section class="workspace">
    <div class="workspace-tabs">
      <button type="button" class:active={activePanel === 'rule'} on:click={() => activePanel = 'rule'}>Rule editor</button>
      <button type="button" class:active={activePanel === 'history'} on:click={() => activePanel = 'history'}>Execution history</button>
    </div>

    {#if loadError}
      <div class="load-error">{loadError}</div>
    {/if}

    {#if activePanel === 'rule'}
      {#if !draft}
        <div class="empty-main">
          <i class="fas fa-bolt"></i>
          <h3>Build repeatable agent workflows</h3>
          <p>Choose a rule or create one. Rules can run command sequences or Sobek JavaScript against matching sessions and beacons.</p>
          <button type="button" class="btn btn-primary" on:click={createRule}>Create automation</button>
        </div>
      {:else}
        <div class="editor">
          <header class="editor-header">
            <div>
              <h2>{draft.id ? draft.name : 'New automation'}</h2>
              <p>{draft.runCount || 0} completed run{draft.runCount === 1 ? '' : 's'}</p>
            </div>
            <label class="enabled-check">
              <input type="checkbox" bind:checked={draft.enabled} on:change={markDirty} />
              Enabled
            </label>
          </header>

          <div class="form-section">
            <h3>Identity</h3>
            <div class="form-grid two">
              <label>Name
                <input type="text" bind:value={draft.name} on:input={markDirty} />
              </label>
              <label>Description
                <input type="text" bind:value={draft.description} on:input={markDirty} placeholder="What this workflow does" />
              </label>
            </div>
          </div>

          <div class="form-section">
            <h3>Trigger and targets</h3>
            <div class="form-grid three">
              <label>Trigger
                <select bind:value={draft.trigger} on:change={markDirty}>
                  {#each triggerOptions as option}
                    <option value={option[0]}>{option[1]}</option>
                  {/each}
                </select>
              </label>
              <label>Target type
                <select bind:value={draft.targetKind} on:change={markDirty}>
                  <option value="any">Sessions and beacons</option>
                  <option value="session">Sessions only</option>
                  <option value="beacon">Beacons only</option>
                </select>
              </label>
              {#if draft.trigger === 'interval'}
                <label>Every (seconds)
                  <input type="number" min="10" bind:value={draft.intervalSeconds} on:input={markDirty} />
                </label>
              {/if}
            </div>

            <p class="hint">Filters accept comma-separated glob patterns such as <code>windows,linux</code> or <code>prod-*</code>. Empty fields match everything.</p>
            <div class="form-grid filters">
              <label>OS
                <input type="text" value={draft.filter.os} on:input={(event) => updateFilter('os', event.currentTarget.value)} placeholder="windows,linux" />
              </label>
              <label>Architecture
                <input type="text" value={draft.filter.arch} on:input={(event) => updateFilter('arch', event.currentTarget.value)} placeholder="amd64" />
              </label>
              <label>Hostname
                <input type="text" value={draft.filter.hostname} on:input={(event) => updateFilter('hostname', event.currentTarget.value)} placeholder="prod-*" />
              </label>
              <label>Username
                <input type="text" value={draft.filter.username} on:input={(event) => updateFilter('username', event.currentTarget.value)} placeholder="admin*" />
              </label>
              <label>Agent name
                <input type="text" value={draft.filter.name} on:input={(event) => updateFilter('name', event.currentTarget.value)} />
              </label>
            </div>
            <div class="match-summary">{matchingTargets.length} current target{matchingTargets.length === 1 ? '' : 's'} match this rule.</div>
          </div>

          <div class="form-section">
            <div class="workflow-heading">
              <div>
                <h3>Workflow</h3>
                <p class="hint">Choose a simple command sequence or a Sobek JavaScript workflow.</p>
              </div>
              <select bind:value={draft.executionMode} on:change={markDirty}>
                <option value="commands">Commands</option>
                <option value="javascript">JavaScript (Sobek)</option>
              </select>
            </div>
            {#if draft.executionMode === 'javascript'}
              <p class="hint">
                API: <code>sliver.run(command)</code>, <code>sliver.log(...values)</code>,
                <code>sliver.sleep(milliseconds)</code>, <code>console.log(...values)</code>.
                Read-only context is available as <code>target</code> and <code>trigger</code>.
              </p>
              <textarea
                class="commands script-editor"
                bind:value={draft.script}
                on:input={markDirty}
                spellcheck="false"
              ></textarea>
            {:else}
              <p class="hint">One Sliver command per line, executed in order. Templates: <code>{'{{id}}'}</code>, <code>{'{{name}}'}</code>, <code>{'{{hostname}}'}</code>, <code>{'{{username}}'}</code>, <code>{'{{os}}'}</code>, <code>{'{{arch}}'}</code>.</p>
              <textarea
                class="commands"
                value={draft.commands.join('\n')}
                on:input={(event) => updateCommands(event.currentTarget.value)}
                placeholder="info&#10;whoami&#10;ps"
                spellcheck="false"
              ></textarea>
            {/if}
            <div class="form-grid four">
              <label>Run timeout (s)
                <input type="number" min="1" max="3600" bind:value={draft.timeoutSeconds} on:input={markDirty} />
                <small>Includes waiting for beacon tasks</small>
              </label>
              {#if draft.executionMode !== 'javascript'}
                <label>Delay between steps (s)
                  <input type="number" min="0" bind:value={draft.delaySeconds} on:input={markDirty} />
                </label>
              {/if}
              <label>Per-target cooldown (s)
                <input type="number" min="0" bind:value={draft.cooldownSeconds} on:input={markDirty} />
              </label>
              <label>Maximum runs
                <input type="number" min="0" bind:value={draft.maxRuns} on:input={markDirty} />
                <small>0 means unlimited</small>
              </label>
              {#if draft.executionMode === 'commands'}
                <label class="checkbox-field">
                  <input type="checkbox" bind:checked={draft.continueOnError} on:change={markDirty} />
                  Continue after errors
                </label>
              {/if}
            </div>
          </div>

          <div class="editor-actions">
            {#if draft.id}
              <button type="button" class="btn btn-danger" on:click={deleteRule} disabled={busy}>Delete</button>
            {/if}
            <div class="action-spacer"></div>
            {#if draft.id}
              <select class="target-select" bind:value={manualTarget}>
                <option value="">All matching current targets</option>
                {#each matchingTargets as target}
                  <option value={target.ID}>{targetLabel(target)}</option>
                {/each}
              </select>
              <button type="button" class="btn" on:click={runNow} disabled={busy || matchingTargets.length === 0}>
                <i class="fas fa-play"></i> Run now
              </button>
            {/if}
            <button type="button" class="btn btn-primary" on:click={saveRule} disabled={busy || !draft.__dirty}>
              {busy ? 'Working...' : 'Save rule'}
            </button>
          </div>
        </div>
      {/if}
    {:else}
      <div class="history">
        <div class="history-header">
          <div>
            <h2>Execution history</h2>
            <p>Recent automated and manual runs. The newest run appears first.</p>
          </div>
          <button type="button" class="btn" on:click={clearHistory} disabled={history.length === 0}>Clear history</button>
        </div>
        <div class="history-split">
          <div class="run-list">
            {#if history.length === 0}
              <div class="empty-small">No automation runs recorded.</div>
            {/if}
            {#each history as run (run.id)}
              <button type="button" class="run-row" class:selected={selectedRun?.id === run.id} on:click={() => selectedRun = run}>
                <span class="status {run.status}"><i class="fas {run.status === 'completed' ? 'fa-check' : run.status === 'running' ? 'fa-spinner' : 'fa-times'}"></i></span>
                <span>
                  <strong>{run.ruleName}</strong>
                  <small>{run.targetName || 'No target'} · {run.trigger}</small>
                </span>
                <time>{formatTime(run.startedAt)}</time>
              </button>
            {/each}
          </div>
          <div class="run-detail">
            {#if selectedRun}
              <div class="run-title">
                <div>
                  <h3>{selectedRun.ruleName}</h3>
                  <p>{selectedRun.targetKind} {selectedRun.targetName} · {formatTime(selectedRun.startedAt)}</p>
                </div>
                <span class="status-label {selectedRun.status}">{selectedRun.status}</span>
              </div>
              {#if selectedRun.error}<div class="run-error">{selectedRun.error}</div>{/if}
              <pre>{selectedRun.output || (selectedRun.status === 'running' ? 'Running...' : 'No command output.')}</pre>
            {:else}
              <div class="empty-main compact">Select a run to inspect its output.</div>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  </section>
</div>

<style>
  .automation { display: flex; flex: 1; min-height: 0; background: var(--bg-color); }
  .rules-sidebar { width: 285px; display: flex; flex-direction: column; border-right: 1px solid var(--border-color); background: var(--toolbar-bg); }
  .sidebar-header { display: flex; align-items: center; justify-content: space-between; padding: 18px; border-bottom: 1px solid var(--border-color); }
  h2, h3, p { margin: 0; }
  .sidebar-header h2 { font-size: 1.15em; }
  .sidebar-header span, .editor-header p, .history-header p { color: var(--text-muted); font-size: .82em; }
  .rule-list { flex: 1; overflow: auto; padding: 8px; }
  .rule-card { width: 100%; display: flex; align-items: center; gap: 10px; padding: 11px 10px; border: 1px solid transparent; border-radius: 5px; background: transparent; color: var(--text-color); text-align: left; cursor: pointer; }
  .rule-card:hover { background: var(--row-hover); }
  .rule-card.selected { border-color: var(--accent-color); background: var(--panel-bg); }
  .rule-state { width: 8px; height: 8px; flex: 0 0 auto; border-radius: 50%; background: var(--text-muted); }
  .rule-state.on { background: var(--success-color); box-shadow: 0 0 7px color-mix(in srgb, var(--success-color) 70%, transparent); }
  .rule-copy { display: flex; min-width: 0; flex: 1; flex-direction: column; gap: 3px; }
  .rule-copy strong, .rule-copy small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .rule-copy small { color: var(--text-muted); }
  .mini-toggle { padding: 2px 6px; border: 1px solid var(--border-color); border-radius: 8px; color: var(--text-muted); font-size: .72em; }
  .history-link { display: flex; align-items: center; gap: 9px; padding: 13px 18px; border: 0; border-top: 1px solid var(--border-color); background: transparent; color: var(--text-muted); cursor: pointer; text-align: left; }
  .history-link:hover, .history-link.active { color: var(--accent-color); background: var(--panel-bg); }
  .history-link strong { margin-left: auto; }
  .workspace { flex: 1; display: flex; min-width: 0; flex-direction: column; }
  .workspace-tabs { display: flex; height: 42px; padding-left: 18px; border-bottom: 1px solid var(--border-color); background: var(--header-bg); }
  .workspace-tabs button { padding: 0 16px; border: 0; border-bottom: 2px solid transparent; background: transparent; color: var(--text-muted); cursor: pointer; }
  .workspace-tabs button.active { border-bottom-color: var(--accent-color); color: var(--accent-color); }
  .editor, .history { flex: 1; overflow: auto; }
  .editor { padding-bottom: 76px; }
  .editor-header, .history-header { display: flex; align-items: center; justify-content: space-between; padding: 22px 26px; border-bottom: 1px solid var(--border-color); }
  .editor-header h2, .history-header h2 { margin-bottom: 4px; font-size: 1.25em; }
  .enabled-check { display: flex; flex-direction: row; align-items: center; gap: 7px; color: var(--text-color); }
  .form-section { padding: 20px 26px; border-bottom: 1px solid var(--border-color); }
  .form-section h3 { margin-bottom: 14px; font-size: .95em; }
  .form-grid { display: grid; gap: 13px; }
  .form-grid.two { grid-template-columns: 1fr 1fr; }
  .form-grid.three { grid-template-columns: repeat(3, minmax(150px, 1fr)); }
  .form-grid.four { grid-template-columns: repeat(4, minmax(130px, 1fr)); margin-top: 13px; }
  .form-grid.filters { grid-template-columns: repeat(5, minmax(115px, 1fr)); }
  label { display: flex; flex-direction: column; gap: 5px; color: var(--text-muted); font-size: .82em; }
  input[type='text'], input[type='number'], select, textarea { box-sizing: border-box; width: 100%; border: 1px solid var(--border-color); border-radius: 4px; background: var(--toolbar-bg); color: var(--text-color); padding: 8px 10px; font: inherit; outline: none; }
  input:focus, select:focus, textarea:focus { border-color: var(--accent-color); }
  .hint { margin: 13px 0 10px; color: var(--text-muted); font-size: .8em; }
  code { color: var(--accent-color); font-family: var(--font-mono); }
  .match-summary { margin-top: 10px; color: var(--success-color); font-size: .82em; }
  .commands { min-height: 130px; resize: vertical; font-family: var(--font-mono); line-height: 1.55; }
  .script-editor { min-height: 280px; tab-size: 2; }
  .workflow-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
  .workflow-heading select { width: 190px; flex: 0 0 auto; }
  .workflow-heading .hint { margin: 4px 0 0; }
  .checkbox-field { flex-direction: row; align-items: center; align-self: end; min-height: 34px; color: var(--text-color); }
  .checkbox-field input { width: auto; }
  label small { color: var(--text-muted); }
  .editor-actions { position: sticky; bottom: 0; display: flex; align-items: center; gap: 9px; padding: 13px 26px; border-top: 1px solid var(--border-color); background: var(--header-bg); }
  .action-spacer { flex: 1; }
  .target-select { width: 245px; padding: 6px 9px; }
  .empty-main { display: flex; flex: 1; height: 100%; min-height: 280px; align-items: center; justify-content: center; flex-direction: column; gap: 10px; color: var(--text-muted); text-align: center; }
  .empty-main i { color: var(--accent-color); font-size: 2.2em; }
  .empty-main h3 { color: var(--text-color); }
  .empty-main p { max-width: 480px; }
  .empty-main.compact { min-height: 0; }
  .empty-small { padding: 24px 12px; color: var(--text-muted); text-align: center; }
  .load-error, .run-error { padding: 9px 14px; background: color-mix(in srgb, var(--danger-color) 12%, transparent); color: var(--danger-color); }
  .history { display: flex; min-height: 0; flex-direction: column; overflow: hidden; }
  .history-split { display: grid; flex: 1; min-height: 0; grid-template-columns: minmax(330px, 42%) 1fr; }
  .run-list { overflow: auto; border-right: 1px solid var(--border-color); }
  .run-row { width: 100%; display: grid; grid-template-columns: 25px 1fr auto; align-items: center; gap: 9px; padding: 11px 14px; border: 0; border-bottom: 1px solid var(--border-color); background: transparent; color: var(--text-color); text-align: left; cursor: pointer; }
  .run-row:hover, .run-row.selected { background: var(--row-hover); }
  .run-row span:nth-child(2) { display: flex; min-width: 0; flex-direction: column; gap: 3px; }
  .run-row small, .run-row time, .run-title p { color: var(--text-muted); font-size: .78em; }
  .run-row time { white-space: nowrap; }
  .status { display: flex; width: 20px; height: 20px; align-items: center; justify-content: center; border-radius: 50%; font-size: .7em; }
  .status.completed { background: color-mix(in srgb, var(--success-color) 18%, transparent); color: var(--success-color); }
  .status.failed { background: color-mix(in srgb, var(--danger-color) 18%, transparent); color: var(--danger-color); }
  .status.running { background: color-mix(in srgb, var(--accent-color) 18%, transparent); color: var(--accent-color); }
  .run-detail { min-width: 0; overflow: auto; padding: 20px; }
  .run-title { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px; }
  .status-label { padding: 3px 8px; border-radius: 10px; text-transform: capitalize; }
  .status-label.completed { color: var(--success-color); }
  .status-label.failed { color: var(--danger-color); }
  .status-label.running { color: var(--accent-color); }
  pre { min-height: 180px; margin: 0; padding: 15px; overflow: auto; border: 1px solid var(--border-color); border-radius: 4px; background: #111; color: #ddd; font-family: var(--font-mono); font-size: .86em; line-height: 1.45; white-space: pre-wrap; word-break: break-word; }
  @media (max-width: 1100px) {
    .form-grid.filters { grid-template-columns: repeat(3, 1fr); }
    .form-grid.four { grid-template-columns: repeat(2, 1fr); }
  }
</style>
