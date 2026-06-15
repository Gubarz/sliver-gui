<script>
  import { createEventDispatcher } from 'svelte';
  import { onMount } from 'svelte';

  import { KillAgent, RemoveBeacon, RenameAgent } from '../../api/agents.js';
  import { ListAutomationRules, RunAutomationRule } from '../../api/automation.js';
  import { CloseShell, GetCommandCatalog, StartShell } from '../../api/console.js';
  import {
    ClearNetworkDiscoveries,
    DiscoverNetwork,
    GetNetworkDiscoveries,
    RemoveNetworkDiscoveries,
  } from '../../api/discovery.js';
  import {
    copyText,
    onAutomationUpdated,
    onNetworkDiscoveryUpdated,
  } from '../../api/runtime.js';
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

  async function openShell(sid, showError = true) {
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
      if (!showError) throw error;
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
    acts.push(
      { id: 'discoverNeighbors', icon: 'fas fa-network-wired', label: 'Discover Neighbors' },
      { id: 'discoverSweep', icon: 'fas fa-search-location', label: 'Ping Sweep...' },
      { id: 'clearDiscoveries', icon: 'fas fa-eraser', label: 'Clear Discovered Devices' },
    );
    return acts;
  }

  let activeCommand = null;
  let activeCommandUsesSession = false;
  let activeCommandTargetIDs = [];
  let topPaneHeight = 50;
  let dragging = false;

  let showContextMenu = false;
  let contextMenuX = 0;
  let contextMenuY = 0;
  let contextMenuTarget = '';
  let contextMenuKind = 'agent';
  let contextDevice = null;
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
  let networkDiscoveries = [];
  $: discoveredHosts = dedupeDiscoveries(networkDiscoveries);
  let selectedItemKeys = [];
  $: selectedAgentIDs = selectedItemKeys
    .filter((key) => key.startsWith('agent:'))
    .map((key) => key.slice(6));
  $: selectedDiscoveryKeys = selectedItemKeys
    .filter((key) => key.startsWith('device:'))
    .map((key) => key.slice(7));
  let agentFilter = '';
  let agentViewMode = 'table';

  $: combinedData = [
    ...beacons.map((beacon) => ({ ...beacon, _kind: 'beacon' })),
    ...sessions.map((session) => ({ ...session, _kind: 'session' })),
  ];
  $: filteredData = combinedData.filter((agent) => {
    const needle = agentFilter.trim().toLowerCase();
    if (!needle) return true;
    const agentMatches = [
      agent.ID, agent.Name, agent.Hostname, agent.Username, agent.OS, agent.Arch,
      agent.Transport, agent.RemoteAddress, agent.Filename, agent.ProcessName, agent.PID,
      isAgentOnline(agent) ? 'online' : 'offline',
      agent._kind,
    ].some((value) => String(value || '').toLowerCase().includes(needle));
    const deviceMatches = networkDiscoveries
      .filter((device) => device.agentID === agent.ID)
      .some((device) => [
        device.ip, device.mac, device.hostname, device.vendor, device.osHint,
        device.method, 'device', 'discovered',
      ].some((value) => String(value || '').toLowerCase().includes(needle)));
    return agentMatches || deviceMatches;
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
  $: contextDeviceActions = contextDevice ? [
    { id: 'copyDeviceIP', icon: 'fas fa-copy', label: 'Copy IP Address' },
    ...(contextDevice.mac
      ? [{ id: 'copyDeviceMAC', icon: 'fas fa-copy', label: 'Copy MAC Address' }]
      : []),
  ] : [];
  $: contextDeviceDangerActions = contextDevice ? [{
    id: 'removeDiscoveries',
    icon: 'fas fa-trash-alt',
    label: selectedDiscoveryKeys.length > 1
      ? `Remove Selected (${selectedDiscoveryKeys.length})`
      : 'Remove Discovered Device',
  }] : [];
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

  onMount(() => {
    refreshNetworkDiscoveries();
    return onNetworkDiscoveryUpdated((devices) => {
      networkDiscoveries = devices || [];
      pruneDiscoverySelection();
    });
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

  async function refreshNetworkDiscoveries() {
    try {
      networkDiscoveries = await GetNetworkDiscoveries();
      pruneDiscoverySelection();
    } catch (error) {
      console.error('Could not load network discoveries:', error);
    }
  }

  function discoveryKey(device) {
    return device.key || hostIdentity(device);
  }

  function hostIdentity(device) {
    const mac = String(device.mac || '').trim().toLowerCase().replaceAll('-', ':');
    return mac ? `mac:${mac}` : `ip:${device.ip}`;
  }

  function dedupeDiscoveries(devices) {
    const hosts = new Set();
    const byMAC = new Map();
    const byIP = new Map();

    function mergeHosts(target, source) {
      if (!source || target === source) return target;
      for (const observation of source.observations) {
        if (!target.observations.some(
          (item) => item.agentID === observation.agentID && item.ip === observation.ip,
        )) {
          target.observations.push(observation);
        }
      }
      for (const observerID of source.observerIDs) {
        if (!target.observerIDs.includes(observerID)) target.observerIDs.push(observerID);
      }
      if (!target.hostname && source.hostname) target.hostname = source.hostname;
      if (!target.vendor && source.vendor) target.vendor = source.vendor;
      if (!target.osHint && source.osHint) {
        target.osHint = source.osHint;
        target.ttl = source.ttl;
      }
      if (source.lastSeen > target.lastSeen) {
        target.lastSeen = source.lastSeen;
        target.method = source.method || target.method;
      }
      for (const [key, host] of byMAC) {
        if (host === source) byMAC.set(key, target);
      }
      for (const [key, host] of byIP) {
        if (host === source) byIP.set(key, target);
      }
      hosts.delete(source);
      return target;
    }

    for (const device of devices || []) {
      const mac = String(device.mac || '').trim().toLowerCase().replaceAll('-', ':');
      const macMatch = mac ? byMAC.get(mac) : null;
      const ipMatch = byIP.get(device.ip);
      let existing = mergeHosts(macMatch || ipMatch, macMatch && ipMatch ? ipMatch : null);
      const observation = { agentID: device.agentID, ip: device.ip };
      if (!existing) {
        existing = {
          ...device,
          key: hostIdentity(device),
          observations: [observation],
          observerIDs: [device.agentID],
        };
        hosts.add(existing);
      } else {
        if (!existing.observations.some(
          (item) => item.agentID === observation.agentID && item.ip === observation.ip,
        )) {
          existing.observations.push(observation);
        }
        if (!existing.observerIDs.includes(device.agentID)) {
          existing.observerIDs.push(device.agentID);
        }
        if (!existing.hostname && device.hostname) existing.hostname = device.hostname;
        if (!existing.vendor && device.vendor) existing.vendor = device.vendor;
        if (!existing.osHint && device.osHint) {
          existing.osHint = device.osHint;
          existing.ttl = device.ttl;
        }
        if (device.lastSeen > existing.lastSeen) {
          existing.lastSeen = device.lastSeen;
          existing.method = device.method || existing.method;
        }
      }

      if (mac) {
        existing.mac = mac;
        existing.key = `mac:${mac}`;
        byMAC.set(mac, existing);
      }
      byIP.set(device.ip, existing);
    }

    return [...hosts].map((host) => ({
      ...host,
      observerIDs: [...host.observerIDs].sort(),
      observations: [...host.observations].sort((a, b) =>
        a.agentID.localeCompare(b.agentID) || a.ip.localeCompare(b.ip)),
    }));
  }

  function pruneDiscoverySelection() {
    const available = new Set(dedupeDiscoveries(networkDiscoveries).map(discoveryKey));
    setSelectedItems(selectedItemKeys.filter(
      (key) => !key.startsWith('device:') || available.has(key.slice(7)),
    ));
  }

  function selectAgent({ id, ids, additive }) {
    if (ids) {
      replaceSelection(ids, selectedDiscoveryKeys);
      return;
    }
    const key = `agent:${id}`;
    if (!additive) {
      selectedItemKeys = [key];
      interactingSession = id;
      return;
    }
    toggleSelection(key);
  }

  function selectDiscovery({ key, keys, additive }) {
    if (keys) {
      replaceSelection(selectedAgentIDs, keys);
      return;
    }
    const itemKey = `device:${key}`;
    if (!additive) {
      selectedItemKeys = [itemKey];
      return;
    }
    toggleSelection(itemKey);
  }

  function toggleSelection(key) {
    setSelectedItems(selectedItemKeys.includes(key)
      ? selectedItemKeys.filter((selected) => selected !== key)
      : [...selectedItemKeys, key]);
  }

  function replaceSelection(agentIDs, deviceKeys) {
    setSelectedItems([
      ...agentIDs.map((id) => `agent:${id}`),
      ...deviceKeys.map((key) => `device:${key}`),
    ]);
  }

  function setSelectedItems(nextKeys) {
    const unique = [...new Set(nextKeys)];
    if (
      unique.length === selectedItemKeys.length &&
      unique.every((key) => selectedItemKeys.includes(key))
    ) return;
    selectedItemKeys = unique;
  }

  function syncGraphSelection({ agentIDs, deviceKeys }) {
    replaceSelection(agentIDs, deviceKeys);
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
    if (!selectedAgentIDs.includes(agent.ID)) {
      selectedItemKeys = [`agent:${agent.ID}`];
    }
    interactingSession = agent.ID;
    showContextMenu = true;
    contextMenuX = nativeEvent.clientX;
    contextMenuY = nativeEvent.clientY;
    contextMenuTarget = agent.ID;
    contextMenuKind = 'agent';
    contextDevice = null;
  }

  function openDeviceContextMenu(nativeEvent, deviceData) {
    const host = discoveredHosts.find((item) => discoveryKey(item) === deviceData.key);
    if (!host) return;
    if (!selectedDiscoveryKeys.includes(deviceData.key)) {
      selectedItemKeys = [`device:${deviceData.key}`];
    }
    showContextMenu = true;
    contextMenuX = nativeEvent.clientX;
    contextMenuY = nativeEvent.clientY;
    contextMenuTarget = '';
    contextMenuKind = 'device';
    contextDevice = host;
  }

  function handleMenuSelect(event) {
    const payload = event.detail;
    showContextMenu = false;

    if (payload.type === 'core') {
      const action = payload.action;
      if (contextMenuKind === 'device') {
        handleDeviceAction(action);
        return;
      }
      const targetIDs = commandTargetIDs();
      if (TAB_IDS.includes(action)) {
        openTabsForTargets(targetIDs, action);
      } else if (action === 'newShell') {
        openShellsForTargets(targetIDs);
      } else if (action === 'kill') {
        killAgents(targetIDs);
      } else if (action === 'removeBeacon') {
        removeBeacons(targetIDs);
      } else if (action === 'rename') {
        renameAgents(targetIDs);
      } else if (action === 'discoverNeighbors') {
        discoverNetworks(targetIDs, 'arp');
      } else if (action === 'discoverSweep') {
        promptNetworkSweep(targetIDs);
      } else if (action === 'clearDiscoveries') {
        clearNetworkDiscoveries(targetIDs);
      }
      return;
    }

    if (payload.command.automationRuleID) {
      runAutomation(payload.command.automationRuleID, commandTargetIDs());
      return;
    }

    selectCommand(payload.command, true, commandTargetIDs());
  }

  function commandTargetIDs() {
    return selectedAgentIDs.includes(contextMenuTarget)
      ? [...selectedAgentIDs]
      : [contextMenuTarget].filter(Boolean);
  }

  function agentsForIDs(targetIDs) {
    const wanted = new Set(targetIDs);
    return combinedData.filter((agent) => wanted.has(agent.ID));
  }

  function compatibleTargets(targetIDs, action) {
    const agents = agentsForIDs(targetIDs);
    if (action === 'tasks') return agents.filter((agent) => agent._kind === 'beacon');
    if (['newShell', 'fileBrowser', 'processExplorer', 'screenshot'].includes(action)) {
      return agents.filter((agent) => agent._kind === 'session');
    }
    if (action === 'registryBrowser') {
      return agents.filter(
        (agent) => agent._kind === 'session' && (agent.OS || '').toLowerCase() === 'windows',
      );
    }
    return agents;
  }

  async function reportBatchFailures(label, failures, skipped = 0) {
    if (failures.length === 0 && skipped === 0) return;
    const details = failures
      .map(({ id, error }) => `${id.slice(0, 8)}: ${errorMessage(error)}`)
      .join('\n');
    const skippedText = skipped > 0
      ? `${skipped} selected agent${skipped === 1 ? '' : 's'} did not support this action.`
      : '';
    await dialog.alert(
      [skippedText, details].filter(Boolean).join('\n\n'),
      `${label} incomplete`,
    );
  }

  function openTabsForTargets(targetIDs, tab) {
    const targets = compatibleTargets(targetIDs, tab);
    for (const agent of targets) openBackgroundTab(agent.ID, tab);
    if (targets.some((agent) => agent.ID === contextMenuTarget)) {
      interactingSession = contextMenuTarget;
      activeTabBySession = { ...activeTabBySession, [contextMenuTarget]: tab };
      eventsActive = false;
    }
    reportBatchFailures(tabMeta(tab)?.label || 'Action', [], targetIDs.length - targets.length);
  }

  async function openShellsForTargets(targetIDs) {
    const targets = compatibleTargets(targetIDs, 'newShell');
    const failures = [];
    for (const agent of targets) {
      try {
        await openShell(agent.ID, false);
      } catch (error) {
        failures.push({ id: agent.ID, error });
      }
    }
    if (targets.some((agent) => agent.ID === contextMenuTarget)) {
      interactingSession = contextMenuTarget;
    }
    await reportBatchFailures('Open shell', failures, targetIDs.length - targets.length);
  }

  async function runAutomation(ruleID, targetIDs) {
    try {
      await Promise.all(targetIDs.map((targetID) => RunAutomationRule(ruleID, targetID)));
    } catch (error) {
      await dialog.alert(`Could not run automation: ${errorMessage(error)}`);
    }
  }

  async function handleDeviceAction(action) {
    if (!contextDevice) return;
    try {
      if (action === 'copyDeviceIP') {
        await copyText(contextDevice.ip);
      } else if (action === 'copyDeviceMAC' && contextDevice.mac) {
        await copyText(contextDevice.mac);
      } else if (action === 'removeDiscoveries') {
        await removeSelectedDiscoveries();
      }
    } catch (error) {
      await dialog.alert(`Device action failed: ${errorMessage(error)}`);
    }
  }

  function handleGlobalMenuSelect(item) {
    showGlobalMenu = null;
    if (item.action) {
      executeGuiAction(item.action);
      return;
    }
    selectCommand(item, false, []);
  }

  function selectCommand(command, useSession, targetIDs = []) {
    activeCommandTargetIDs = [...targetIDs];
    if (!command.supported) {
      activeCommand = command;
      activeCommandUsesSession = useSession;
      return;
    }
    if (!command.needsInput) {
      executeSliverCommand(command.path, useSession, targetIDs);
      return;
    }
    activeCommand = command;
    activeCommandUsesSession = useSession;
  }

  async function killAgents(targetIDs) {
    if (targetIDs.length === 0 || !(await dialog.confirm(
      `Kill ${targetIDs.length} selected agent${targetIDs.length === 1 ? '' : 's'}?`,
      'Confirm Kill',
    ))) return;
    const results = await Promise.allSettled(targetIDs.map((id) => KillAgent(id)));
    const failures = [];
    results.forEach((result, index) => {
      const id = targetIDs[index];
      if (result.status === 'fulfilled') closeAgentWorkspace(id);
      else failures.push({ id, error: result.reason });
    });
    refreshAgents();
    await reportBatchFailures('Kill', failures);
  }

  async function removeBeacons(targetIDs) {
    const targets = compatibleTargets(targetIDs, 'tasks');
    if (targets.length === 0 || !(await dialog.confirm(
      `Remove ${targets.length} selected beacon record${targets.length === 1 ? '' : 's'}? This also permanently deletes all associated tasks.`,
      'Remove Beacon Record',
    ))) return;
    const results = await Promise.allSettled(targets.map((agent) => RemoveBeacon(agent.ID)));
    const failures = [];
    results.forEach((result, index) => {
      const id = targets[index].ID;
      if (result.status === 'fulfilled') closeAgentWorkspace(id);
      else failures.push({ id, error: result.reason });
    });
    refreshAgents();
    await reportBatchFailures(
      'Remove beacon',
      failures,
      targetIDs.length - targets.length,
    );
  }

  function closeAgentWorkspace(id) {
    if (interactingSession === id) interactingSession = '';
    selectedItemKeys = selectedItemKeys.filter((key) => key !== `agent:${id}`);
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

  async function renameAgents(targetIDs) {
    if (targetIDs.length === 0) return;
    const name = await dialog.prompt(
      `New name for ${targetIDs.length} selected agent${targetIDs.length === 1 ? '' : 's'}:`,
      'Rename Agent',
    );
    if (!name) return;
    const results = await Promise.allSettled(
      targetIDs.map((id) => RenameAgent(id, name)),
    );
    const failures = results.flatMap((result, index) =>
      result.status === 'rejected'
        ? [{ id: targetIDs[index], error: result.reason }]
        : []);
    refreshAgents();
    await reportBatchFailures('Rename', failures);
  }

  async function discoverNetworks(targetIDs, method, cidr = '') {
    if (targetIDs.length === 0) return;
    const results = await Promise.allSettled(
      targetIDs.map((id) => DiscoverNetwork(id, method, cidr)),
    );
    const failures = results.flatMap((result, index) =>
      result.status === 'rejected'
        ? [{ id: targetIDs[index], error: result.reason }]
        : []);
    await refreshNetworkDiscoveries();
    if (networkDiscoveries.some((device) => targetIDs.includes(device.agentID))) {
      setAgentViewMode('graph');
    } else if (failures.length === 0) {
      await dialog.alert('No devices were discovered.', 'Network Discovery');
    }
    await reportBatchFailures('Network discovery', failures);
  }

  async function promptNetworkSweep(targetIDs) {
    const cidr = await dialog.prompt(
      'IPv4 CIDR to sweep (maximum /24):',
      'Network Discovery',
      '192.168.1.0/24',
    );
    if (!cidr) return;
    await discoverNetworks(targetIDs, 'sweep', cidr);
  }

  async function clearNetworkDiscoveries(targetIDs) {
    if (targetIDs.length === 0 || !(await dialog.confirm(
      `Clear all discovered devices beneath ${targetIDs.length} selected agent${targetIDs.length === 1 ? '' : 's'}?`,
      'Clear Network Discoveries',
    ))) return;
    const results = await Promise.allSettled(
      targetIDs.map((id) => ClearNetworkDiscoveries(id)),
    );
    const failures = results.flatMap((result, index) =>
      result.status === 'rejected'
        ? [{ id: targetIDs[index], error: result.reason }]
        : []);
    await refreshNetworkDiscoveries();
    await reportBatchFailures('Clear discoveries', failures);
  }

  async function removeSelectedDiscoveries() {
    if (selectedDiscoveryKeys.length === 0) return;
    const count = selectedDiscoveryKeys.length;
    if (!(await dialog.confirm(
      `Remove ${count} selected discovered device${count === 1 ? '' : 's'}?`,
      'Remove Discovered Devices',
    ))) return;

    const byAgent = new Map();
    for (const key of selectedDiscoveryKeys) {
      const host = discoveredHosts.find((device) => discoveryKey(device) === key);
      for (const observation of host?.observations || []) {
        if (!byAgent.has(observation.agentID)) byAgent.set(observation.agentID, []);
        byAgent.get(observation.agentID).push(observation.ip);
      }
    }

    try {
      await Promise.all(
        [...byAgent.entries()].map(([agentID, ips]) =>
          RemoveNetworkDiscoveries(agentID, ips)),
      );
      selectedItemKeys = selectedItemKeys.filter((key) => !key.startsWith('device:'));
      await refreshNetworkDiscoveries();
    } catch (error) {
      await dialog.alert(`Could not remove discoveries: ${errorMessage(error)}`);
    }
  }

  function navigate(view, serverTab = '') {
    dispatch('navigate', { view, serverTab });
  }

  function openBackgroundTab(sid, tab) {
    const tabs = openTabsBySession[sid] || [];
    if (!tabs.includes(tab)) {
      openTabsBySession = { ...openTabsBySession, [sid]: [...tabs, tab] };
    }
    if (!activeTabBySession[sid]) {
      activeTabBySession = { ...activeTabBySession, [sid]: tab };
    }
  }

  function executeSliverCommand(cmd, useSession, targetIDs = activeCommandTargetIDs) {
    if (useSession && targetIDs.length > 0) {
      for (const id of targetIDs) {
        openBackgroundTab(id, 'console');
        dispatchCommand(id, cmd);
      }
      if (contextMenuTarget && targetIDs.includes(contextMenuTarget)) {
        interactingSession = contextMenuTarget;
        activeTabBySession = { ...activeTabBySession, [contextMenuTarget]: 'console' };
      }
    } else {
      dispatchCommand('', cmd);
      // User explicitly requested: "if i'm on the agents tab IT SHOULD NEVER EVER SWITCH ME TO THE SERVER TAB"
    }
    activeCommand = null;
    activeCommandTargetIDs = [];
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
      {#if selectedDiscoveryKeys.length > 0}
        <button
          type="button"
          class="remove-discoveries"
          title="Remove selected discovered devices"
          on:click={removeSelectedDiscoveries}
        >
          <i class="fas fa-trash-alt"></i>
          Remove selected ({selectedDiscoveryKeys.length})
        </button>
      {/if}
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
          discoveries={discoveredHosts}
          {selectedAgentIDs}
          {selectedDiscoveryKeys}
          filterable={false}
          on:contextmenu={handleContextMenu}
          on:select={(event) => selectAgent(event.detail)}
          on:interact={(event) => openTab(event.detail, 'console')}
          on:discoveryselect={(event) => selectDiscovery(event.detail)}
          on:discoverycontextmenu={(event) =>
            openDeviceContextMenu(event.detail.event, {
              ...event.detail.device,
              key: discoveryKey(event.detail.device),
            })}
        />
      {:else}
        <NetworkGraph
          embedded
          sessions={graphData.filter((agent) => agent._kind === 'session')}
          beacons={graphData.filter((agent) => agent._kind === 'beacon')}
          {pivotGraph}
          {pivotListeners}
          discoveries={discoveredHosts}
          {selectedAgentIDs}
          {selectedDiscoveryKeys}
          onSelect={(id) => interactingSession = id}
          onInteract={(id) => openTab(id, 'console')}
          onContextMenu={openAgentContextMenu}
          onSelectionChange={syncGraphSelection}
          onDeviceContextMenu={openDeviceContextMenu}
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
    sessionID={activeCommandUsesSession
      ? (activeCommandTargetIDs.length > 1
        ? `${activeCommandTargetIDs.length} selected agents`
        : activeCommandTargetIDs[0] || interactingSession)
      : ''}
    command={activeCommand}
    on:close={() => {
      activeCommand = null;
      activeCommandTargetIDs = [];
    }}
    on:execute={(event) =>
      executeSliverCommand(event.detail.cmd, activeCommandUsesSession, activeCommandTargetIDs)}
  />
{/if}
{#if showContextMenu}
  <ContextMenu
    x={contextMenuX}
    y={contextMenuY}
    coreActions={contextMenuKind === 'device' ? contextDeviceActions : contextCoreActions}
    footerActions={contextMenuKind === 'device'
      ? []
      : [{ id: 'rename', icon: 'fas fa-pen', label: 'Rename' }]}
    dangerActions={contextMenuKind === 'device' ? contextDeviceDangerActions : contextDangerActions}
    categories={contextMenuKind === 'device' ? [] : contextCategories}
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

  .remove-discoveries {
    padding: 4px 8px;
    border: 1px solid var(--danger-color);
    border-radius: 3px;
    background: rgba(255, 74, 74, 0.1);
    color: var(--danger-color);
    cursor: pointer;
  }

  .remove-discoveries:hover {
    background: rgba(255, 74, 74, 0.2);
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
