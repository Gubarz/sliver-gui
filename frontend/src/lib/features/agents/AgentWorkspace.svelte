<script>
  import { createEventDispatcher } from 'svelte';
  import { onMount } from 'svelte';

  import { KillAgent, RemoveBeacon, RenameAgent } from '../../api/agents.js';
  import { ListAutomationRules, RunAutomationRule } from '../../api/automation.js';
  import { CloseShell, GetCommandCatalog, StartShell } from '../../api/console.js';
  import { onAutomationUpdated } from '../../api/runtime.js';
  import { GetPivotListeners, GetPivots } from '../../api/server.js';
  import { GuiActionGroups } from '../../commands/GuiActions.js';
  import CommandModal from '../../components/ui/CommandModal.svelte';
  import { dispatchCommand } from '../../stores/consoleStore.js';
  import { dialog } from '../../stores/dialogStore.js';
  import { matchesAutomationTarget } from '../../utils/automation.js';
  import { isAgentOnline, pivotParentMap } from '../../utils/agents.js';
  import { errorMessage } from '../../utils/errors.js';
  import { startPolling } from '../../utils/polling.js';
  import ScreenshotGallery from '../gallery/ScreenshotGallery.svelte';
  import NetworkGraph from '../graph/NetworkGraph.svelte';
  import CredentialsPanel from '../server/CredentialsPanel.svelte';
  import EventsLog from '../server/EventsLog.svelte';
  import ArmoryPanel from '../server/ArmoryPanel.svelte';
  import GeneratePanel from '../server/GeneratePanel.svelte';
  import ListenersPanel from '../server/ListenersPanel.svelte';
  import LootPanel from '../server/LootPanel.svelte';
  import OperatorsPanel from '../server/OperatorsPanel.svelte';
  import PivotsPanel from '../server/PivotsPanel.svelte';

  import ContextMenu from './ContextMenu.svelte';
  import BeaconTasks from './BeaconTasks.svelte';
  import FileBrowser from './FileBrowser.svelte';
  import ProcessExplorer from './ProcessExplorer.svelte';
  import ScreenshotViewer from './ScreenshotViewer.svelte';
  import SessionConsole from './SessionConsole.svelte';
  import ShellTerminal from './ShellTerminal.svelte';
  import SessionsTable from './SessionsTable.svelte';
  import RegistryBrowser from './RegistryBrowser.svelte';

  export let sessions = [];
  export let beacons = [];
  export let eventHistory = [];
  export let refreshAgents = () => {};

  const dispatch = createEventDispatcher();

  let interactingSession = '';
  // On-demand, per-session tabs. Tabs only mount when explicitly opened (via
  // double-click for Console, or the right-click menu), stay mounted while the
  // workspace lives so re-selecting an agent restores state with no refetch,
  // and are individually closeable.
  let openTabsBySession = {}; // { [sessionID]: tabId[] }
  let activeTabBySession = {}; // { [sessionID]: tabId }
  let shellsByID = {};
  let eventsActive = false; // the global Events tab (not per-session)

  const TAB_META = {
    console: { icon: 'fas fa-terminal', label: 'Console' },
    tasks: { icon: 'fas fa-tasks', label: 'Tasks' },
    fileBrowser: { icon: 'fas fa-folder', label: 'Files' },
    processExplorer: { icon: 'fas fa-microchip', label: 'Processes' },
    registryBrowser: { icon: 'fas fa-database', label: 'Registry' },
    screenshot: { icon: 'fas fa-desktop', label: 'Screenshot' },
  };
  const TAB_IDS = Object.keys(TAB_META);

  function tabMeta(tab) {
    if (tab.startsWith('shell-')) {
      const shell = shellsByID[tab];
      return { icon: 'fas fa-terminal', label: shell?.label || 'Shell' };
    }
    return TAB_META[tab];
  }

  function openTab(sid, tab) {
    if (!sid || !TAB_META[tab]) return;
    interactingSession = sid;
    const tabs = openTabsBySession[sid] || [];
    if (!tabs.includes(tab)) {
      openTabsBySession = { ...openTabsBySession, [sid]: [...tabs, tab] };
    }
    activeTabBySession = { ...activeTabBySession, [sid]: tab };
    eventsActive = false;
  }

  function selectTab(tab) {
    eventsActive = false;
    activeTabBySession = { ...activeTabBySession, [interactingSession]: tab };
  }

  function closeTab(sid, tab) {
    const tabs = (openTabsBySession[sid] || []).filter((t) => t !== tab);
    openTabsBySession = { ...openTabsBySession, [sid]: tabs };
    if (activeTabBySession[sid] === tab) {
      activeTabBySession = { ...activeTabBySession, [sid]: tabs[tabs.length - 1] || null };
    }
    if (tab.startsWith('shell-')) {
      CloseShell(tab).catch(() => {});
      const remainingShells = { ...shellsByID };
      delete remainingShells[tab];
      shellsByID = remainingShells;
    }
  }

  async function openShell(sid) {
    if (!sid) return;
    try {
      const shell = await StartShell(sid, '', true, 24, 100);
      const shellNumber = Object.values(shellsByID).filter(
        (item) => item.sessionID === sid,
      ).length + 1;
      shellsByID = {
        ...shellsByID,
        [shell.id]: { ...shell, label: `Shell ${shellNumber}` },
      };
      interactingSession = sid;
      openTabsBySession = {
        ...openTabsBySession,
        [sid]: [...(openTabsBySession[sid] || []), shell.id],
      };
      activeTabBySession = { ...activeTabBySession, [sid]: shell.id };
      eventsActive = false;
    } catch (error) {
      await dialog.alert(`Could not open shell: ${error}`);
    }
  }

  // Which tabs make sense for a given agent (beacons can't live-browse).
  function coreActionsFor(agent) {
    if (!agent) return [];
    const acts = [
      { id: 'console', icon: TAB_META.console.icon, label: 'Console' },
    ];
    if (agent._kind === 'beacon') {
      acts.push({ id: 'tasks', icon: TAB_META.tasks.icon, label: 'Tasks' });
    } else {
      acts.splice(1, 0, { id: 'newShell', icon: 'fas fa-terminal', label: 'New Shell' });
      acts.push({ id: 'fileBrowser', icon: TAB_META.fileBrowser.icon, label: 'File Browser' });
      acts.push({ id: 'processExplorer', icon: TAB_META.processExplorer.icon, label: 'Process Explorer' });
      if ((agent.OS || '').toLowerCase() === 'windows') {
        acts.push({ id: 'registryBrowser', icon: TAB_META.registryBrowser.icon, label: 'Registry' });
      }
      acts.push({ id: 'screenshot', icon: TAB_META.screenshot.icon, label: 'Take Screenshot' });
    }
    return acts;
  }

  let activeCommand = null;
  let activeCommandUsesSession = false;
  let topPaneHeight = 50;
  let dragging = false;

  let showContextMenu = false;
  let contextMenuX = 0;
  let contextMenuY = 0;
  let contextMenuTarget = '';
  let showGlobalMenu = null;

  let showListeners = false;
  let showGenerate = false;
  let showLoot = false;
  let showCreds = false;
  let showPivots = false;
  let showOperators = false;
  let showGallery = false;
  let showArmory = false;
  let sessionCategories = [];
  let serverCategories = [];
  let automationRules = [];
  let pivotGraph = null;
  let pivotListeners = [];
  let agentFilter = '';
  let agentViewMode = 'table';

  $: combinedData = [
    ...beacons.map((beacon) => ({ ...beacon, _kind: 'beacon' })),
    ...sessions.map((session) => ({ ...session, _kind: 'session' })),
  ];
  $: filteredData = combinedData.filter((agent) => {
    const needle = agentFilter.trim().toLowerCase();
    if (!needle) return true;
    return [
      agent.ID, agent.Name, agent.Hostname, agent.Username, agent.OS, agent.Arch,
      agent.Transport, agent.RemoteAddress, agent.Filename, agent.ProcessName, agent.PID,
      isAgentOnline(agent) ? 'online' : 'offline',
      agent._kind,
    ].some((value) => String(value || '').toLowerCase().includes(needle));
  });
  $: graphData = (() => {
    if (!agentFilter.trim()) return combinedData;

    const parents = pivotParentMap(pivotGraph);
    const visibleIDs = new Set(filteredData.map((agent) => agent.ID));

    for (const agent of filteredData) {
      let parentID = parents.get(agent.ID);
      while (parentID && !visibleIDs.has(parentID)) {
        visibleIDs.add(parentID);
        parentID = parents.get(parentID);
      }
    }

    return combinedData.filter((agent) => visibleIDs.has(agent.ID));
  })();
  $: globalCategories = [...GuiActionGroups, ...serverCategories];
  // Tabs available in the right-click menu depend on the agent under the cursor.
  $: contextAgent = combinedData.find((a) => a.ID === contextMenuTarget);
  $: contextCoreActions = coreActionsFor(contextAgent);
  $: contextDangerActions = dangerActionsFor(contextAgent);
  $: matchingAutomationRules = contextAgent
    ? automationRules.filter((rule) => rule.enabled && matchesAutomationTarget(contextAgent, rule))
    : [];
  $: contextCategories = [
    ...(matchingAutomationRules.length > 0 ? [{
      category: 'Automation',
      commands: matchingAutomationRules.map((rule) => ({
        name: rule.name,
        description: rule.description || 'Run this automation against the selected agent.',
        supported: true,
        automationRuleID: rule.id,
      })),
    }] : []),
    ...sessionCategories,
  ];

  function dangerActionsFor(agent) {
    const actions = [{ id: 'kill', icon: 'fas fa-skull', label: 'Kill Agent' }];
    if (agent?._kind === 'beacon') {
      actions.push({ id: 'removeBeacon', icon: 'fas fa-trash-alt', label: 'Remove Beacon Record' });
    }
    return actions;
  }

  onMount(async () => {
    try {
      const [sessionCatalog, serverCatalog] = await Promise.all([
        GetCommandCatalog('session'),
        GetCommandCatalog('server'),
      ]);
      sessionCategories = catalogCategories(sessionCatalog);
      serverCategories = catalogCategories(serverCatalog);
    } catch (error) {
      await dialog.alert(`Could not load Sliver commands: ${error}`);
    }
  });

  onMount(() => {
    try {
      const storedMode = localStorage.getItem('sliver_agent_view_mode');
      if (storedMode === 'table' || storedMode === 'graph') agentViewMode = storedMode;
    } catch {}

    refreshAutomationRules();
    const stop = onAutomationUpdated(refreshAutomationRules);
    return stop;
  });

  onMount(() => startPolling(refreshPivotGraph, 5000));

  async function refreshAutomationRules() {
    try {
      automationRules = await ListAutomationRules();
    } catch (error) {
      console.error('Could not load automation rules:', error);
    }
  }

  async function refreshPivotGraph() {
    try {
      [pivotGraph, pivotListeners] = await Promise.all([
        GetPivots(),
        GetPivotListeners(),
      ]);
    } catch (error) {
      console.error('Could not load pivot graph:', error);
    }
  }

  function setAgentViewMode(mode) {
    agentViewMode = mode;
    try {
      localStorage.setItem('sliver_agent_view_mode', mode);
    } catch {}
  }

  function catalogCategories(catalog) {
    return (catalog?.groups ?? []).map((group) => ({
      category: group.title,
      commands: group.commands,
    }));
  }

  function startDrag() {
    dragging = true;
  }

  function resizeWithKeyboard(event) {
    if (event.key !== 'ArrowUp' && event.key !== 'ArrowDown') return;
    event.preventDefault();
    const delta = event.key === 'ArrowUp' ? -5 : 5;
    topPaneHeight = Math.max(10, Math.min(90, topPaneHeight + delta));
  }

  function stopDrag() {
    dragging = false;
  }

  function onDrag(event) {
    if (!dragging) return;
    const headerHeight = 60;
    const totalHeight = window.innerHeight - headerHeight - 25;
    const newHeight = ((event.clientY - headerHeight) / totalHeight) * 100;
    if (newHeight > 10 && newHeight < 90) topPaneHeight = newHeight;
  }

  function handleContextMenu(event) {
    const detail = event.detail;
    openAgentContextMenu(detail.event, detail.session);
  }

  function openAgentContextMenu(nativeEvent, agent) {
    showContextMenu = true;
    contextMenuX = nativeEvent.clientX;
    contextMenuY = nativeEvent.clientY;
    contextMenuTarget = agent.ID;
  }

  function handleMenuSelect(event) {
    const payload = event.detail;
    showContextMenu = false;

    if (contextMenuTarget) interactingSession = contextMenuTarget;

    if (payload.type === 'core') {
      const action = payload.action;
      if (TAB_IDS.includes(action)) {
        openTab(contextMenuTarget, action);
      } else if (action === 'newShell') {
        openShell(contextMenuTarget);
      } else if (action === 'kill') {
        killAgent(contextMenuTarget);
      } else if (action === 'removeBeacon') {
        removeBeacon(contextMenuTarget);
      } else if (action === 'rename') {
        renameAgent(contextMenuTarget);
      }
      return;
    }

    if (payload.command.automationRuleID) {
      runAutomation(payload.command.automationRuleID, contextMenuTarget);
      return;
    }

    selectCommand(payload.command, true);
  }

  async function runAutomation(ruleID, targetID) {
    try {
      await RunAutomationRule(ruleID, targetID);
    } catch (error) {
      await dialog.alert(`Could not run automation: ${errorMessage(error)}`);
    }
  }

  function handleGlobalMenuSelect(item) {
    showGlobalMenu = null;
    if (item.action) {
      executeGuiAction(item.action);
      return;
    }
    selectCommand(item, false);
  }

  function selectCommand(command, useSession) {
    if (!command.supported) {
      activeCommand = command;
      activeCommandUsesSession = useSession;
      return;
    }
    if (!command.needsInput) {
      executeSliverCommand(command.path, useSession);
      return;
    }
    activeCommand = command;
    activeCommandUsesSession = useSession;
  }

  async function killAgent(id) {
    if (!id || !(await dialog.confirm('Kill this agent?', 'Confirm Kill'))) return;
    try {
      await KillAgent(id);
      closeAgentWorkspace(id);
      refreshAgents();
    } catch (error) {
      await dialog.alert(`Kill failed: ${error}`);
    }
  }

  async function removeBeacon(id) {
    if (!id || !(await dialog.confirm(
      'Remove this beacon record? This also permanently deletes all tasks associated with it.',
      'Remove Beacon Record',
    ))) return;
    try {
      await RemoveBeacon(id);
      closeAgentWorkspace(id);
      refreshAgents();
    } catch (error) {
      await dialog.alert(`Remove failed: ${error}`);
    }
  }

  function closeAgentWorkspace(id) {
    if (interactingSession === id) interactingSession = '';
    for (const tab of openTabsBySession[id] || []) {
      if (tab.startsWith('shell-')) CloseShell(tab).catch(() => {});
    }
    const restOpen = { ...openTabsBySession };
    const restActive = { ...activeTabBySession };
    const remainingShells = { ...shellsByID };
    for (const [shellID, shell] of Object.entries(shellsByID)) {
      if (shell.sessionID === id) delete remainingShells[shellID];
    }
    delete restOpen[id];
    delete restActive[id];
    openTabsBySession = restOpen;
    activeTabBySession = restActive;
    shellsByID = remainingShells;
  }

  async function renameAgent(id) {
    if (!id) return;
    const name = await dialog.prompt('New name for this agent:', 'Rename Agent');
    if (!name) return;
    try {
      await RenameAgent(id, name);
      refreshAgents();
    } catch (error) {
      await dialog.alert(`Rename failed: ${error}`);
    }
  }

  function navigate(view, serverTab = '') {
    dispatch('navigate', { view, serverTab });
  }

  function executeSliverCommand(cmd, useSession) {
    if (useSession && interactingSession) {
      openTab(interactingSession, 'console');
      dispatchCommand(interactingSession, cmd);
    } else {
      dispatchCommand('', cmd);
      // User explicitly requested: "if i'm on the agents tab IT SHOULD NEVER EVER SWITCH ME TO THE SERVER TAB"
    }
    activeCommand = null;
  }

  function executeGuiAction(action) {
    if (action === 'listeners') showListeners = true;
    else if (action === 'generate') showGenerate = true;
    else if (action === 'loot') showLoot = true;
    else if (action === 'credentials') showCreds = true;
    else if (action === 'pivots') showPivots = true;
    else if (action === 'operators') showOperators = true;
    else if (action === 'graph') setAgentViewMode('graph');
    else if (action === 'gallery') showGallery = true;
    else if (action === 'events') eventsActive = true;
    else if (action === 'settings') navigate('settings');
    else if (action === 'profiles') { /* do nothing, keep user on agents tab */ }
    else if (action === 'armory') showArmory = true;
  }

  function hideMenus() {
    showContextMenu = false;
    showGlobalMenu = null;
  }
