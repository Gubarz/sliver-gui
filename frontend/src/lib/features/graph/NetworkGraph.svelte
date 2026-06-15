<script>
  import { SvelteFlow, Background, Controls, MiniMap, Position } from '@xyflow/svelte';
  import { onDestroy } from 'svelte';
  import '@xyflow/svelte/dist/style.css';
  import dagre from '@dagrejs/dagre';
  import GraphNode from './GraphNode.svelte';
  import PanelShell from '../../components/ui/PanelShell.svelte';
  import {
    agentRemoteAddress,
    isAgentOnline,
    pivotParentMap,
    shortAgentID,
  } from '../../utils/agents.js';

  let {
    sessions = [],
    beacons = [],
    pivotGraph = null,
    pivotListeners = [],
    discoveries = [],
    selectedAgentIDs = [],
    selectedDiscoveryKeys = [],
    embedded = false,
    onClose = () => {},
    onSelect = () => {},
    onInteract = () => {},
    onContextMenu = () => {},
    onSelectionChange = () => {},
    onDeviceContextMenu = () => {},
  } = $props();

  const nodeTypes = { box: GraphNode };

  let nodes = $state.raw([]);
  let edges = $state.raw([]);
  let lastSig = '';
  let now = $state(Math.floor(Date.now() / 1000));
  const statusTicker = setInterval(() => {
    now = Math.floor(Date.now() / 1000);
  }, 5000);
  onDestroy(() => clearInterval(statusTicker));

  const SERVER_W = 180, SERVER_H = 44;
  const LISTENER_W = 220, LISTENER_H = 40;
  const NODE_W = 210, NODE_H = 88;
  const DEVICE_W = 210, DEVICE_H = 88;

  function c2Details(value) {
    const raw = String(value || 'unknown');
    const [scheme, remainder = raw] = raw.includes('://') ? raw.split('://', 2) : ['tcp', raw];
    const chain = remainder.split('->').map((part) => part.trim()).filter(Boolean);
    const endpoint = (chain[0] || 'unknown').split('/')[0];
    return {
      key: `${scheme}_${endpoint}`,
      // Ports on connection nodes identify the listener or pivot endpoint.
      // Only agent cards hide their ephemeral outbound source ports.
      label: `${scheme.toUpperCase()} ${endpoint}`,
      endpoint,
      isPivot: chain.length > 1,
      parentName: chain.length > 1 ? chain[chain.length - 1] : '',
    };
  }

  function pivotParentFromC2(c2, agentsByAddress, agentsByName) {
    if (!c2.isPivot) return '';
    const addressMatch = agentsByAddress.get(c2.endpoint.toLowerCase());
    if (addressMatch) return addressMatch;
    return agentsByName.get(c2.parentName.toLowerCase()) || '';
  }

  function endpointHost(endpoint) {
    const value = String(endpoint || '');
    const ipv6 = value.match(/^\[(.+)\](?::\d+)?$/);
    if (ipv6) return `[${ipv6[1]}]`;
    return value.replace(/:\d+$/, '');
  }

  function endpointPort(endpoint) {
    const match = String(endpoint || '').match(/:(\d+)$/);
    return match ? match[1] : '';
  }

  function pivotListenerFor(parentID, c2) {
    const remote = c2.endpoint.toLowerCase();
    const parentListeners = (pivotListeners || []).filter((listener) =>
      listener.ParentSessionID === parentID);
    return parentListeners.find((listener) =>
      (listener.Pivots || []).some((pivot) =>
        String(pivot.RemoteAddress || '').toLowerCase() === remote)) ||
      (parentListeners.length === 1 ? parentListeners[0] : null);
  }

  function pivotDetails(parentID, c2) {
    const listener = pivotListenerFor(parentID, c2);
    if (!listener?.BindAddress) {
      return {
        id: `p_${parentID}_${c2.key}`,
        label: c2.label,
      };
    }
    const port = endpointPort(listener.BindAddress);
    const endpoint = port ? `${endpointHost(c2.endpoint)}:${port}` : listener.BindAddress;
    return {
      id: `p_${parentID}_${listener.ID}`,
      label: `${listener.Type || 'TCP'} ${endpoint}`,
    };
  }

  // Lay the nodes out top-to-bottom with dagre (Svelte Flow has no auto-layout).
  function layout(rawNodes, rawEdges) {
    const g = new dagre.graphlib.Graph();
    g.setGraph({ rankdir: 'TB', nodesep: 45, ranksep: 90, marginx: 20, marginy: 20 });
    g.setDefaultEdgeLabel(() => ({}));
    rawNodes.forEach((n) => g.setNode(n.id, { width: n.w, height: n.h }));
    rawEdges.forEach((e) => g.setEdge(e.source, e.target));
    dagre.layout(g);
    return rawNodes.map((n) => {
      const p = g.node(n.id);
      return {
        id: n.id,
        type: 'box',
        data: n.data,
        position: { x: p.x - n.w / 2, y: p.y - n.h / 2 },
        targetPosition: Position.Top,
        sourcePosition: Position.Bottom,
        draggable: true,
        selected: n.selected ?? false,
      };
    });
  }

  function preservePositions(nextNodes, nextEdges) {
    const positions = new Map(nodes.map((node) => [node.id, node.position]));
    const layoutPositions = new Map(nextNodes.map((node) => [node.id, node.position]));
    const parentByNode = new Map(nextEdges.map((edge) => [edge.target, edge.source]));
    return nextNodes.map((node) => {
      const previous = nodes.find((current) => current.id === node.id);
      const position = positions.get(node.id);
      if (position) return {
        ...node,
        position,
        selected: previous?.selected ?? node.selected ?? false,
      };

      const parentID = parentByNode.get(node.id);
      const oldParent = positions.get(parentID);
      const layoutParent = layoutPositions.get(parentID);
      if (!oldParent || !layoutParent) return node;
      return {
        ...node,
        position: {
          x: node.position.x + oldParent.x - layoutParent.x,
          y: node.position.y + oldParent.y - layoutParent.y,
        },
      };
    });
  }

  function topologySignature() {
    const pivotRelations = [...pivotParentMap(pivotGraph).entries()]
      .map(([child, parent]) => `${parent}>${child}`)
      .sort();
    return [
      ...(sessions || []).map((session) =>
        `${session.ID}:${session.ActiveC2}:${session.RemoteAddress}:${session.Name}:${isAgentOnline(session, now)}`),
      ...(beacons || []).map((beacon) =>
        `${beacon.ID}:${beacon.ActiveC2}:${beacon.RemoteAddress}:${beacon.Name}:${isAgentOnline(beacon, now)}`),
      ...pivotRelations,
      ...(pivotListeners || []).map((listener) =>
        `${listener.ParentSessionID}:${listener.ID}:${listener.Type}:${listener.BindAddress}`),
      ...(discoveries || []).map((device) =>
        `${device.key}:${(device.observerIDs || [device.agentID]).join(',')}:${device.ip}:${device.mac}:${device.hostname}:${device.vendor}:${device.osHint}:${device.ttl}:${device.method}:${device.lastSeen}`),
    ].sort().join('|');
  }

  function build() {
    const rawNodes = [];
    const rawEdges = [];

    rawNodes.push({
      id: 'ts',
      w: SERVER_W,
      h: SERVER_H,
      data: { variant: 'server', label: 'Sliver Teamserver' },
    });

    const allAgents = [
      ...(sessions || []).map((agent) => ({ ...agent, _kind: 'session' })),
      ...(beacons || []).map((agent) => ({ ...agent, _kind: 'beacon' })),
    ];
    const allAgentIds = new Set(allAgents.map((agent) => agent.ID));
    const agentsByName = new Map();
    const agentsByAddress = new Map();
    for (const agent of allAgents) {
      if (agent.RemoteAddress) {
        agentsByAddress.set(String(agent.RemoteAddress).toLowerCase(), agent.ID);
      }
      for (const name of [agent.Name, agent.Hostname]) {
        if (name && !agentsByName.has(name.toLowerCase())) {
          agentsByName.set(name.toLowerCase(), agent.ID);
        }
      }
    }
    const parentBySession = pivotParentMap(pivotGraph);

    // Build every agent node before relationships so an edge can never point
    // at a parent that was skipped or replaced by a listener placeholder.
    for (const impl of allAgents) {
      const kind = impl._kind;
      const os = (impl.OS || '').toLowerCase();
      const icon = os.includes('win') ? 'fab fa-windows'
        : (os.includes('darwin') || os.includes('mac')) ? 'fab fa-apple'
        : os.includes('linux') ? 'fab fa-linux' : 'fas fa-microchip';

      const online = isAgentOnline(impl, now);
      rawNodes.push({
        id: impl.ID, w: NODE_W, h: NODE_H,
        data: {
          variant: 'agent', kind, icon,
          agentID: shortAgentID(impl.ID),
          implantName: impl.Name || '-',
          user: impl.Username || '?', host: impl.Hostname || '?',
          addr: agentRemoteAddress(impl, parentBySession, allAgents),
          dead: !online,
        },
        selected: selectedAgentIDs.includes(impl.ID),
      });
    }

    const seenListeners = new Set();
    const seenPivotListeners = new Set();
    for (const impl of allAgents) {
      const kind = impl._kind;
      const c2 = c2Details(impl.ActiveC2 || impl.RemoteAddress);
      const candidateParent = parentBySession.get(impl.ID) ||
        pivotParentFromC2(c2, agentsByAddress, agentsByName);
      const parentID = candidateParent !== impl.ID && allAgentIds.has(candidateParent)
        ? candidateParent
        : '';
      let sourceId = `l_${c2.key}`;

      if (parentID) {
        const pivot = pivotDetails(parentID, c2);
        sourceId = pivot.id;
        if (!seenPivotListeners.has(sourceId)) {
          seenPivotListeners.add(sourceId);
          rawNodes.push({
            id: sourceId,
            w: LISTENER_W,
            h: LISTENER_H,
            data: { variant: 'listener', label: pivot.label },
          });
          rawEdges.push({
            id: `e_${parentID}_${sourceId}`,
            source: parentID,
            target: sourceId,
            style: 'stroke:var(--success-color);stroke-width:3',
            animated: false,
          });
        }
      } else if (!seenListeners.has(sourceId)) {
        seenListeners.add(sourceId);
        rawNodes.push({
          id: sourceId,
          w: LISTENER_W,
          h: LISTENER_H,
          data: { variant: 'listener', label: c2.label },
        });
        rawEdges.push({
          id: `e_ts_${sourceId}`,
          source: 'ts',
          target: sourceId,
          style: 'stroke:var(--border-color);stroke-dasharray:4',
          animated: false,
        });
      }

      rawEdges.push({
        id: `e_${sourceId}_${impl.ID}`, source: sourceId, target: impl.ID,
        animated: kind === 'beacon',
        style: `stroke:${kind === 'beacon' ? '#d6a23e' : 'var(--success-color)'};stroke-width:${parentID ? 3 : 2}`,
      });
    }

    for (const device of discoveries || []) {
      const observerIDs = (device.observerIDs || [device.agentID])
        .filter((agentID) => allAgentIds.has(agentID));
      if (observerIDs.length === 0) continue;
      const key = device.key || `${device.agentID}|${device.ip}`;
      const deviceID = `d_${key}`;
      rawNodes.push({
        id: deviceID,
        w: DEVICE_W,
        h: DEVICE_H,
        data: {
          variant: 'device',
          ip: device.ip,
          mac: device.mac || '',
          hostname: device.hostname || '',
          vendor: device.vendor || '',
          osHint: device.osHint || '',
          ttl: device.ttl || 0,
          method: device.method || 'discovery',
          lastSeen: device.lastSeen || 0,
          agentID: observerIDs[0],
          observerIDs,
          key,
        },
        selected: selectedDiscoveryKeys.includes(key),
      });
      for (const observerID of observerIDs) {
        rawEdges.push({
          id: `e_${observerID}_${deviceID}`,
          source: observerID,
          target: deviceID,
          style: 'stroke:#58a6ff;stroke-width:1.5;stroke-dasharray:5',
          animated: false,
        });
      }
    }

    nodes = preservePositions(layout(rawNodes, rawEdges), rawEdges);
    edges = rawEdges;
  }

  // Refresh labels and relationships without treating volatile pivot/session
  // fields as topology changes. Existing node coordinates survive rebuilds.
  $effect(() => {
    const sig = topologySignature();
    if (sig === lastSig) return;
    lastSig = sig;
    build();
  });

  function handleSelectionChange(selection) {
    onSelectionChange({
      agentIDs: selection.nodes
        .filter((node) => node.data?.variant === 'agent')
        .map((node) => node.id),
      deviceKeys: selection.nodes
        .filter((node) => node.data?.variant === 'device')
        .map((node) => node.data.key),
    });
  }

  function handleNodeClick(evt) {
    const node = evt?.node || evt?.detail?.node;
    const nativeEvent = evt?.event || evt?.detail?.event;
    if (node?.data?.variant === 'device') return;
    if (node && node.data && node.data.variant === 'agent') {
      if (!embedded || nativeEvent?.detail === 2) onInteract(node.id);
      else onSelect(node.id);
    }
  }

  function handleNodeContextMenu(evt) {
    // SvelteFlow events sometimes wrap native events in detail.event, or pass native event as first arg and node as second.
    // In @xyflow/svelte, typically it's an object with `event` and `node` properties, or a CustomEvent with detail.
    let nativeEvent = evt?.event || evt?.detail?.event || evt;
    if (nativeEvent && typeof nativeEvent.preventDefault === 'function') {
      nativeEvent.preventDefault();
    }
    
    const node = evt?.node || evt?.detail?.node;
    if (node?.data?.variant === 'device') {
      onDeviceContextMenu(nativeEvent, node.data);
      return;
    }
    if (node && node.data && node.data.variant === 'agent') {
      const source = node.data.kind === 'beacon' ? beacons : sessions;
      const agent = source.find((item) => item.ID === node.id);
      if (agent) onContextMenu(nativeEvent, { ...agent, _kind: node.data.kind });
    }
  }

