<script>
  import { onMount } from 'svelte';

  import { listBeacons, listSessions } from './lib/api/agents.js';
  import { Connect, GetVersion } from './lib/api/connection.js';
  import { onSliverEvent } from './lib/api/runtime.js';
  import PrimaryNavigation from './lib/components/layout/PrimaryNavigation.svelte';
  import StatusBar from './lib/components/layout/StatusBar.svelte';
  import TitleBar from './lib/components/layout/TitleBar.svelte';
  import GlobalDialog from './lib/components/ui/GlobalDialog.svelte';
  import Toasts from './lib/components/ui/Toasts.svelte';
  import AgentWorkspace from './lib/features/agents/AgentWorkspace.svelte';
  import AutomationView from './lib/features/automation/AutomationView.svelte';
  import ConnectionScreen from './lib/features/connection/ConnectionScreen.svelte';
  import ServerView from './lib/features/server/ServerView.svelte';
  import SettingsView from './lib/features/settings/SettingsView.svelte';

  let connected = false;
  let activeView = 'agents';
  let serverTab = 'listeners';
  let activeProfile = '';
  let serverVersion = '';
  let sessions = [];
  let beacons = [];
  let eventHistory = [];
  let reconnectInterval = null;
  let refreshTimer = null;
  let agentRefreshInterval = null;
  let agentRefreshInFlight = false;

  // Coalesce event bursts. A slower periodic refresh below also reconciles
  // ordinary check-ins, which do not emit lifecycle events.
  function scheduleAgentRefresh() {
    if (refreshTimer) return;
    refreshTimer = setTimeout(() => {
      refreshTimer = null;
      refreshAgents();
    }, 250);
  }

  function onConnected(profile) {
    connected = true;
    activeProfile = profile;
    if (reconnectInterval) clearInterval(reconnectInterval);
    reconnectInterval = null;

    fetchVersion();
    refreshAgents();
    if (agentRefreshInterval) clearInterval(agentRefreshInterval);
    agentRefreshInterval = setInterval(refreshAgents, 5000);
  }

  function handleConnectionFailure() {
    connected = false;
    serverVersion = '';
    if (agentRefreshInterval) clearInterval(agentRefreshInterval);
    agentRefreshInterval = null;
    if (!reconnectInterval && activeProfile) {
      reconnectInterval = setInterval(async () => {
        try {
          await Connect(activeProfile);
          onConnected(activeProfile);
        } catch {
          // Keep retrying until the teamserver is available.
        }
      }, 5000);
    }
  }

  async function fetchBeacons() {
    if (!connected) return;
    try {
      beacons = await listBeacons();
    } catch (error) {
      // Disconnects are detected via the event stream closing, not here.
      console.error(error);
    }
  }

  async function fetchSessions() {
    if (!connected) return;
    try {
      sessions = await listSessions();
    } catch (error) {
      console.error(error);
    }
  }

  async function refreshAgents() {
    if (!connected || agentRefreshInFlight) return;
    agentRefreshInFlight = true;
    try {
      await Promise.all([fetchBeacons(), fetchSessions()]);
    } finally {
      agentRefreshInFlight = false;
    }
  }

  async function fetchVersion() {
    try {
      const version = await GetVersion();
      serverVersion = `${version.Major}.${version.Minor}.${version.Patch}`;
    } catch {
      // Version display is non-critical.
    }
  }

  function handleWorkspaceNavigation(event) {
    activeView = event.detail.view;
    if (event.detail.serverTab) serverTab = event.detail.serverTab;
  }

  onMount(() => {
    try {
      const savedTheme = localStorage.getItem('sliver-theme');
      if (savedTheme) document.documentElement.setAttribute('data-theme', savedTheme);
    } catch {
      // localStorage can be unavailable in hardened webviews.
    }

    const stopEvents = onSliverEvent((event) => {
      const type = event.type || '';
      if (type === 'stream-closed') {
        if (connected) handleConnectionFailure();
        return;
      }
      eventHistory = [...eventHistory.slice(-1999), { ...event, time: Date.now() }];
      // Any session/beacon lifecycle change reconciles the agent list.
      if (type.startsWith('session-') || type.startsWith('beacon-')) {
        scheduleAgentRefresh();
      }
    });

    return () => {
      if (refreshTimer) clearTimeout(refreshTimer);
      if (reconnectInterval) clearInterval(reconnectInterval);
      if (agentRefreshInterval) clearInterval(agentRefreshInterval);
      stopEvents();
    };
  });
</script>

<div class="app-container">
  <TitleBar />

  <main class="app-content">
    {#if !connected}
      {#if activeProfile}
        <div class="reconnecting">
          Connection lost. Reconnecting to {activeProfile}...
        </div>
      {:else}
        <ConnectionScreen on:connected={(event) => onConnected(event.detail)} />
      {/if}
    {:else}
      <PrimaryNavigation active={activeView} onSelect={(view) => activeView = view} />

      <!-- Keep the workspace mounted (just hidden) on other views so open agent
           tabs and their state survive navigating to Server/Settings and back. -->
      <div class="view-fill" class:hidden={activeView !== 'agents'}>
        <AgentWorkspace
          {sessions}
          {beacons}
          {eventHistory}
          {refreshAgents}
          on:navigate={handleWorkspaceNavigation}
        />
      </div>
      {#if activeView === 'server'}
        <div class="view-fill">
          <ServerView eventHistory={eventHistory} bind:active={serverTab} />
        </div>
      {:else if activeView === 'automation'}
        <div class="view-fill">
          <AutomationView {sessions} {beacons} />
        </div>
      {:else if activeView === 'settings'}
        <div class="view-fill">
          <SettingsView />
        </div>
      {/if}

      <StatusBar
        {serverVersion}
        sessionCount={sessions.length}
        beaconCount={beacons.length}
      />
      <Toasts />
      <GlobalDialog />
    {/if}
  </main>
</div>

<style>
  .app-container {
    display: flex;
    flex-direction: column;
    width: 100vw;
    height: 100vh;
    overflow: hidden;
  }

  .app-content {
    position: relative;
    display: flex;
    flex: 1;
    flex-direction: column;
    overflow: hidden;
  }

  .reconnecting {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
    color: var(--text-color);
    font-size: 1.2em;
  }

  .view-fill.hidden {
    display: none;
  }
</style>