</script>

<svelte:window
  on:click={hideMenus}
  on:mousemove={onDrag}
  on:mouseup={stopDrag}
/>

<div class="top-menu-bar">
  {#each globalCategories as category}
    <div class="top-menu-item">
      <button
        type="button"
        class="top-menu-trigger"
        on:click|stopPropagation={() => showGlobalMenu = category.category}
      >
        {category.category}
      </button>
      {#if showGlobalMenu === category.category}
        <div class="dropdown-menu">
          {#each category.commands as cmd}
            <button
              type="button"
              class="dropdown-item"
              on:click|stopPropagation={() => handleGlobalMenuSelect(cmd)}
            >
              {cmd.name}
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/each}
</div>

<div class="toolbar">
  <button type="button" class="tool-icon" title="Listeners" on:click={() => showListeners = true}><i class="fas fa-headphones"></i></button>
  <button type="button" class="tool-icon" title="Armory" on:click={() => executeGuiAction('armory')}><i class="fas fa-shield-alt"></i></button>
  <button type="button" class="tool-icon" title="Screenshot Gallery" on:click={() => executeGuiAction('gallery')}><i class="fas fa-images"></i></button>
  <button type="button" class="tool-icon" title="Credentials" on:click={() => showCreds = true}><i class="fas fa-key"></i></button>
  <button type="button" class="tool-icon" title="Loot" on:click={() => showLoot = true}><i class="fas fa-download"></i></button>
  <button type="button" class="tool-icon" title="Generate Implant" on:click={() => showGenerate = true}><i class="fas fa-industry"></i></button>
</div>

<div class="main-split">
  <div class="top-pane" style="flex: 0 0 {topPaneHeight}%;">
    <div class="agent-view-toolbar">
      <i class="fas fa-search"></i>
      <input class="search" placeholder="Filter agents..." bind:value={agentFilter} />
      <label for="agent-view-mode">View</label>
      <select
        id="agent-view-mode"
        class="view-select"
        value={agentViewMode}
        on:change={(event) => setAgentViewMode(event.currentTarget.value)}
      >
        <option value="table">Table</option>
        <option value="graph">Graph</option>
      </select>
      <span class="count">{filteredData.length} agent{filteredData.length === 1 ? '' : 's'}</span>
    </div>
    <div class="agent-view-content">
      {#if agentViewMode === 'table'}
        <SessionsTable
          data={filteredData}
          {pivotGraph}
          selectedId={interactingSession}
          filterable={false}
          on:contextmenu={handleContextMenu}
          on:select={(event) => interactingSession = event.detail}
          on:interact={(event) => openTab(event.detail, 'console')}
        />
      {:else}
        <NetworkGraph
          embedded
          sessions={graphData.filter((agent) => agent._kind === 'session')}
          beacons={graphData.filter((agent) => agent._kind === 'beacon')}
          {pivotGraph}
          {pivotListeners}
          onSelect={(id) => interactingSession = id}
          onInteract={(id) => openTab(id, 'console')}
          onContextMenu={openAgentContextMenu}
        />
      {/if}
    </div>
  </div>

  <button
    type="button"
    class="resizer"
    aria-label="Resize sessions and interaction panes"
    on:mousedown={startDrag}
    on:keydown={resizeWithKeyboard}
  ></button>

  <div class="bottom-pane">
    <div class="bottom-tabs">
      {#if interactingSession}
        {#each (openTabsBySession[interactingSession] || []) as tab (tab)}
          <button
            type="button"
            class="bottom-tab"
            class:active={!eventsActive && activeTabBySession[interactingSession] === tab}
            on:click={() => selectTab(tab)}
          >
            <i class={tabMeta(tab).icon}></i> {tabMeta(tab).label}
            <span
              class="tab-close"
              role="button"
              tabindex="0"
              title="Close tab"
              on:click|stopPropagation={() => closeTab(interactingSession, tab)}
              on:keydown|stopPropagation={(e) => (e.key === 'Enter' || e.key === ' ') && closeTab(interactingSession, tab)}
            >×</span>
          </button>
        {/each}
      {/if}
      <div class="tab-spacer"></div>
      <button type="button" class="bottom-tab" class:active={eventsActive} on:click={() => eventsActive = true}><i class="fas fa-list"></i> Events</button>
    </div>

    <div class="bottom-content">
      <div class="tab-content" class:hidden={!eventsActive}>
        <EventsLog events={eventHistory} />
      </div>
      <!-- Every opened tab for every agent stays mounted (hidden when inactive)
           so switching agents/tabs preserves state without refetching. -->
      {#each Object.entries(openTabsBySession) as [sid, tabs] (sid)}
        {#each tabs as tab (tab)}
          <div class="tab-content" class:hidden={eventsActive || sid !== interactingSession || activeTabBySession[sid] !== tab}>
            {#if tab === 'console'}<SessionConsole sessionID={sid} on:shell={() => openShell(sid)} />{/if}
            {#if tab === 'tasks'}<BeaconTasks beaconID={sid} active={!eventsActive && sid === interactingSession && activeTabBySession[sid] === 'tasks'} />{/if}
            {#if tab.startsWith('shell-') && shellsByID[tab]}<ShellTerminal shell={shellsByID[tab]} />{/if}
            {#if tab === 'fileBrowser'}<FileBrowser sessionID={sid} />{/if}
            {#if tab === 'processExplorer'}<ProcessExplorer sessionID={sid} on:command={(e) => { openTab(sid, 'console'); dispatchCommand(sid, e.detail.cmd); }} />{/if}
            {#if tab === 'registryBrowser'}<RegistryBrowser sessionID={sid} />{/if}
            {#if tab === 'screenshot'}<ScreenshotViewer sessionID={sid} />{/if}
          </div>
        {/each}
      {/each}
      {#if !eventsActive && interactingSession && (openTabsBySession[interactingSession] || []).length === 0}
        <div class="empty-state">Double-click the agent to open a console, or right-click for more tabs.</div>
      {:else if !eventsActive && !interactingSession}
        <div class="empty-state">Select an agent above. Double-click to open a console, or right-click for more tabs.</div>
      {/if}
    </div>
  </div>
</div>

{#if showListeners}<ListenersPanel on:close={() => showListeners = false} />{/if}
{#if showGenerate}<GeneratePanel on:close={() => showGenerate = false} />{/if}
{#if showLoot}<LootPanel on:close={() => showLoot = false} />{/if}
{#if showCreds}<CredentialsPanel on:close={() => showCreds = false} />{/if}
{#if showPivots}<PivotsPanel on:close={() => showPivots = false} />{/if}
{#if showOperators}<OperatorsPanel on:close={() => showOperators = false} />{/if}
{#if showGallery}<ScreenshotGallery on:close={() => showGallery = false} />{/if}
{#if showArmory}
  <ArmoryPanel on:close={() => showArmory = false} />
{/if}
{#if activeCommand}
  <CommandModal
    sessionID={activeCommandUsesSession ? interactingSession : ''}
    command={activeCommand}
    on:close={() => activeCommand = null}
    on:execute={(event) => executeSliverCommand(event.detail.cmd, activeCommandUsesSession)}
  />
{/if}
{#if showContextMenu}
  <ContextMenu
    x={contextMenuX}
    y={contextMenuY}
    coreActions={contextCoreActions}
    dangerActions={contextDangerActions}
    categories={contextCategories}
    on:select={handleMenuSelect}
  />
{/if}

<style>
  .agent-view-toolbar {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    border-bottom: 1px solid var(--border-color);
    background: var(--header-bg);
    color: var(--text-muted);
    font-size: 0.85em;
  }

  .agent-view-toolbar label {
    margin-left: auto;
  }

  .agent-view-toolbar .count {
    margin-left: 4px;
    min-width: 64px;
    text-align: right;
  }

  .view-select {
    min-width: 92px;
    padding: 4px 26px 4px 8px;
    border: 1px solid var(--border-color);
    border-radius: 3px;
    background: var(--bg-color);
    color: var(--text-color);
  }

  .agent-view-content {
    flex: 1;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
  }

  .bottom-pane {
    flex: 1;
  }

  .tab-spacer {
    flex: 1;
  }

  .tab-close {
    margin-left: 8px;
    padding: 0 4px;
    border-radius: 3px;
    opacity: 0.6;
    font-size: 1.1em;
    line-height: 1;
  }
  .tab-close:hover {
    opacity: 1;
    background: var(--danger-color, #ff4a4a);
    color: #fff;
  }

  .tab-content {
    display: flex;
    height: 100%;
    flex-direction: column;
  }

  .tab-content.hidden {
    display: none;
  }

  .empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
  }

  .dropdown-menu {
    position: absolute;
    top: 25px;
    left: 0;
    background-color: var(--panel-bg);
    border: 1px solid var(--panel-border);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    border-radius: 4px;
    padding: 5px 0;
    min-width: 200px;
    z-index: 2000;
    max-height: 80vh;
    overflow-y: auto;
  }

  .dropdown-item {
    display: block;
    width: 100%;
    padding: 8px 15px;
    border: 0;
    background: transparent;
    color: var(--text-color);
    text-align: left;
    font-size: 0.9em;
    cursor: pointer;
  }

  .dropdown-item:hover {
    background-color: var(--accent-color);
    color: #fff;
  }

  .top-menu-item {
    position: relative;
  }

  .top-menu-trigger {
    border: 0;
    background: transparent;
    color: inherit;
    font: inherit;
    cursor: pointer;
  }
</style>