</script>
<PanelShell
  title="Network Topology"
  icon="fa-project-diagram"
  width="90vw"
  height={embedded ? '100%' : 'auto'}
  bodyPadding="0"
  {embedded}
  showHeader={!embedded}
  on:close={onClose}
>
  <svelte:fragment slot="actions">
    <div class="legend">
      <span><span class="sw session"></span> session</span>
      <span><span class="sw beacon"></span> beacon</span>
      <span><span class="sw device"></span> discovered</span>
      <span class="hint">click an agent to interact</span>
    </div>
  </svelte:fragment>

  <div class="flow-container" class:embedded>
    <SvelteFlow
      colorMode="dark"
      bind:nodes
        bind:edges
        {nodeTypes}
        fitView
        minZoom={0.2}
        onnodeclick={handleNodeClick}
        onnodecontextmenu={handleNodeContextMenu}
        onselectionchange={handleSelectionChange}
        multiSelectionKey={['Control', 'Meta']}
        proOptions={{ hideAttribution: true }}
      >
        <Background gap={18} />
        <Controls />
        <MiniMap pannable zoomable />
      </SvelteFlow>
    </div>
</PanelShell>

<style>
  .flow-container {
    height: 75vh;
    width: 100%;
    position: relative;
    background-color: var(--panel-body-background);
  }
  .flow-container.embedded {
    height: 100%;
  }
  
  .legend {
    display: flex;
    gap: 15px;
    align-items: center;
    font-size: 0.9em;
    margin-right: 15px;
  }
  .sw {
    display: inline-block;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    margin-right: 5px;
    vertical-align: middle;
  }
  .legend .sw.session { background: var(--success-color); }
  .legend .sw.beacon { background: #d6a23e; }
  .legend .sw.device { background: #58a6ff; }
  .legend .hint { font-style: italic; }
</style>
